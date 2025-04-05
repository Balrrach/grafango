package config

import (
	"flag"
	"log/slog"
	"os"
	"time"
)

// Config holds application configuration
type Config struct {
	Port          int
	MetricsPath   string
	ScrapeInterval time.Duration
	LogLevel      string
}

// ParseFlags parses command line flags and returns config
func ParseFlags() *Config {
	config := &Config{}

	flag.IntVar(&config.Port, "port", 8080, "Port to serve metrics on")
	flag.StringVar(&config.MetricsPath, "metrics-path", "/metrics", "Path to expose metrics on")
	flag.DurationVar(&config.ScrapeInterval, "scrape-interval", 5*time.Second, "Interval between metric collections")
	flag.StringVar(&config.LogLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	flag.Parse()

	return config
}

// SetupLogger configures structured logging
func SetupLogger(logLevel string) {
	var level slog.Level
	switch logLevel {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})
	
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

