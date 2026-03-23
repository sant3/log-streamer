package main

import (
	"flag"
	"log"
	"os"
	"strconv"
	"strings"
)

// Config holds all runtime configuration for the streamer server.
type Config struct {
	Port        int
	HTTPSPort   int
	EnableHTTPS bool
	CertPath    string
	KeyPath     string
	LogsDir     string
	CORSOrigins []string
	AllowedIPs  map[string]bool
	JWTSecret   []byte
}

// DefaultConfig returns a Config populated with default values.
func DefaultConfig() *Config {
	return &Config{
		Port:        5005,
		HTTPSPort:   8443,
		EnableHTTPS: true,
		CertPath:    "cert-streamer.pem",
		KeyPath:     "decrypted_key.pem",
		LogsDir:     ".",
		CORSOrigins: []string{"http://localhost:3000", "https://anotherdomain.com", "http://localhost:1972"},
		AllowedIPs: map[string]bool{
			"192.168.1.100": true,
			"203.0.113.0":   true,
		},
		JWTSecret: []byte("qwertyuiopasdfghjklzxcvbnm123456"),
	}
}

// ApplyEnvOverrides reads STREAMER_* environment variables and overrides the
// corresponding Config fields. This allows environment variables to take
// precedence over CLI flags and defaults.
func (cfg *Config) ApplyEnvOverrides() {
	if v := os.Getenv("STREAMER_PORT"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil {
			log.Fatalf("Error parsing port from $STREAMER_PORT: %v", err)
		}
		cfg.Port = p
	}

	if v := os.Getenv("STREAMER_HTTPS_PORT"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil {
			log.Fatalf("Error parsing HTTPS port from $STREAMER_HTTPS_PORT: %v", err)
		}
		cfg.HTTPSPort = p
	}

	if v := os.Getenv("STREAMER_ENABLE_HTTPS"); v != "" {
		lower := strings.ToLower(v)
		cfg.EnableHTTPS = lower == "1" || lower == "true" || lower == "yes"
	}

	if v := os.Getenv("STREAMER_TLS_CERT"); v != "" {
		cfg.CertPath = v
	}
	if v := os.Getenv("STREAMER_TLS_KEY"); v != "" {
		cfg.KeyPath = v
	}

	if v := os.Getenv("STREAMER_LOGS_DIR"); v != "" {
		cfg.LogsDir = v
	}

	if v := os.Getenv("STREAMER_CORS_ORIGINS"); v != "" {
		parts := strings.Split(v, ",")
		cfg.CORSOrigins = nil
		for _, p := range parts {
			cfg.CORSOrigins = append(cfg.CORSOrigins, strings.TrimSpace(p))
		}
	}

	if v := os.Getenv("STREAMER_ALLOWED_IPS"); v != "" {
		cfg.AllowedIPs = map[string]bool{}
		for _, ip := range strings.Split(v, ",") {
			ip = strings.TrimSpace(ip)
			if ip != "" {
				cfg.AllowedIPs[ip] = true
			}
		}
	}

	if v := os.Getenv("STREAMER_JWT_SECRET"); v != "" {
		cfg.JWTSecret = []byte(v)
	}
}

// ParseConfig parses CLI flags and environment variables, applying the
// precedence: Environment variables > CLI flags > Defaults.
func ParseConfig() *Config {
	cfg := DefaultConfig()

	portPtr := flag.Int("port", cfg.Port, "Listening port")
	httpsPortPtr := flag.Int("httpsPort", cfg.HTTPSPort, "HTTPS listening port")
	enableHTTPSPtr := flag.Bool("enableHTTPS", cfg.EnableHTTPS, "Enable HTTPS server")
	certPathPtr := flag.String("tlsCert", cfg.CertPath, "Path to TLS certificate")
	keyPathPtr := flag.String("tlsKey", cfg.KeyPath, "Path to TLS private key")
	logsDirPtr := flag.String("logsDir", cfg.LogsDir, "Directory where .log files are located")
	corsOriginsPtr := flag.String("corsOrigins", strings.Join(cfg.CORSOrigins, ","), "Comma-separated list of allowed CORS origins (use * for all)")
	allowedIPsPtr := flag.String("allowedIPs", "", "Comma-separated list of allowed client IPs (empty to disable whitelist)")
	jwtSecretPtr := flag.String("jwtSecret", string(cfg.JWTSecret), "JWT secret for token validation")
	flag.Parse()

	// Apply CLI flag values
	cfg.Port = *portPtr
	cfg.HTTPSPort = *httpsPortPtr
	cfg.EnableHTTPS = *enableHTTPSPtr
	cfg.CertPath = *certPathPtr
	cfg.KeyPath = *keyPathPtr
	cfg.LogsDir = *logsDirPtr

	corsCfg := *corsOriginsPtr
	if corsCfg != "" {
		parts := strings.Split(corsCfg, ",")
		cfg.CORSOrigins = nil
		for _, p := range parts {
			cfg.CORSOrigins = append(cfg.CORSOrigins, strings.TrimSpace(p))
		}
	}

	ipsCfg := *allowedIPsPtr
	if ipsCfg != "" {
		cfg.AllowedIPs = map[string]bool{}
		for _, ip := range strings.Split(ipsCfg, ",") {
			ip = strings.TrimSpace(ip)
			if ip != "" {
				cfg.AllowedIPs[ip] = true
			}
		}
	}

	if *jwtSecretPtr != string(DefaultConfig().JWTSecret) {
		cfg.JWTSecret = []byte(*jwtSecretPtr)
	}

	// Environment variables override CLI flags
	cfg.ApplyEnvOverrides()

	return cfg
}
