package collector

import (
	"fmt"
	"log/slog"

	"github.com/shirou/gopsutil/v3/cpu"
)

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

