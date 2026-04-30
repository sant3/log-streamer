package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

var (
	version   = "1.2.1"
	buildDate string
)

// gracefulShutdown installs the platform-specific signal handler and exits
// cleanly on interrupt.
func gracefulShutdown() {
	c := make(chan os.Signal, 1)
	setupSignalHandler(c)

	go func() {
		<-c
		log.Println("Received interrupt signal")
		log.Println("Server stopped")
		os.Exit(0)
	}()
}

// setupRouter configures all HTTP routes using the provided Config.
func setupRouter(cfg *Config) *http.ServeMux {
	cors := func(next http.HandlerFunc) http.HandlerFunc {
		return CORSMiddleware(cfg.CORSOrigins, next)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/stream-logs", cors(StreamLogsHandler(cfg.LogsDir)))
	mux.HandleFunc("/alive", cors(AliveHandler))
	mux.HandleFunc("/restart", RestartHandler)
	mux.HandleFunc("/stop", StopHandler)
	mux.HandleFunc("/version", cors(VersionHandler))
	mux.HandleFunc("/list-files", cors(ListFilesHandler(cfg.LogsDir)))
	mux.HandleFunc("/download-file", cors(DownloadFileHandler(cfg.LogsDir)))
	return mux
}

// startServer starts an HTTP or HTTPS server on the given port with the
// provided handler. For TLS, provide non-empty certPath and keyPath.
func startServer(port int, useTLS bool, certPath string, keyPath string, mux http.Handler) {
	addr := fmt.Sprintf(":%d", port)
	if useTLS {
		log.Printf("HTTPS server started on port: %d", port)
		log.Fatal(http.ListenAndServeTLS(addr, certPath, keyPath, mux))
	} else {
		log.Printf("HTTP server started on port: %d", port)
		log.Fatal(http.ListenAndServe(addr, mux))
	}
}

func main() {
	logFile, err := os.OpenFile("streamer.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
	defer logFile.Close()

	multiWriter := io.MultiWriter(logFile, os.Stdout)
	log.SetOutput(multiWriter)

	gracefulShutdown()

	log.Printf("..... Streamer application version %s - buildDate %s starting ......", version, buildDate)

	cfg := ParseConfig()

	router := setupRouter(cfg)

	go startServer(cfg.Port, false, "", "", router)
	if cfg.EnableHTTPS {
		startServer(cfg.HTTPSPort, true, cfg.CertPath, cfg.KeyPath, router)
	} else {
		select {}
	}
}
