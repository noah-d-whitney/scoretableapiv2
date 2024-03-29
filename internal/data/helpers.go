package data

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type DateRange struct {
	AfterDate  *time.Time `json:"after_date,omitempty"`
	BeforeDate *time.Time `json:"before_date,omitempty"`
}

func (r DateRange) IsEmpty() bool {
	return r.AfterDate == nil && r.BeforeDate == nil
}

func (r DateRange) IsFull() bool {
	return r.AfterDate != nil && r.BeforeDate != nil
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

type keysWithTypes struct {
	key     string
	keyType int
	dest    any
}

var ErrNoValueForKey = errors.New("no value found for key")
var ErrValueNotAsserted = errors.New("value could not be asserted to specified type")

func checkAndAssertStringFromMap(src map[string]any, key string) (string, error) {
	data, ok := src[key]
	if !ok {
		return "", ErrNoValueForKey
	}
	value, ok := data.(string)
	if !ok {
		return "", ErrValueNotAsserted
	}

	return value, nil
}

func checkAndAssertIntFromMap(src map[string]any, key string) (int, error) {
	data, ok := src[key]
	if !ok {
		return 0, ErrNoValueForKey
	}
	fmt.Printf("data value: %v\n", data)

	value, ok := data.(float64)
	if !ok {
		return 0, ErrValueNotAsserted
	}

	return int(value), nil

}

func checkAndAssertBoolFromMap(src map[string]any, key string) (bool, error) {
	data, ok := src[key]
	if !ok {
		return false, ErrNoValueForKey
	}
	fmt.Printf("data value: %v\n", data)

	value, ok := data.(bool)
	if !ok {
		return false, ErrValueNotAsserted
	}

	return value, nil

}
