package data

import (
	"slices"
	"strings"
)

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

func removePinFromSlice(s []string, pin string) []string {
	i := slices.Index(s, pin)
	return append(s[:i], s[i+1:]...)
}
