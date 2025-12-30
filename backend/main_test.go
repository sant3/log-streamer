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

// Note: We use the real setupRouter from main.go to ensure tests run against the actual configuration.

func TestAliveHandler(t *testing.T) {
	router := setupRouter()
	req, _ := http.NewRequest("GET", "/alive", nil)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// The handler now returns "OK" without a newline
	expected := "OK"
	if strings.TrimSpace(rr.Body.String()) != expected {
		t.Errorf("handler returned unexpected body: got %q want %q", rr.Body.String(), expected)
	}
}

func TestVersionHandler(t *testing.T) {
	router := setupRouter()
	// Set dummy version info for the test
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

	originalLogsDir := logsDir
	logsDir = tempDir
	defer func() { logsDir = originalLogsDir }()

	expectedFiles := []string{"test1.log", "test2.log"}
	for _, f := range expectedFiles {
		if err := os.WriteFile(filepath.Join(tempDir, f), []byte("log data"), 0644); err != nil {
			t.Fatalf("could not create temp file: %v", err)
		}
	}
	if err := os.WriteFile(filepath.Join(tempDir, "not-a-log.txt"), []byte("other data"), 0644); err != nil {
		t.Fatalf("could not create temp file: %v", err)
	}

	router := setupRouter()
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
	router := setupRouter()
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
