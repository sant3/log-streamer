package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"

	// "os/signal" // handled in platform-specific files
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Version of the application
const version = "0.0.1-SNAPSHOT"

var buildDate string

// jwtSecret is initialized from flags/env in main
var jwtSecret = []byte("qwertyuiopasdfghjklzxcvbnm123456")

// runtime configuration (initialized in main)
var (
	logsDir           = "."
	allowedOriginsCfg = []string{"http://localhost:3000", "https://anotherdomain.com", "http://localhost:1972"}
)

// streamLogs streams the content of a selected .log file using Server-Sent Events.
// It enforces that the file resides inside the configured logs directory and
// terminates when the client disconnects.
func streamLogs(w http.ResponseWriter, r *http.Request) {
	// set CORS header
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	fileName := r.URL.Query().Get("file")
	if fileName == "" {
		fileName = "mylog.log"
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

	fullPath := filepath.Join(logsDir, fileName)

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

// isAlive is a basic liveness probe that returns HTTP 200 when the process is healthy.
func isAlive(w http.ResponseWriter, r *http.Request) {
	// set CORS header
	// w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization,X-CSRF-Token")
	// w.Header().Set("Access-Control-Allow-Origin", "*")
	// w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	// w.Header().Set("Connection", "keep-alive")

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK\n")
}

// gracefulShutdown installs the platform-specific signal handler and exits cleanly on interrupt.
func gracefulShutdown() {
	c := make(chan os.Signal, 1)
	// platform-specific setup
	setupSignalHandler(c)

	go func() {
		<-c
		log.Println("Received interrupt signal")
		// Esegui le operazioni di pulizia necessarie prima di terminare
		// ...

		log.Println("Server stopped")
		os.Exit(0)
	}()
}

// stopHandler attempts a graceful application shutdown, allowing cleanup before exit.
func stopHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf(".....STOP INVOKED.....")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Stopping application...\n")

	// Signal the main goroutine to stop gracefully
	stopCh := make(chan bool)
	go func() {
		// Perform any necessary cleanup tasks here (e.g., closing connections, saving data)
		// ...

		close(stopCh)
	}()

	// Wait for cleanup to complete before exiting
	select {
	case <-stopCh:
		log.Println("Cleanup completed. Exiting...")
		os.Exit(0)
	case <-time.After(30 * time.Second): // Optional timeout for cleanup
		log.Println("Cleanup timed out. Exiting forcefully...")
		os.Exit(1)
	}
}

// restartHandler relaunches the current binary and terminates the current process.
// Note: behavior can vary across platforms; confirm support on Windows.
func restartHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf(".....RESTART INVOKED.....")
	// Crea un nuovo processo figlio
	cmd := exec.Command(os.Args[0], os.Args[1:]...)
	cmd.Start()

	// Termina il processo corrente in modo compatibile con il sistema operativo
	terminateSelf()
}

// listLogFiles returns the list of available .log files under the configured logs directory.
func listLogFiles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	files, err := os.ReadDir(logsDir)
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

// startServer starts an HTTP or HTTPS server on the given port with the provided mux.
// For TLS, provide non-empty certPath and keyPath.
func startServer(port int, useTLS bool, certPath string, keyPath string) {
	mux := http.NewServeMux() // Create a new ServeMux for each server
	mux.HandleFunc("/stream-logs", streamLogs)
	// mux.HandleFunc("/alive", corsMiddleware(authenticateJWT(isAlive)))
	mux.HandleFunc("/alive", corsMiddleware(isAlive))
	mux.HandleFunc("/restart", restartHandler)
	mux.HandleFunc("/stop", stopHandler)
	mux.HandleFunc("/version", getVersion)
	mux.HandleFunc("/list-files", listLogFiles)

	addr := fmt.Sprintf(":%d", port)
	if useTLS {
		log.Printf("HTTPS server started on port: %d", port)
		log.Fatal(http.ListenAndServeTLS(addr, certPath, keyPath, mux))
	} else {
		log.Printf("HTTP server started on port: %d", port)
		log.Fatal(http.ListenAndServe(addr, mux))
	}
}

// getVersion returns the application version and build date.
func getVersion(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Version: %s \nBuild date: %s", version, buildDate)
}

var allowedIPs = map[string]bool{
	"192.168.1.100": true,
	"203.0.113.0":   true,
}

// ipWhitelist restricts access to the provided handler to clients whose IP is
// included in the configured whitelist. When empty, the whitelist is effectively disabled.
func ipWhitelist(next http.HandlerFunc) http.HandlerFunc {
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

// authenticateJWT validates the Authorization: Bearer <token> header using the configured secret.
// On success, it forwards the request to the next handler; otherwise, it returns 403.
func authenticateJWT(next http.HandlerFunc) http.HandlerFunc {
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
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Forbidden", http.StatusForbidden)
			log.Printf("Token is NOT valid")
			return
		}

		next.ServeHTTP(w, r)
	}
}

// corsMiddleware sets CORS headers based on the configured allowed origins.
// Use "*" to allow any origin.
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		for _, allowedOrigin := range allowedOriginsCfg {
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

func main() {
	// Open log file
	logFile, err := os.OpenFile("streamer.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}
	defer logFile.Close()

	// Create a MultiWriter that writes to both file and console
	multiWriter := io.MultiWriter(logFile, os.Stdout)

	// Redirect standard logger to both file and console
	log.SetOutput(multiWriter)

	gracefulShutdown()

	// Log the version at startup
	log.Printf("..... Streamer application version %s - buildDate %s starting ......", version, buildDate)

	// cli params
	portPtr := flag.Int("port", 5005, "Listening port")
	httpsPortPtr := flag.Int("httpsPort", 8443, "HTTPS listening port")
	enableHTTPSPtr := flag.Bool("enableHTTPS", true, "Enable HTTPS server")
	certPathPtr := flag.String("tlsCert", "cert-streamer.pem", "Path to TLS certificate")
	keyPathPtr := flag.String("tlsKey", "decrypted_key.pem", "Path to TLS private key")
	logsDirPtr := flag.String("logsDir", ".", "Directory where .log files are located")
	corsOriginsPtr := flag.String("corsOrigins", "http://localhost:3000,https://anotherdomain.com,http://localhost:1972", "Comma-separated list of allowed CORS origins (use * for all)")
	allowedIPsPtr := flag.String("allowedIPs", "", "Comma-separated list of allowed client IPs (empty to disable whitelist)")
	jwtSecretPtr := flag.String("jwtSecret", string(jwtSecret), "JWT secret for token validation")
	flag.Parse()

	portEnv := os.Getenv("STREAMER_PORT")
	httpsPortEnv := os.Getenv("STREAMER_HTTPS_PORT")
	enableHTTPSEnv := os.Getenv("STREAMER_ENABLE_HTTPS")
	certPathEnv := os.Getenv("STREAMER_TLS_CERT")
	keyPathEnv := os.Getenv("STREAMER_TLS_KEY")
	logsDirEnv := os.Getenv("STREAMER_LOGS_DIR")
	corsOriginsEnv := os.Getenv("STREAMER_CORS_ORIGINS")
	allowedIPsEnv := os.Getenv("STREAMER_ALLOWED_IPS")
	jwtSecretEnv := os.Getenv("STREAMER_JWT_SECRET")

	var port int
	var httpsPort int

	if *portPtr != 5005 {
		port = *portPtr
	} else if portEnv != "" {
		var err error
		port, err = strconv.Atoi(portEnv)
		if err != nil {
			log.Fatalf("Error parsing port from $STREAMER_PORT: %v", err)
		}
	} else {
		port = 5005
	}

	if *httpsPortPtr != 8443 {
		httpsPort = *httpsPortPtr
	} else if httpsPortEnv != "" {
		var err error
		httpsPort, err = strconv.Atoi(httpsPortEnv)
		if err != nil {
			log.Fatalf("Error parsing HTTPS port from $STREAMER_HTTPS_PORT: %v", err)
		}
	} else {
		httpsPort = 8443
	}

	// Enable HTTPS
	useHTTPS := *enableHTTPSPtr
	if enableHTTPSEnv != "" {
		lower := strings.ToLower(enableHTTPSEnv)
		useHTTPS = lower == "1" || lower == "true" || lower == "yes"
	}

	// TLS cert/key paths
	certPath := *certPathPtr
	keyPath := *keyPathPtr
	if certPathEnv != "" {
		certPath = certPathEnv
	}
	if keyPathEnv != "" {
		keyPath = keyPathEnv
	}

	// Logs directory
	logsDir = *logsDirPtr
	if logsDirEnv != "" {
		logsDir = logsDirEnv
	}

	// CORS origins
	corsCfg := *corsOriginsPtr
	if corsOriginsEnv != "" {
		corsCfg = corsOriginsEnv
	}
	if corsCfg != "" {
		parts := strings.Split(corsCfg, ",")
		allowedOriginsCfg = nil
		for _, p := range parts {
			allowedOriginsCfg = append(allowedOriginsCfg, strings.TrimSpace(p))
		}
	}

	// Allowed IPs
	ipsCfg := *allowedIPsPtr
	if allowedIPsEnv != "" {
		ipsCfg = allowedIPsEnv
	}
	if ipsCfg != "" {
		allowedIPs = map[string]bool{}
		for _, ip := range strings.Split(ipsCfg, ",") {
			ip = strings.TrimSpace(ip)
			if ip != "" {
				allowedIPs[ip] = true
			}
		}
	}

	// JWT secret
	if jwtSecretEnv != "" {
		jwtSecret = []byte(jwtSecretEnv)
	} else if *jwtSecretPtr != string(jwtSecret) {
		jwtSecret = []byte(*jwtSecretPtr)
	}

	go startServer(port, false, "", "") // Start HTTP server
	if useHTTPS {
		startServer(httpsPort, true, certPath, keyPath) // Start HTTPS server
	} else {
		select {}
	}

}
