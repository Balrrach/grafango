package main

import (
	"context"
	"os/signal"
	"syscall"
	
	"grafango/internal/config"
	"grafango/internal/collector"
	"grafango/internal/server"
)

func main() {
	// Parse command line arguments
	cfg := config.ParseFlags()
	
	// Set up logging
	config.SetupLogger(cfg.LogLevel)
	
	// Create context that listens for termination signals
	ctx, stop := signal.NotifyContext(context.Background(), 
		syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	
	// Create and start metrics collector
	metricCollector := collector.NewMetricsCollector(cfg.ScrapeInterval)
	metricCollector.Start(ctx)
	
	// Start HTTP server and handle graceful shutdown
	server.Start(ctx, cfg)
}

