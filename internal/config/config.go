// Package config reads Bifrost's settings from environment variables.
package config

import (
	"net"
	"os"
	"path/filepath"
)

const (
	defaultAddr    = ":8080"
	defaultDataDir = "/data"
)

// Config holds the settings the server needs at start-up.
type Config struct {
	// Addr is the address the HTTP server listens on (BIFROST_ADDR).
	Addr string
	// DataDir is the directory that holds the database and uploads (BIFROST_DATA_DIR).
	DataDir string
}

// Load reads the configuration from the environment, applying defaults.
func Load() Config {
	return Config{
		Addr:    getenv("BIFROST_ADDR", defaultAddr),
		DataDir: getenv("BIFROST_DATA_DIR", defaultDataDir),
	}
}

// DBPath is the location of the SQLite database file.
func (c Config) DBPath() string {
	return filepath.Join(c.DataDir, "bifrost.db")
}

// HealthURL is the address the healthcheck subcommand requests. It always
// targets the loopback interface, whatever host the server is bound to.
func (c Config) HealthURL() string {
	_, port, err := net.SplitHostPort(c.Addr)
	if err != nil || port == "" {
		port = "8080"
	}
	return "http://" + net.JoinHostPort("127.0.0.1", port) + "/healthz"
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
