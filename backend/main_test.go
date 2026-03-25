package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// testConfig returns a Config suitable for testing.
func testConfig() *Config {
	return &Config{
		Port:        5005,
		HTTPSPort:   8443,
		EnableHTTPS: false,
		CertPath:    "",
		KeyPath:     "",
		LogsDir:     ".",
		CORSOrigins: []string{"http://localhost:3000"},
		AllowedIPs:  map[string]bool{},
		JWTSecret:   []byte("test-secret"),
	}
}

func TestAliveHandler(t *testing.T) {
	cfg := testConfig()
	router := setupRouter(cfg)
	req, _ := http.NewRequest("GET", "/alive", nil)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := "OK"
	if strings.TrimSpace(rr.Body.String()) != expected {
		t.Errorf("handler returned unexpected body: got %q want %q", rr.Body.String(), expected)
	}
}

func TestVersionHandler(t *testing.T) {
	cfg := testConfig()
	router := setupRouter(cfg)

	originalVersion := version
	originalBuildDate := buildDate
	version = "test-1.0"
	buildDate = "2025-01-01"
	defer func() {
		version = originalVersion
		buildDate = originalBuildDate
	}()

	req, _ := http.NewRequest("GET", "/version", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("could not deserialize response: %v", err)
	}

	if resp["version"] != version || resp["buildDate"] != buildDate {
		t.Errorf("handler returned unexpected body: got %v want {\"version\":\"%s\",\"buildDate\":\"%s\"}", rr.Body.String(), version, buildDate)
	}
}

func TestListFilesHandler(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-logs-")
	if err != nil {
		t.Fatalf("could not create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := testConfig()
	cfg.LogsDir = tempDir

	expectedFiles := []string{"test1.log", "test2.log"}
	for _, f := range expectedFiles {
		if err := os.WriteFile(filepath.Join(tempDir, f), []byte("log data"), 0644); err != nil {
			t.Fatalf("could not create temp file: %v", err)
		}
	}
	if err := os.WriteFile(filepath.Join(tempDir, "not-a-log.txt"), []byte("other data"), 0644); err != nil {
		t.Fatalf("could not create temp file: %v", err)
	}

	router := setupRouter(cfg)
	req, _ := http.NewRequest("GET", "/list-files", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var actualFiles []string
	err = json.Unmarshal(rr.Body.Bytes(), &actualFiles)
	if err != nil {
		t.Fatalf("could not deserialize response: %v", err)
	}

	expectedSet := make(map[string]bool)
	for _, f := range expectedFiles {
		expectedSet[f] = true
	}
	actualSet := make(map[string]bool)
	for _, f := range actualFiles {
		actualSet[f] = true
	}

	if !reflect.DeepEqual(expectedSet, actualSet) {
		t.Errorf("handler returned unexpected file list: got %v want %v", actualFiles, expectedFiles)
	}
}

func TestStreamLogsHandler_FileNotFound(t *testing.T) {
	cfg := testConfig()
	router := setupRouter(cfg)
	req, _ := http.NewRequest("GET", "/stream-logs?file=nonexistent.log", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code for non-existent file: got %v want %v", status, http.StatusInternalServerError)
	}

	expectedError := "Error: log file not found or is not a regular file\n"
	if rr.Body.String() != expectedError {
		t.Errorf("handler returned unexpected body: got %q want %q", rr.Body.String(), expectedError)
	}
}

func TestSetupRouter_AllRoutes(t *testing.T) {
	cfg := testConfig()
	router := setupRouter(cfg)

	routes := []struct {
		method string
		path   string
		expect int
	}{
		{"GET", "/alive", http.StatusOK},
		{"GET", "/version", http.StatusOK},
	}

	for _, rt := range routes {
		req := httptest.NewRequest(rt.method, rt.path, nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		if rr.Code != rt.expect {
			t.Errorf("%s %s: expected status %d, got %d", rt.method, rt.path, rt.expect, rr.Code)
		}
	}
}

func TestSetupRouter_CORSOnEndpoints(t *testing.T) {
	cfg := testConfig()
	cfg.CORSOrigins = []string{"http://myapp.com"}
	router := setupRouter(cfg)

	endpoints := []string{"/alive", "/version", "/list-files"}
	for _, ep := range endpoints {
		req := httptest.NewRequest("OPTIONS", ep, nil)
		req.Header.Set("Origin", "http://myapp.com")
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("OPTIONS %s: expected 200, got %d", ep, rr.Code)
		}
		if rr.Header().Get("Access-Control-Allow-Origin") != "http://myapp.com" {
			t.Errorf("OPTIONS %s: expected CORS origin header, got %q", ep, rr.Header().Get("Access-Control-Allow-Origin"))
		}
	}
}

