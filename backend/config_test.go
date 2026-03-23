package main

import (
	"testing"
)

func TestConfigDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Port != 5005 {
		t.Errorf("expected default port 5005, got %d", cfg.Port)
	}
	if cfg.HTTPSPort != 8443 {
		t.Errorf("expected default HTTPS port 8443, got %d", cfg.HTTPSPort)
	}
	if !cfg.EnableHTTPS {
		t.Error("expected EnableHTTPS to be true by default")
	}
	if cfg.LogsDir != "." {
		t.Errorf("expected default LogsDir '.', got %q", cfg.LogsDir)
	}
	if len(cfg.CORSOrigins) != 3 {
		t.Errorf("expected 3 default CORS origins, got %d", len(cfg.CORSOrigins))
	}
	if len(cfg.AllowedIPs) != 2 {
		t.Errorf("expected 2 default allowed IPs, got %d", len(cfg.AllowedIPs))
	}
	if len(cfg.JWTSecret) == 0 {
		t.Error("expected non-empty default JWT secret")
	}
}

func TestConfigApplyEnvOverrides_AllFields(t *testing.T) {
	t.Setenv("STREAMER_PORT", "9090")
	t.Setenv("STREAMER_HTTPS_PORT", "9443")
	t.Setenv("STREAMER_ENABLE_HTTPS", "false")
	t.Setenv("STREAMER_TLS_CERT", "/tmp/cert.pem")
	t.Setenv("STREAMER_TLS_KEY", "/tmp/key.pem")
	t.Setenv("STREAMER_LOGS_DIR", "/var/log")
	t.Setenv("STREAMER_CORS_ORIGINS", "http://myapp.com,http://other.com")
	t.Setenv("STREAMER_ALLOWED_IPS", "10.0.0.1,10.0.0.2")
	t.Setenv("STREAMER_JWT_SECRET", "super-secret")

	cfg := DefaultConfig()
	cfg.ApplyEnvOverrides()

	if cfg.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Port)
	}
	if cfg.HTTPSPort != 9443 {
		t.Errorf("expected HTTPS port 9443, got %d", cfg.HTTPSPort)
	}
	if cfg.EnableHTTPS {
		t.Error("expected EnableHTTPS to be false")
	}
	if cfg.CertPath != "/tmp/cert.pem" {
		t.Errorf("expected cert path /tmp/cert.pem, got %q", cfg.CertPath)
	}
	if cfg.KeyPath != "/tmp/key.pem" {
		t.Errorf("expected key path /tmp/key.pem, got %q", cfg.KeyPath)
	}
	if cfg.LogsDir != "/var/log" {
		t.Errorf("expected logs dir /var/log, got %q", cfg.LogsDir)
	}
	if len(cfg.CORSOrigins) != 2 {
		t.Errorf("expected 2 CORS origins, got %d", len(cfg.CORSOrigins))
	}
	if cfg.CORSOrigins[0] != "http://myapp.com" {
		t.Errorf("unexpected first CORS origin: %q", cfg.CORSOrigins[0])
	}
	if len(cfg.AllowedIPs) != 2 {
		t.Errorf("expected 2 allowed IPs, got %d", len(cfg.AllowedIPs))
	}
	if !cfg.AllowedIPs["10.0.0.1"] {
		t.Error("expected 10.0.0.1 in allowed IPs")
	}
	if string(cfg.JWTSecret) != "super-secret" {
		t.Errorf("expected JWT secret 'super-secret', got %q", string(cfg.JWTSecret))
	}
}

func TestConfigApplyEnvOverrides_HTTPSVariations(t *testing.T) {
	tests := []struct {
		envValue string
		expected bool
	}{
		{"true", true},
		{"1", true},
		{"yes", true},
		{"false", false},
		{"0", false},
		{"no", false},
		{"TRUE", true},
		{"YES", true},
	}

	for _, tc := range tests {
		t.Run(tc.envValue, func(t *testing.T) {
			t.Setenv("STREAMER_ENABLE_HTTPS", tc.envValue)
			cfg := DefaultConfig()
			cfg.ApplyEnvOverrides()
			if cfg.EnableHTTPS != tc.expected {
				t.Errorf("for env value %q, expected EnableHTTPS=%v, got %v", tc.envValue, tc.expected, cfg.EnableHTTPS)
			}
		})
	}
}

func TestConfigApplyEnvOverrides_PartialOverride(t *testing.T) {
	t.Setenv("STREAMER_PORT", "7070")

	cfg := DefaultConfig()
	cfg.ApplyEnvOverrides()

	if cfg.Port != 7070 {
		t.Errorf("expected port 7070, got %d", cfg.Port)
	}
	// Other fields should remain at defaults
	if cfg.HTTPSPort != 8443 {
		t.Errorf("expected HTTPS port to remain 8443, got %d", cfg.HTTPSPort)
	}
	if cfg.LogsDir != "." {
		t.Errorf("expected LogsDir to remain '.', got %q", cfg.LogsDir)
	}
}
