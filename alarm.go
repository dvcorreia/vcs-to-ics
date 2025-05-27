// Copyright (c) Diogo Correia
// SPDX-License-Identifier: MIT

package vcstoics

import (
	"fmt"
	"strings"
	"time"
)

// Alarm represents a calendar alarm
type Alarm struct {
	Difference int64 // Difference in seconds
}

// NewAlarm creates a new alarm with start time and alarm time
func NewAlarm(start, alarmTime time.Time) *Alarm {
	return &Alarm{
		Difference: int64(alarmTime.Sub(start).Seconds()),
	}
}

// parseDuration converts the time difference to ICS duration format
func (a *Alarm) parseDuration() string {
	// Constants for duration parts
	const (
		minLength  = 60
		hourLength = minLength * 60
		dayLength  = hourLength * 24
		weekLength = dayLength * 7
	)

	var sb strings.Builder
	seconds := a.Difference

	if seconds > 0 {
		sb.WriteString("P")
	} else {
		sb.WriteString("-P")
		seconds = -seconds
	}

	if seconds%weekLength == 0 {
		sb.WriteString(fmt.Sprintf("%dW", seconds/weekLength))
		return sb.String()
	}

	days := seconds / dayLength
	seconds %= dayLength
	hours := seconds / hourLength
	seconds %= hourLength
	minutes := seconds / minLength
	seconds %= minLength

	if days > 0 {
		sb.WriteString(fmt.Sprintf("%dD", days))
	}

	if hours > 0 || minutes > 0 || seconds > 0 {
		sb.WriteString("T")
		if hours > 0 {
			sb.WriteString(fmt.Sprintf("%dH", hours))
		}
		if minutes > 0 {
			sb.WriteString(fmt.Sprintf("%dM", minutes))
		}
		if seconds > 0 {
			sb.WriteString(fmt.Sprintf("%dS", seconds))
		}
	}

	return sb.String()
}

// ToICS converts an Alarm to ICS format string
func (a *Alarm) ToICS(description string) string {
	var sb strings.Builder
	sb.WriteString("BEGIN:VALARM" + newLine)
	sb.WriteString("ACTION:DISPLAY" + newLine)
	sb.WriteString("DESCRIPTION:" + description + newLine)
	sb.WriteString("TRIGGER:" + a.parseDuration() + newLine)
	sb.WriteString("END:VALARM")
	return sb.String()
}
