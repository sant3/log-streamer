package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// StreamLogsHandler returns an HTTP handler that streams the content of a
// selected .log file using Server-Sent Events. It enforces that the file
// resides inside the configured logs directory and terminates when the client
// disconnects.
func StreamLogsHandler(logsDirPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		fileName := r.URL.Query().Get("file")
		if fileName == "" {
			http.Error(w, "Error: 'file' query parameter is required", http.StatusBadRequest)
			return
		}

		// Prevent directory traversal: only allow base file names within logsDir
		if filepath.Base(fileName) != fileName {
			http.Error(w, "Invalid file name", http.StatusBadRequest)
			return
		}
		if !strings.HasSuffix(fileName, ".log") {
			http.Error(w, "Only .log files are allowed", http.StatusBadRequest)
			return
		}

		fullPath := filepath.Join(logsDirPath, fileName)

		fileInfo, err := os.Stat(fullPath)
		if err != nil || fileInfo.IsDir() {
			log.Printf("Log file not found or is a directory: %s", fullPath)
			http.Error(w, "Error: log file not found or is not a regular file", http.StatusInternalServerError)
			return
		}

		file, err := os.Open(fullPath)
		if err != nil {
			log.Printf("File not found %s", fullPath)
			fmt.Fprintf(w, "data: Error: File not found: %s\n\n", fileName)
			w.(http.Flusher).Flush()
			return
		}
		defer file.Close()

		offset := int64(0)
		for {
			select {
			case <-r.Context().Done():
				log.Printf("Client closed connection for %s", fileName)
				return
			default:
			}
			time.Sleep(1 * time.Second)

			stat, err := file.Stat()
			if err != nil {
				fmt.Fprintf(w, "data: Error: Could not get file info\n\n")
				w.(http.Flusher).Flush()
				return
			}

			if stat.Size() < offset {
				offset = 0 // If file is truncated, restart
			}

			file.Seek(offset, io.SeekStart)
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				fmt.Fprintf(w, "data: %s\n\n", scanner.Text())
			}
			offset, _ = file.Seek(0, io.SeekCurrent)

			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}
}

// AliveHandler is a basic liveness probe that returns HTTP 200 when the
// process is healthy.
func AliveHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}

// VersionHandler returns the application version and build date as JSON.
func VersionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]string{
		"version":   version,
		"buildDate": buildDate,
	}
	json.NewEncoder(w).Encode(response)
}

// ListFilesHandler returns the list of available .log files under the
// configured logs directory as a JSON array.
func ListFilesHandler(logsDirPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")

		files, err := os.ReadDir(logsDirPath)
		if err != nil {
			log.Printf("Error reading directory: %v", err)
			http.Error(w, "Unable to read directory", http.StatusInternalServerError)
			return
		}

		var logFiles []string
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".log") {
				logFiles = append(logFiles, file.Name())
			}
		}

		jsonResponse, err := json.Marshal(logFiles)
		if err != nil {
			log.Printf("Error marshaling JSON: %v", err)
			http.Error(w, "Unable to marshal JSON", http.StatusInternalServerError)
			return
		}

		w.Write(jsonResponse)
	}
}

// StopHandler attempts a graceful application shutdown, allowing cleanup
// before exit.
func StopHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf(".....STOP INVOKED.....")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Stopping application...\n")

	stopCh := make(chan bool)
	go func() {
		close(stopCh)
	}()

	select {
	case <-stopCh:
		log.Println("Cleanup completed. Exiting...")
		os.Exit(0)
	case <-time.After(30 * time.Second):
		log.Println("Cleanup timed out. Exiting forcefully...")
		os.Exit(1)
	}
}

// RestartHandler relaunches the current binary and terminates the current
// process. Note: behavior can vary across platforms.
func RestartHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf(".....RESTART INVOKED.....")
	cmd := exec.Command(os.Args[0], os.Args[1:]...)
	cmd.Start()
	terminateSelf()
}
