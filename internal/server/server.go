package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"grafango/internal/config"
)

// Start initializes and runs the HTTP server
func Start(ctx context.Context, cfg *config.Config) {
	// Set up HTTP server
	mux := http.NewServeMux()
	mux.Handle(cfg.MetricsPath, promhttp.Handler())
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `<html>
<head><title>System Metrics Exporter</title></head>
<body>
<h1>System Metrics Exporter</h1>
<p><a href="%s">Metrics</a></p>
</body>
</html>`, cfg.MetricsPath)
	})
	
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: mux,
	}
	
	// Start server in a goroutine
	go func() {
		slog.Info("Starting metrics server", 
			"address", server.Addr, 
			"metrics_path", cfg.MetricsPath)
		
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server error", "error", err)
			os.Exit(1)
		}
	}()
	
	// Wait for termination signal
	<-ctx.Done()
	slog.Info("Shutting down server")
	
	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("Server shutdown error", "error", err)
		os.Exit(1)
	}
	
	slog.Info("Server gracefully stopped")
}

