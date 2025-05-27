package vcstoics

import (
	"fmt"
	"io"
	"strings"
	"time"
)

const (
	newLine = "\r\n" // ICS format requires CRLF
)

// ICSWriter handles writing calendar data in ICS format
type ICSWriter struct {
	Email         string
	writer        io.Writer
	contents      strings.Builder
	headerWritten bool
	closed        bool
}

func NewICSWriter(email string, writer io.Writer) *ICSWriter {
	return &ICSWriter{
		Email:  email,
		writer: writer,
	}
}

func (w *ICSWriter) Write(p []byte) (n int, err error) {
	if w.closed {
		return 0, fmt.Errorf("writer is closed")
	}

	// Ensure header is written before any content
	if !w.headerWritten {
		if err := w.writeHeader(); err != nil {
			return 0, err
		}
	}

	return w.writer.Write(p)
}

func (w *ICSWriter) writeHeader() error {
	header := "BEGIN:VCALENDAR" + newLine
	if w.Email != "" {
		header += "PRODID:" + w.Email + newLine
	} else {
		header += "PRODID:" + newLine
	}
	header += "VERSION:2.0" + newLine

	_, err := w.writer.Write([]byte(header))
	if err != nil {
		return err
	}

	w.headerWritten = true
	return nil
}
func (w *ICSWriter) AddEvent(isEvent bool, summary, description, location, dtStart, dtEnd, rrule, dtStamp, sequence, due, status, alarm string) error {
	if w.closed {
		return fmt.Errorf("writer is closed")
	}

	// Ensure header is written
	if !w.headerWritten {
		if err := w.writeHeader(); err != nil {
			return err
		}
	}

	// Build event content
	w.contents.Reset() // Clear the buffer for this event

	if isEvent {
		w.contents.WriteString("BEGIN:VEVENT" + newLine)
		if w.Email != "" {
			w.contents.WriteString("ORGANIZER:" + w.Email + newLine)
		}

		if summary != "" {
			w.contents.WriteString("SUMMARY:" + summary + newLine)
		}

		if description != "" {
			w.contents.WriteString("DESCRIPTION:" + description + newLine)
		}

		if location != "" {
			w.contents.WriteString("LOCATION:" + location + newLine)
		}

		if rrule != "" {
			repeatRule, err := ParseRepeatRule(rrule, false)
			if err == nil && repeatRule != nil {
				w.contents.WriteString(repeatRule.ToICS() + newLine)
			}
		}

		if dtStart != "" {
			w.contents.WriteString("DTSTART")
			if checkForSameStartAndEndTime(dtStart, dtEnd) {
				if checkForStartTimeIsZero(dtStart) {
					start, err := ParseDate(dtStart)
					if err == nil {
						w.contents.WriteString(";VALUE=DATE:" + FormatTimeForDayEvent(start) + newLine)
					} else {
						// Fallback if parsing fails
						w.contents.WriteString(":" + dtStart + newLine)
					}
				} else {
					w.contents.WriteString(":" + dtStart + newLine)
				}
			} else {
				w.contents.WriteString(":" + dtStart + newLine)
				if dtEnd != "" {
					w.contents.WriteString("DTEND:" + dtEnd + newLine)
				}
			}
		} else {
			return fmt.Errorf("no start date specified")
		}

		if dtStamp != "" {
			w.contents.WriteString("DTSTAMP:" + dtStamp + newLine)
		} else {
			// Get current UTC time if no dtstamp provided
			w.contents.WriteString("DTSTAMP:" + FormatDate(time.Now()) + newLine)
		}

		if alarm != "" {
			start, errStart := ParseDate(dtStart)
			alarmTime, errAlarm := ParseDate(alarm)

			if errStart == nil && errAlarm == nil {
				alarmObj := NewAlarm(start, alarmTime)
				w.contents.WriteString(alarmObj.ToICS(summary) + newLine)
			}
		}

		w.contents.WriteString("END:VEVENT" + newLine)
	} else {
		// Handle VTODO
		w.contents.WriteString("BEGIN:VTODO" + newLine)

		if dtStamp != "" {
			w.contents.WriteString("DTSTAMP:" + dtStamp + newLine)
		} else {
			w.contents.WriteString("DTSTAMP:" + FormatDate(time.Now()) + newLine)
		}

		if sequence != "" {
			w.contents.WriteString("SEQUENCE:" + sequence + newLine)
		} else {
			w.contents.WriteString("SEQUENCE:0" + newLine)
		}

		if w.Email != "" {
			w.contents.WriteString("ORGANIZER:" + w.Email + newLine)
		}

		if due != "" {
			w.contents.WriteString("DUE:" + due + newLine)
		}

		if status != "" {
			w.contents.WriteString("STATUS:" + status + newLine)
		}

		if summary != "" {
			w.contents.WriteString("SUMMARY:" + summary + newLine)
		}

		w.contents.WriteString("END:VTODO" + newLine)
	}

	// Write the event content to the underlying writer
	_, err := w.writer.Write([]byte(w.contents.String()))
	return err
}

// Close implements io.Closer interface and writes the calendar footer
func (w *ICSWriter) Close() error {
	if w.closed {
		return nil // Already closed
	}

	// Ensure header is written even if no events were added
	if !w.headerWritten {
		if err := w.writeHeader(); err != nil {
			return err
		}
	}

	// Write footer
	_, err := w.writer.Write([]byte("END:VCALENDAR"))
	w.closed = true

	// If the underlying writer implements io.Closer, close it too
	if closer, ok := w.writer.(io.Closer); ok {
		if closeErr := closer.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}

	return err
}

func checkForStartTimeIsZero(dtStart string) bool {
	if dtStart == "" {
		return false
	}

	dt, err := ParseDate(dtStart)
	if err != nil {
		return false
	}

	hour, min, _ := dt.Clock()
	return hour == 0 && min == 0
}

func checkForSameStartAndEndTime(dtStart, dtEnd string) bool {
	if dtStart == "" || dtEnd == "" {
		return false
	}
	return dtStart == dtEnd
}
