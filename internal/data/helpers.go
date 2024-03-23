package data

import (
	"strings"
	"time"
)

type DateRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

func parsePinList(pins []string) (assignList []string, unassignList []string) {
	for _, pin := range pins {
		if pin[0] != '-' {
			assignList = append(assignList, pin)
		} else if pin[0] == '-' {
			cleanStr := strings.TrimPrefix(pin, "-")
			unassignList = append(unassignList, cleanStr)
		}
	}

	return
}
