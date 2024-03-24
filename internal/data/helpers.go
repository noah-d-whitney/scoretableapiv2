package data

import (
	"strings"
	"time"
)

type DateRange struct {
	AfterDate  time.Time `json:"after_date,omitempty"`
	BeforeDate time.Time `json:"before_date,omitempty"`
}

func (r DateRange) IsZero() bool {
	return r.AfterDate.IsZero() && r.BeforeDate.IsZero()
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
