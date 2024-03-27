package data

import (
	json2 "encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

type PeriodLength time.Duration

func (pl *PeriodLength) UnmarshalJSON(b []byte) error {
	unquoted, err := strconv.Unquote(string(b))
	if err != nil {
		return &json2.UnmarshalTypeError{Field: "period_length"}
	}
	parts := strings.Split(unquoted, ":")
	if len(parts) != 2 {
		return &json2.UnmarshalTypeError{Field: "period_length"}
	}
	minutes, err := strconv.Atoi(parts[0])
	if err != nil {
		return &json2.UnmarshalTypeError{Field: "period_length"}
	}
	seconds, err := strconv.Atoi(parts[1])
	if err != nil {
		return &json2.UnmarshalTypeError{Field: "period_length"}
	}
	totalTime := (time.Duration(minutes) * time.Minute) + (time.Duration(seconds) * time.Second)

	*pl = PeriodLength(totalTime)
	return nil
}

func (pl *PeriodLength) MarshalJSON() ([]byte, error) {
	duration := time.Duration(*pl)
	mins := int(math.Floor(duration.Minutes()))
	minsDuration := time.Duration(mins) * time.Minute
	secs := int(math.Round((duration - minsDuration).Seconds()))
	var padMin string
	var padSec string
	switch {
	case mins < 10:
		padMin = "0"
	default:
		padMin = ""
	}
	switch {
	case secs < 10:
		padSec = "0"
	default:
		padSec = ""
	}
	json := fmt.Sprintf(`"%s%d:%s%d"`, padMin, mins, padSec, secs)
	return []byte(json), nil
}
func (pl *PeriodLength) Duration() time.Duration {
	return time.Duration(*pl)
}
