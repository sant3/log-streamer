package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func dummyHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}

// --- CORS Middleware ---

func TestCORSMiddleware_AllowedOrigin(t *testing.T) {
	origins := []string{"http://localhost:3000", "https://example.com"}
	handler := CORSMiddleware(origins, dummyHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Errorf("expected CORS origin http://localhost:3000, got %q", rr.Header().Get("Access-Control-Allow-Origin"))
	}
	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

func TestCORSMiddleware_DisallowedOrigin(t *testing.T) {
	origins := []string{"http://localhost:3000"}
	handler := CORSMiddleware(origins, dummyHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://evil.com")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Errorf("expected no CORS origin header, got %q", rr.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORSMiddleware_Wildcard(t *testing.T) {
	origins := []string{"*"}
	handler := CORSMiddleware(origins, dummyHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://any-origin.com")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected CORS origin *, got %q", rr.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORSMiddleware_OptionsPreflightReturns200(t *testing.T) {
	origins := []string{"http://localhost:3000"}
	handler := CORSMiddleware(origins, dummyHandler)

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200 for OPTIONS, got %d", rr.Code)
	}
	// Body should be empty for OPTIONS (handler not called)
	if rr.Body.String() == "OK" {
		t.Error("expected handler NOT to be called for OPTIONS request")
	}
}

func TestCORSMiddleware_Headers(t *testing.T) {
	origins := []string{"http://localhost:3000"}
	handler := CORSMiddleware(origins, dummyHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Header().Get("Access-Control-Allow-Methods") != "GET, POST, OPTIONS" {
		t.Errorf("unexpected Allow-Methods header: %q", rr.Header().Get("Access-Control-Allow-Methods"))
	}
	if rr.Header().Get("Access-Control-Allow-Headers") != "Content-Type, Authorization" {
		t.Errorf("unexpected Allow-Headers header: %q", rr.Header().Get("Access-Control-Allow-Headers"))
	}
}

// --- JWT Authentication ---

func createTestJWT(secret []byte) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "test-user",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString(secret)
	return tokenString
}

func TestAuthenticateJWT_ValidToken(t *testing.T) {
	secret := []byte("my-test-secret")
	handler := AuthenticateJWT(secret, dummyHandler)

	token := createTestJWT(secret)
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200 for valid token, got %d", rr.Code)
	}
}

func TestAuthenticateJWT_MissingHeader(t *testing.T) {
	secret := []byte("my-test-secret")
	handler := AuthenticateJWT(secret, dummyHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status 403 for missing header, got %d", rr.Code)
	}
}

func TestAuthenticateJWT_InvalidScheme(t *testing.T) {
	secret := []byte("my-test-secret")
	handler := AuthenticateJWT(secret, dummyHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status 403 for invalid scheme, got %d", rr.Code)
	}
}

func TestAuthenticateJWT_InvalidToken(t *testing.T) {
	secret := []byte("my-test-secret")
	handler := AuthenticateJWT(secret, dummyHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status 403 for invalid token, got %d", rr.Code)
	}
}

func TestAuthenticateJWT_WrongSecret(t *testing.T) {
	handler := AuthenticateJWT([]byte("correct-secret"), dummyHandler)

	token := createTestJWT([]byte("wrong-secret"))
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status 403 for wrong secret, got %d", rr.Code)
	}
}

func TestAuthenticateJWT_ExpiredToken(t *testing.T) {
	secret := []byte("my-test-secret")
	handler := AuthenticateJWT(secret, dummyHandler)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "test-user",
		"exp": time.Now().Add(-time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString(secret)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status 403 for expired token, got %d", rr.Code)
	}
}

// --- IP Whitelist ---

func TestIPWhitelist_AllowedIP(t *testing.T) {
	allowedIPs := map[string]bool{"192.168.1.100": true}
	handler := IPWhitelist(allowedIPs, dummyHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200 for allowed IP, got %d", rr.Code)
	}
}

func TestIPWhitelist_BlockedIP(t *testing.T) {
	allowedIPs := map[string]bool{"192.168.1.100": true}
	handler := IPWhitelist(allowedIPs, dummyHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status 403 for blocked IP, got %d", rr.Code)
	}
}

func TestIPWhitelist_NoPort(t *testing.T) {
	allowedIPs := map[string]bool{"192.168.1.100": true}
	handler := IPWhitelist(allowedIPs, dummyHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.100"
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200 for allowed IP without port, got %d", rr.Code)
	}
}
