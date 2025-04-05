package main

import (
	"net/http"
	"log"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	cpuUsage = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "system_cpu_usage",
		Help: "Current CPU usage percentage",
	})
	memoryUsage = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "system_memory_usage",
		Help: "Current Memory usage percentage",
	})
)

func recordMetrics() {
	// Simulate metrics (replace with real system metrics)
	cpuUsage.Set(40.5)      // Example: 40.5% CPU
	memoryUsage.Set(70.3)   // Example: 70.3% Memory
}

func main() {
	prometheus.MustRegister(cpuUsage, memoryUsage)

	// Metrics collection loop
	go func() {
		for {
			recordMetrics()
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	log.Println("Metrics exporter running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

