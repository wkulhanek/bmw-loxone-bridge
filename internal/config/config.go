package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	ClientID             string
	VIN                  string
	ListenAddr           string
	DataDir              string
	LogLevel             string
	RefreshInterval      time.Duration
	ActiveRefreshInterval time.Duration
}

func Load() (*Config, error) {
	cfg := &Config{
		ClientID:        os.Getenv("BMW_CLIENT_ID"),
		VIN:             os.Getenv("BMW_VIN"),
		ListenAddr:      envOrDefault("LISTEN_ADDR", ":8400"),
		DataDir:         envOrDefault("DATA_DIR", "/data"),
		LogLevel:        envOrDefault("LOG_LEVEL", "info"),
		RefreshInterval:      parseDurationMinutes("REFRESH_MINUTES", 30),
		ActiveRefreshInterval: parseDurationMinutes("ACTIVE_REFRESH_MINUTES", 5),
	}
	if cfg.ClientID == "" {
		return nil, fmt.Errorf("BMW_CLIENT_ID environment variable is required")
	}
	if cfg.VIN == "" {
		return nil, fmt.Errorf("BMW_VIN environment variable is required")
	}
	return cfg, nil
}

func (c *Config) SessionPath() string {
	return c.DataDir + "/session.json"
}

func envOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func parseDurationMinutes(key string, defaultMinutes int) time.Duration {
	if v := os.Getenv(key); v != "" {
		if m, err := strconv.Atoi(v); err == nil && m > 0 {
			return time.Duration(m) * time.Minute
		}
	}
	return time.Duration(defaultMinutes) * time.Minute
}
