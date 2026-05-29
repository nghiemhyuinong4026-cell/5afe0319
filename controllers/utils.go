package controllers

import (
	"errors"
	"time"
)

func parseTime(timeStr string) (time.Time, error) {
	// Try parsing with multiple formats
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	for _, format := range formats {
		t, err := time.Parse(format, timeStr)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, errors.New("unsupported time format")
}
