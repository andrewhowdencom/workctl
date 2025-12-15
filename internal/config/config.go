package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/spf13/viper"
)

// Config holds the top-level configuration.
type Config struct {
	Otel OtelConfig `mapstructure:"otel"`
}

// OtelConfig holds OpenTelemetry configuration.
type OtelConfig struct {
	Exporter ExporterConfig `mapstructure:"exporter"`
}

// ExporterConfig holds exporter configuration for traces and metrics.
type ExporterConfig struct {
	Traces  OtlpConfig `mapstructure:"traces"`
	Metrics OtlpConfig `mapstructure:"metrics"`
}

// OtlpConfig holds configuration for a specific OTLP signal (traces or metrics).
type OtlpConfig struct {
	Endpoint string            `mapstructure:"endpoint"`
	Headers  map[string]string `mapstructure:"headers"`
}

// Load reads the configuration from file.
// It searches in XDG config directories and /etc/workctl.
func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// Search paths
	// 1. $XDG_CONFIG_HOME/workctl (e.g. ~/.config/workctl)
	v.AddConfigPath(filepath.Join(xdg.ConfigHome, "workctl"))

	// 2. $XDG_CONFIG_DIRS/workctl (e.g. /etc/xdg/workctl)
	for _, dir := range xdg.ConfigDirs {
		v.AddConfigPath(filepath.Join(dir, "workctl"))
	}

	// 3. /etc/workctl (fallback)
	v.AddConfigPath("/etc/workctl")

	// Allow environment variable overrides
	// e.g. WORKCTL_OTEL_EXPORTER_TRACES_ENDPOINT
	v.SetEnvPrefix("workctl")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			// It's okay if config file is not found, we just return empty config
			// or default values if we had any.
			return &Config{}, nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
