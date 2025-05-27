package vcstoics

import (
	"strings"
	"time"
)

// FormatTimeForDayEvent formats a time into YYYYMMDD format
func FormatTimeForDayEvent(t time.Time) string {
	return t.Format("20060102")
}

// FormatDate formats a time into ICS format (UTC)
func FormatDate(t time.Time) string {
	return t.UTC().Format("20060102T150405Z")
}

// ParseDate parses a date string in ICS format
func ParseDate(date string) (time.Time, error) {
	if strings.HasSuffix(date, "Z") {
		// UTC date format
		t, err := time.Parse("20060102T150405Z", date)
		if err != nil {
			return time.Time{}, err
		}
		return t, nil
	}

	// Local time format
	t, err := time.Parse("20060102T150405", date)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}
