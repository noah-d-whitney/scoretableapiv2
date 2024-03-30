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
		return fmt.Sprintf("%.2f%%", value*100)
	}
}
