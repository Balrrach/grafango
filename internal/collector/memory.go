package collector

import (
	"log/slog"

	"github.com/shirou/gopsutil/v3/mem"
)

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

