package collector

import (
	"context"
	"log/slog"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

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

