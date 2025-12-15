package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
)

func TestLoad_NoConfig(t *testing.T) {
	// Set XDG vars to empty temp dir to ensure no config is found
	tmpDir := t.TempDir()

	// Mock XDG ConfigHome
	origConfigHome := xdg.ConfigHome
	xdg.ConfigHome = tmpDir
	t.Cleanup(func() { xdg.ConfigHome = origConfigHome })

	// Also clear ConfigDirs to avoid picking up distinct system configs
	origConfigDirs := xdg.ConfigDirs
	xdg.ConfigDirs = []string{tmpDir}
	t.Cleanup(func() { xdg.ConfigDirs = origConfigDirs })

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if cfg == nil {
		t.Fatal("Expected config, got nil")
	}
	if cfg.Otel.Exporter.Traces.Endpoint != "" {
		t.Errorf("Expected empty endpoint, got %s", cfg.Otel.Exporter.Traces.Endpoint)
	}
}

func TestLoad_WithConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "workctl")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	content := `
otel:
  exporter:
    traces:
      endpoint: "http://localhost:4318/v1/traces"
      headers:
        Authorization: "Bearer token"
    metrics:
      endpoint: "http://localhost:4318/v1/metrics"
`
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Mock XDG ConfigHome
	origConfigHome := xdg.ConfigHome
	xdg.ConfigHome = tmpDir
	t.Cleanup(func() { xdg.ConfigHome = origConfigHome })

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.Otel.Exporter.Traces.Endpoint != "http://localhost:4318/v1/traces" {
		t.Errorf("Expected traces endpoint http://localhost:4318/v1/traces, got %s", cfg.Otel.Exporter.Traces.Endpoint)
	}
	// Viper lowercases map keys from config
	if cfg.Otel.Exporter.Traces.Headers["authorization"] != "Bearer token" {
		t.Errorf("Expected Authorization header 'Bearer token', got %s", cfg.Otel.Exporter.Traces.Headers["authorization"])
	}
}
