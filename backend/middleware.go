package main

import (
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// CORSMiddleware returns a middleware that sets CORS headers based on the
// configured allowed origins. Use "*" in the origins list to allow any origin.
func CORSMiddleware(allowedOrigins []string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		for _, allowedOrigin := range allowedOrigins {
			if allowedOrigin == "*" || origin == allowedOrigin {
				val := origin
				if allowedOrigin == "*" {
					val = "*"
				}
				w.Header().Set("Access-Control-Allow-Origin", val)
				break
			}
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	}
}

// AuthenticateJWT validates the Authorization: Bearer <token> header using the
// provided secret. On success, it forwards the request to the next handler;
// otherwise it returns HTTP 403.
func AuthenticateJWT(secret []byte, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Headers: %v", r.Header.Get("authorization"))

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			log.Printf("Missing Authorization Header")
			return
		}

		if !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			http.Error(w, "Forbidden", http.StatusForbidden)
			log.Printf("Invalid Authorization scheme")
			return
		}

		tokenString := strings.TrimSpace(authHeader[len("Bearer "):])

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return secret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Forbidden", http.StatusForbidden)
			log.Printf("Token is NOT valid")
			return
		}

		next.ServeHTTP(w, r)
	}
}

// IPWhitelist restricts access to the provided handler to clients whose IP is
// included in the allowedIPs map. When the map is empty, all requests are blocked.
func IPWhitelist(allowedIPs map[string]bool, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientIP := r.RemoteAddr
		if host, _, err := net.SplitHostPort(clientIP); err == nil {
			clientIP = host
		}
		if !allowedIPs[clientIP] {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	}
}
