package config

import (
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Server struct {
		Port string
	}
	Registry struct {
		Address string
	}
	Function struct {
		PortDetectionTimeout time.Duration
		DataDir              string
	}
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	cfg := &Config{}

	// Server configuration
	cfg.Server.Port = "8080"

	// Registry configuration
	cfg.Registry.Address = "localhost:5000"

	// Function configuration
	cfg.Function.PortDetectionTimeout = 10 * time.Second
	cfg.Function.DataDir = "./data"

	return cfg
}
