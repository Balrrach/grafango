package collector

import (
	"log/slog"

	"github.com/shirou/gopsutil/v3/disk"
)

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

