package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

// Config holds application configuration
type Config struct {
	Port          int
	MetricsPath   string
	ScrapeInterval time.Duration
	LogLevel      string
}

// MetricsCollector manages the metrics collection
type MetricsCollector struct {
	cpuUsage      prometheus.Gauge
	cpuUsagePerCore *prometheus.GaugeVec
	memUsage      prometheus.Gauge
	memAvailable  prometheus.Gauge
	memTotal      prometheus.Gauge
	diskUsage     *prometheus.GaugeVec
	diskTotal     *prometheus.GaugeVec
	scrapeInterval time.Duration
}

// NewMetricsCollector creates and registers all metrics
func NewMetricsCollector(scrapeInterval time.Duration) *MetricsCollector {
	cpuUsage := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "system_cpu_usage_percent",
		Help: "Current CPU usage percentage (all cores)",
	})

	cpuUsagePerCore := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "system_cpu_core_usage_percent",
			Help: "Current CPU usage percentage per core",
		},
		[]string{"core"},
	)

	memUsage := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "system_memory_usage_percent",
		Help: "Current memory usage percentage",
	})

	memAvailable := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "system_memory_available_bytes",
		Help: "Available memory in bytes",
	})

	memTotal := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "system_memory_total_bytes",
		Help: "Total memory in bytes",
	})

	diskUsage := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "system_disk_usage_percent",
			Help: "Current disk usage percentage",
		},
		[]string{"mount", "device"},
	)

	diskTotal := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "system_disk_total_bytes",
			Help: "Total disk space in bytes",
		},
		[]string{"mount", "device"},
	)

	collector := &MetricsCollector{
		cpuUsage:      cpuUsage,
		cpuUsagePerCore: cpuUsagePerCore,
		memUsage:      memUsage,
		memAvailable:  memAvailable,
		memTotal:      memTotal,
		diskUsage:     diskUsage,
		diskTotal:     diskTotal,
		scrapeInterval: scrapeInterval,
	}

	// Register all metrics
	prometheus.MustRegister(
		collector.cpuUsage,
		collector.cpuUsagePerCore,
		collector.memUsage,
		collector.memAvailable,
		collector.memTotal,
		collector.diskUsage,
		collector.diskTotal,
	)

	return collector
}

// Start begins collecting metrics in background
func (m *MetricsCollector) Start(ctx context.Context) {
	slog.Info("Starting metrics collection", 
		"interval", m.scrapeInterval)
	
	go func() {
		ticker := time.NewTicker(m.scrapeInterval)
		defer ticker.Stop()

		// Collect once immediately at startup
		m.collect()

		for {
			select {
			case <-ticker.C:
				m.collect()
			case <-ctx.Done():
				slog.Info("Stopping metrics collection")
				return
			}
		}
	}()
}

// collect gathers all system metrics
func (m *MetricsCollector) collect() {
	m.collectCPU()
	m.collectMemory()
	m.collectDisk()
}

// collectCPU gathers CPU metrics
func (m *MetricsCollector) collectCPU() {
	// Overall CPU usage
	percent, err := cpu.Percent(0, false)
	if err != nil {
		slog.Error("Failed to get CPU usage", "error", err)
		return
	}
	
	if len(percent) > 0 {
		m.cpuUsage.Set(percent[0])
	}

	// Per-core CPU usage
	perCorePercent, err := cpu.Percent(0, true)
	if err != nil {
		slog.Error("Failed to get per-core CPU usage", "error", err)
		return
	}

	for i, p := range perCorePercent {
		m.cpuUsagePerCore.WithLabelValues(fmt.Sprintf("%d", i)).Set(p)
	}
}

// collectMemory gathers memory metrics
func (m *MetricsCollector) collectMemory() {
	v, err := mem.VirtualMemory()
	if err != nil {
		slog.Error("Failed to get memory usage", "error", err)
		return
	}

	m.memUsage.Set(v.UsedPercent)
	m.memAvailable.Set(float64(v.Available))
	m.memTotal.Set(float64(v.Total))
}

// collectDisk gathers disk metrics
func (m *MetricsCollector) collectDisk() {
	partitions, err := disk.Partitions(false)
	if err != nil {
		slog.Error("Failed to get disk partitions", "error", err)
		return
	}

	for _, partition := range partitions {
		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			slog.Error("Failed to get disk usage", 
				"mountpoint", partition.Mountpoint, 
				"error", err)
			continue
		}

		m.diskUsage.WithLabelValues(
			partition.Mountpoint, 
			partition.Device,
		).Set(usage.UsedPercent)
		
		m.diskTotal.WithLabelValues(
			partition.Mountpoint, 
			partition.Device,
		).Set(float64(usage.Total))
	}
}

// parseFlags parses command line flags and returns config
func parseFlags() *Config {
	config := &Config{}

	flag.IntVar(&config.Port, "port", 8080, "Port to serve metrics on")
	flag.StringVar(&config.MetricsPath, "metrics-path", "/metrics", "Path to expose metrics on")
	flag.DurationVar(&config.ScrapeInterval, "scrape-interval", 5*time.Second, "Interval between metric collections")
	flag.StringVar(&config.LogLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	flag.Parse()

	return config
}

// setupLogger configures structured logging
func setupLogger(logLevel string) {
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

func main() {
	// Parse command line arguments
	config := parseFlags()
	
	// Set up logging
	setupLogger(config.LogLevel)
	
	// Create context that listens for termination signals
	ctx, stop := signal.NotifyContext(context.Background(), 
		syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	
	// Create and start metrics collector
	collector := NewMetricsCollector(config.ScrapeInterval)
	collector.Start(ctx)
	
	// Set up HTTP server
	mux := http.NewServeMux()
	mux.Handle(config.MetricsPath, promhttp.Handler())
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(
			w,
			`<html>
				<head><title>System Metrics Exporter</title></head>
				<body>
					<h1>System Metrics Exporter</h1>
					<p><a href="%s">Metrics</a></p>
				</body>
			</html>`, config.MetricsPath)
	})
	
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: mux,
	}
	
	// Start server in a goroutine
	go func() {
		slog.Info("Starting metrics server", 
			"address", server.Addr, 
			"metrics_path", config.MetricsPath)
		
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