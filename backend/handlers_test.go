package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestStreamLogsHandler_MissingFileParam(t *testing.T) {
	handler := StreamLogsHandler(".")

	req := httptest.NewRequest("GET", "/stream-logs", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "'file' query parameter is required") {
		t.Errorf("unexpected body: %q", rr.Body.String())
	}
}

func TestStreamLogsHandler_DirectoryTraversal(t *testing.T) {
	handler := StreamLogsHandler(".")

	req := httptest.NewRequest("GET", "/stream-logs?file=../etc/passwd.log", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Invalid file name") {
		t.Errorf("unexpected body: %q", rr.Body.String())
	}
}

func TestStreamLogsHandler_NonLogExtension(t *testing.T) {
	handler := StreamLogsHandler(".")

	req := httptest.NewRequest("GET", "/stream-logs?file=data.txt", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Only .log files are allowed") {
		t.Errorf("unexpected body: %q", rr.Body.String())
	}
}

func TestStreamLogsHandler_FileIsDirectory(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-logs-")
	if err != nil {
		t.Fatalf("could not create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a directory named "fake.log"
	dirPath := filepath.Join(tempDir, "fake.log")
	if err := os.Mkdir(dirPath, 0755); err != nil {
		t.Fatalf("could not create dir: %v", err)
	}

	handler := StreamLogsHandler(tempDir)

	req := httptest.NewRequest("GET", "/stream-logs?file=fake.log", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500 for directory, got %d", rr.Code)
	}
}

func TestListFilesHandler_EmptyDirectory(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-empty-logs-")
	if err != nil {
		t.Fatalf("could not create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	handler := ListFilesHandler(tempDir)

	req := httptest.NewRequest("GET", "/list-files", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
	// null is valid JSON for nil slice
	body := strings.TrimSpace(rr.Body.String())
	if body != "null" {
		t.Errorf("expected null for empty dir, got %q", body)
	}
}

func TestListFilesHandler_InvalidDirectory(t *testing.T) {
	handler := ListFilesHandler("/nonexistent/path/that/does/not/exist")

	req := httptest.NewRequest("GET", "/list-files", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rr.Code)
	}
}

func TestAliveHandler_CacheControl(t *testing.T) {
	req := httptest.NewRequest("GET", "/alive", nil)
	rr := httptest.NewRecorder()

	AliveHandler(rr, req)

	if rr.Header().Get("Cache-Control") != "no-cache" {
		t.Errorf("expected Cache-Control: no-cache, got %q", rr.Header().Get("Cache-Control"))
	}
}

func TestVersionHandler_ContentType(t *testing.T) {
	req := httptest.NewRequest("GET", "/version", nil)
	rr := httptest.NewRecorder()

	VersionHandler(rr, req)

	if rr.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type: application/json, got %q", rr.Header().Get("Content-Type"))
	}
}

func TestStreamLogsHandler_SubdirectoryTraversal(t *testing.T) {
	handler := StreamLogsHandler(".")

	tests := []struct {
		name     string
		file     string
		wantCode int
		wantBody string
	}{
		{"dot-dot-slash", "../secret.log", http.StatusBadRequest, "Invalid file name"},
		{"absolute-path", "/etc/passwd.log", http.StatusBadRequest, "Invalid file name"},
		{"subdirectory", "sub/file.log", http.StatusBadRequest, "Invalid file name"},
		{"non-log-ext", "file.txt", http.StatusBadRequest, "Only .log files are allowed"},
		{"no-extension", "logfile", http.StatusBadRequest, "Only .log files are allowed"},
		{"empty-file-param", "", http.StatusBadRequest, "'file' query parameter is required"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			url := "/stream-logs"
			if tc.file != "" {
				url += "?file=" + tc.file
			}
			req := httptest.NewRequest("GET", url, nil)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tc.wantCode {
				t.Errorf("expected status %d, got %d", tc.wantCode, rr.Code)
			}
			if !strings.Contains(rr.Body.String(), tc.wantBody) {
				t.Errorf("expected body to contain %q, got %q", tc.wantBody, rr.Body.String())
			}
		})
	}
}

func TestListFilesHandler_MixedFiles(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-mixed-")
	if err != nil {
		t.Fatalf("could not create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create various file types
	files := map[string]bool{
		"app.log":     true,
		"error.log":   true,
		"readme.md":   false,
		"config.json": false,
		"data.txt":    false,
	}
	for name := range files {
		os.WriteFile(filepath.Join(tempDir, name), []byte("content"), 0644)
	}
	// Create a subdirectory (should be excluded)
	os.Mkdir(filepath.Join(tempDir, "subdir.log"), 0755)

	handler := ListFilesHandler(tempDir)
	req := httptest.NewRequest("GET", "/list-files", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var result []string
	if err := json.Unmarshal(rr.Body.Bytes(), &result); err != nil {
		t.Fatalf("could not parse JSON: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 log files, got %d: %v", len(result), result)
	}

	for _, f := range result {
		if !strings.HasSuffix(f, ".log") {
			t.Errorf("non-log file in results: %s", f)
		}
	}
}

func TestListFilesHandler_ContentType(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "test-ct-")
	defer os.RemoveAll(tempDir)

	handler := ListFilesHandler(tempDir)
	req := httptest.NewRequest("GET", "/list-files", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", rr.Header().Get("Content-Type"))
	}
}

// flusherRecorder wraps httptest.ResponseRecorder to implement http.Flusher.
type flusherRecorder struct {
	*httptest.ResponseRecorder
}

func (f *flusherRecorder) Flush() {}

func TestStreamLogsHandler_StreamsExistingContent(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-stream-")
	if err != nil {
		t.Fatalf("could not create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	logContent := "first line\nsecond line\n"
	if err := os.WriteFile(filepath.Join(tempDir, "test.log"), []byte(logContent), 0644); err != nil {
		t.Fatalf("could not create test log: %v", err)
	}

	handler := StreamLogsHandler(tempDir)

	// Create a request with a cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	req := httptest.NewRequest("GET", "/stream-logs?file=test.log", nil).WithContext(ctx)
	rr := &flusherRecorder{httptest.NewRecorder()}

	// Cancel after a short delay to break the loop
	go func() {
		time.Sleep(1500 * time.Millisecond)
		cancel()
	}()

	handler.ServeHTTP(rr, req)

	body := rr.Body.String()
	if !strings.Contains(body, "data: first line") {
		t.Errorf("expected body to contain 'data: first line', got %q", body)
	}
	if !strings.Contains(body, "data: second line") {
		t.Errorf("expected body to contain 'data: second line', got %q", body)
	}
}

func TestStreamLogsHandler_SSEHeaders(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-sse-headers-")
	if err != nil {
		t.Fatalf("could not create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	os.WriteFile(filepath.Join(tempDir, "test.log"), []byte("line\n"), 0644)

	handler := StreamLogsHandler(tempDir)

	ctx, cancel := context.WithCancel(context.Background())
	req := httptest.NewRequest("GET", "/stream-logs?file=test.log", nil).WithContext(ctx)
	rr := &flusherRecorder{httptest.NewRecorder()}

	go func() {
		time.Sleep(1500 * time.Millisecond)
		cancel()
	}()

	handler.ServeHTTP(rr, req)

	if rr.Header().Get("Content-Type") != "text/event-stream" {
		t.Errorf("expected Content-Type text/event-stream, got %q", rr.Header().Get("Content-Type"))
	}
	if rr.Header().Get("Cache-Control") != "no-cache" {
		t.Errorf("expected Cache-Control no-cache, got %q", rr.Header().Get("Cache-Control"))
	}
}

// ----- DownloadFileHandler tests -----

func newDownloadTestDir(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

func writeLogFile(t *testing.T, dir, name, contents string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}

func TestDownloadFileHandler_Success(t *testing.T) {
	dir := newDownloadTestDir(t)
	writeLogFile(t, dir, "app.log", "hello world\nline two\n")

	req := httptest.NewRequest("GET", "/download-file?file=app.log", nil)
	rr := httptest.NewRecorder()
	DownloadFileHandler(dir)(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body=%q)", rr.Code, rr.Body.String())
	}
	if got := rr.Body.String(); got != "hello world\nline two\n" {
		t.Errorf("body mismatch: %q", got)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/octet-stream" {
		t.Errorf("Content-Type = %q", got)
	}
	wantCD := `attachment; filename="app.log"`
	if got := rr.Header().Get("Content-Disposition"); got != wantCD {
		t.Errorf("Content-Disposition = %q, want %q", got, wantCD)
	}
	if got := rr.Header().Get("Content-Length"); got != "21" {
		t.Errorf("Content-Length = %q, want 21", got)
	}
}

func TestDownloadFileHandler_MissingFileParam(t *testing.T) {
	dir := newDownloadTestDir(t)
	req := httptest.NewRequest("GET", "/download-file", nil)
	rr := httptest.NewRecorder()
	DownloadFileHandler(dir)(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestDownloadFileHandler_PathTraversal(t *testing.T) {
	dir := newDownloadTestDir(t)
	cases := []string{
		"../etc/passwd",
		"..\\windows",
		"/abs/path.log",
		"subdir/file.log",
	}
	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/download-file?file="+url.QueryEscape(name), nil)
			rr := httptest.NewRecorder()
			DownloadFileHandler(dir)(rr, req)
			if rr.Code != http.StatusBadRequest {
				t.Fatalf("expected 400 for %q, got %d", name, rr.Code)
			}
		})
	}
}

func TestDownloadFileHandler_InvalidExtension(t *testing.T) {
	dir := newDownloadTestDir(t)
	cases := []string{"config.txt", "file"}
	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/download-file?file="+name, nil)
			rr := httptest.NewRecorder()
			DownloadFileHandler(dir)(rr, req)
			if rr.Code != http.StatusBadRequest {
				t.Fatalf("expected 400 for %q, got %d", name, rr.Code)
			}
		})
	}
}

func TestDownloadFileHandler_FileNotFound(t *testing.T) {
	dir := newDownloadTestDir(t)
	req := httptest.NewRequest("GET", "/download-file?file=nonexistent.log", nil)
	rr := httptest.NewRecorder()
	DownloadFileHandler(dir)(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestDownloadFileHandler_IsDirectory(t *testing.T) {
	dir := newDownloadTestDir(t)
	if err := os.Mkdir(filepath.Join(dir, "sub.log"), 0755); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("GET", "/download-file?file=sub.log", nil)
	rr := httptest.NewRecorder()
	DownloadFileHandler(dir)(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for directory, got %d", rr.Code)
	}
}

func TestDownloadFileHandler_LargeFile(t *testing.T) {
	dir := newDownloadTestDir(t)
	const size = 10 * 1024 * 1024 // 10 MB
	buf := make([]byte, size)
	for i := range buf {
		buf[i] = byte(i % 251)
	}
	if err := os.WriteFile(filepath.Join(dir, "big.log"), buf, 0644); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("GET", "/download-file?file=big.log", nil)
	rr := httptest.NewRecorder()
	DownloadFileHandler(dir)(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if rr.Body.Len() != size {
		t.Fatalf("size mismatch: got %d, want %d", rr.Body.Len(), size)
	}
	got := sha256.Sum256(rr.Body.Bytes())
	want := sha256.Sum256(buf)
	if got != want {
		t.Errorf("hash mismatch: got %x, want %x", got, want)
	}
}

func TestDownloadFileHandler_ClientAbort(t *testing.T) {
	dir := newDownloadTestDir(t)
	const size = 50 * 1024 * 1024 // 50 MB
	buf := make([]byte, size)
	if err := os.WriteFile(filepath.Join(dir, "big.log"), buf, 0644); err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(DownloadFileHandler(dir)))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", srv.URL+"/?file=big.log", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer resp.Body.Close()

	// Read a small prefix, then cancel mid-stream.
	prefix := make([]byte, 4096)
	if _, err := io.ReadFull(resp.Body, prefix); err != nil {
		t.Fatalf("read prefix: %v", err)
	}
	cancel()

	// Expect the next read to fail because the context was cancelled.
	_, err = io.Copy(io.Discard, resp.Body)
	if err == nil {
		t.Fatalf("expected error after cancel, got nil (server may have buffered the whole response)")
	}
	// Test passes if the server didn't panic and returned the prefix successfully.
}
