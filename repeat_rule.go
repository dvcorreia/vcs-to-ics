package vcstoics

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Frequency represents repeat rule frequency
type Frequency int

// Frequency constants
const (
	Daily Frequency = iota
	Weekly
	Monthly
	Yearly
)

func (f Frequency) String() string {
	return [...]string{"DAILY", "WEEKLY", "MONTHLY", "YEARLY"}[f]
}

// ParseFrequency converts a string to Frequency type
func ParseFrequency(s string) (Frequency, error) {
	switch strings.ToUpper(s) {
	case "DAILY", "D":
		return Daily, nil
	case "WEEKLY", "W":
		return Weekly, nil
	case "MONTHLY", "MD":
		return Monthly, nil
	case "YEARLY", "YM":
		return Yearly, nil
	default:
		return 0, fmt.Errorf("unknown frequency: %s", s)
	}
}

// RepeatRule represents a recurrence rule for calendar events
type RepeatRule struct {
	Frequency  Frequency
	Interval   int
	Until      time.Time
	Occurences int
}

// ToICS converts a RepeatRule to ICS format string
func (r *RepeatRule) ToICS() string {
	var sb strings.Builder
	sb.WriteString("RRULE:FREQ=" + r.Frequency.String())

	if r.Interval > 0 {
		sb.WriteString(";INTERVAL=" + strconv.Itoa(r.Interval))
	}

	if r.Occurences > 0 {
		sb.WriteString(";COUNT=" + strconv.Itoa(r.Occurences))
	} else if !r.Until.IsZero() {
		sb.WriteString(";UNTIL=" + FormatDate(r.Until))
	}

	return sb.String()
}

// ParseRepeatRule parses a VCS format repeat rule
func ParseRepeatRule(rrule string, useEndDate bool) (*RepeatRule, error) {
	parts := strings.Split(rrule, " ")
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty rule")
	}

	part0 := parts[0]

	// Check for end date or occurrences
	var occurDate time.Time
	occurCnt := 0

	occur := parts[len(parts)-1]
	if strings.HasPrefix(occur, "#") {
		var err error
		occurCnt, err = strconv.Atoi(occur[1:])
		if err != nil {
			return nil, fmt.Errorf("could not parse occurrences: %w", err)
		}
	} else if strings.Contains(occur, "T") && useEndDate {
		var err error
		occurDate, err = ParseDate(occur)
		if err != nil {
			return nil, fmt.Errorf("could not parse date: %w", err)
		}
	}

	// Parse frequency type
	var freq Frequency
	var intervalStr string

	switch {
	case strings.HasPrefix(part0, "D"):
		freq = Daily
		intervalStr = part0[1:]
	case strings.HasPrefix(part0, "W"):
		freq = Weekly
		intervalStr = part0[1:]
	case strings.HasPrefix(part0, "MD"):
		freq = Monthly
		intervalStr = part0[2:]
	case strings.HasPrefix(part0, "YM"):
		freq = Yearly
		intervalStr = part0[2:]
	default:
		return nil, fmt.Errorf("unknown frequency type in: %s", part0)
	}

	// Parse interval
	interval, err := strconv.Atoi(intervalStr)
	if err != nil {
		return nil, fmt.Errorf("could not parse interval: %w", err)
	}

	rr := RepeatRule{
		Frequency: freq,
		Interval:  interval,
	}

	// Create appropriate rule type
	if !occurDate.IsZero() {
		rr.Until = occurDate
	} else if occurCnt > 0 {
		rr.Occurences = occurCnt
	}

	return &rr, nil
}
