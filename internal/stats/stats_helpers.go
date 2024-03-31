package stats

import (
	"fmt"
	"math"
)

func float64ToPercent(value float64) string {
	switch {
	case value == 1:
		return "100%"
	case math.IsNaN(value):
		return "N/A"
	default:
		return fmt.Sprintf("%.1f%%", value*100)
	}
}

func assertAndCopyStatsToMap[T Stat](stats []Stat) map[string]T {
	asserted := make(map[string]T)
	for _, s := range stats {
		asserted[s.getName()] = s.(T)
	}
	return asserted
}
