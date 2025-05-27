// Copyright (c) Diogo Correia
// SPDX-License-Identifier: MIT

package vcstoics

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"
)

// Decode decodes quoted-printable text
func Decode(input string) string {
	// Replace CRLF codes
	input = strings.ReplaceAll(input, "=0D=0A", "\\n")

	// Decode the quoted-printable content
	result, err := decodeQuotedPrintable(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding quoted-printable: %v\n", err)
		return input
	}

	return result
}

// Implementation of quoted-printable decoding
func decodeQuotedPrintable(input string) (string, error) {
	var buf bytes.Buffer

	for i := 0; i < len(input); i++ {
		if input[i] == '=' && i+2 < len(input) {
			// Check if it's a valid hex sequence
			if isHexChar(input[i+1]) && isHexChar(input[i+2]) {
				b, err := hex.DecodeString(input[i+1 : i+3])
				if err != nil {
					return "", err
				}
				buf.Write(b)
				i += 2 // Skip the two characters we just processed
			} else if input[i+1] == '\r' && i+2 < len(input) && input[i+2] == '\n' {
				// Soft line break, skip it
				i += 2
			} else {
				// Not a hex sequence or soft line break, treat as literal
				buf.WriteByte(input[i])
			}
		} else {
			buf.WriteByte(input[i])
		}
	}

	return buf.String(), nil
}

func isHexChar(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'A' && c <= 'F') || (c >= 'a' && c <= 'f')
}

// Read a field that may continue on the next line
func readPossibleMultiline(fieldContent string, reader *bufio.Reader) (string, error) {
	var sb strings.Builder
	sb.WriteString(fieldContent)

	for {
		// Mark the current position for possible reset
		reader.Reset(reader)

		// Peek at the next character
		c, err := reader.ReadByte()
		if err != nil {
			if err == io.EOF {
				// End of file, return what we have
				return sb.String(), nil
			}
			return "", err
		}

		// If it's a space, this is a continuation line
		if unicode.IsSpace(rune(c)) {
			// Read the rest of the line
			line, err := reader.ReadString('\n')
			if err != nil && err != io.EOF {
				return "", err
			}
			sb.WriteString(strings.TrimRight(line, "\r\n"))
		} else {
			// Not a continuation, unread the byte and return
			reader.UnreadByte()
			break
		}
	}

	return sb.String(), nil
}

// Read a quoted-printable field that may span multiple lines
func readEncryptedField(fieldContent string, reader *bufio.Reader) (string, error) {
	var sb strings.Builder
	sb.WriteString(fieldContent)

	// If the line ends with =, it continues on the next line
	for strings.HasSuffix(sb.String(), "=") {
		// Remove the trailing =
		str := sb.String()
		sb.Reset()
		sb.WriteString(str[:len(str)-1])

		// Read the next line
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return "", fmt.Errorf("unexpected EOF while reading multiline field")
			}
			return "", err
		}
		sb.WriteString(strings.TrimRight(line, "\r\n"))
	}

	// Decode the content
	return Decode(sb.String()), nil
}

func Convert(in io.Reader, out io.Writer, email string) error {
	writer := NewICSWriter(email, out)
	defer writer.Close()

	reader := bufio.NewReader(in)

	var (
		line string
		err  error
	)

	for {
		line, err = reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return fmt.Errorf("error reading file: %w", err)
		}

		// Trim trailing newlines
		line = strings.TrimRight(line, "\r\n")

		// Process the current line
		if strings.EqualFold(line, "END:VCALENDAR") {
			break
		} else if !strings.EqualFold(line, "BEGIN:VTODO") && !strings.EqualFold(line, "BEGIN:VEVENT") {
			// Skip headers and other non-event data
			if strings.EqualFold(line, "BEGIN:VCALENDAR") ||
				strings.HasPrefix(strings.ToUpper(line), "PRODID:") ||
				strings.HasPrefix(strings.ToUpper(line), "VERSION:") {
				// Known headers, skip silently
			} else if line != "" {
				fmt.Fprintf(os.Stderr, "Unknown header entry: %s\n", line)
			}
		} else {
			// Processing an event or todo
			isEvent := strings.EqualFold(line, "BEGIN:VEVENT")

			var summary, location, description, status, due, sequence string
			var dtstart, dtend, dtstamp, rrule, alarm string

			// Process the event or todo
			for {
				line, err = reader.ReadString('\n')
				if err != nil && err != io.EOF {
					return fmt.Errorf("error reading file: %w", err)
				}

				// Trim trailing newlines
				line = strings.TrimRight(line, "\r\n")

				// Check for end of event/todo
				if strings.EqualFold(line, "END:VEVENT") || strings.EqualFold(line, "END:VTODO") {
					break
				}

				// Process fields
				if strings.HasPrefix(strings.ToUpper(line), "SUMMARY:") {
					summary, err = readPossibleMultiline(line[8:], reader)
					if err != nil {
						return err
					}
				} else if strings.HasPrefix(strings.ToUpper(line), "SUMMARY;ENCODING=QUOTED-PRINTABLE") {
					encodedPart := line[strings.Index(line, ":")+1:]
					summary, err = readEncryptedField(encodedPart, reader)
					if err != nil {
						return err
					}
				} else if strings.HasPrefix(strings.ToUpper(line), "LOCATION:") {
					location, err = readPossibleMultiline(line[9:], reader)
					if err != nil {
						return err
					}
				} else if strings.HasPrefix(strings.ToUpper(line), "LOCATION;ENCODING=QUOTED-PRINTABLE") {
					encodedPart := line[strings.Index(line, ":")+1:]
					location, err = readEncryptedField(encodedPart, reader)
					if err != nil {
						return err
					}
				} else if strings.HasPrefix(strings.ToUpper(line), "DESCRIPTION:") {
					description, err = readPossibleMultiline(line[12:], reader)
					if err != nil {
						return err
					}
				} else if strings.HasPrefix(strings.ToUpper(line), "DESCRIPTION;ENCODING=QUOTED-PRINTABLE") {
					encodedPart := line[strings.Index(line, ":")+1:]
					description, err = readEncryptedField(encodedPart, reader)
					if err != nil {
						return err
					}
				} else if strings.HasPrefix(strings.ToUpper(line), "DTSTART:") {
					dtstart = line[8:]
				} else if strings.HasPrefix(strings.ToUpper(line), "DTEND:") {
					dtend = line[6:]
				} else if strings.HasPrefix(strings.ToUpper(line), "DUE:") {
					due = line[4:]
				} else if strings.HasPrefix(strings.ToUpper(line), "STATUS:") {
					status = line[7:]
				} else if strings.HasPrefix(strings.ToUpper(line), "SEQUENCE:") {
					sequence = line[9:]
				} else if strings.HasPrefix(strings.ToUpper(line), "RRULE:") {
					rrule = line[6:]
				} else if strings.HasPrefix(strings.ToUpper(line), "AALARM:") ||
					strings.HasPrefix(strings.ToUpper(line), "AALARM;TYPE=X-EPOCSOUND:") {
					colonIdx := strings.Index(line, ":")
					if colonIdx != -1 {
						semicolonIdx := strings.Index(line[colonIdx+1:], ";")
						if semicolonIdx != -1 {
							alarm = line[colonIdx+1 : colonIdx+1+semicolonIdx]
						}
					}
				} else if strings.HasPrefix(strings.ToUpper(line), "LAST-MODIFIED:") {
					dtstamp = line[14:]
				}
				// Skip other fields for now

				if err == io.EOF {
					break
				}
			}

			// Add the event to the ICS writer
			err = writer.AddEvent(isEvent, summary, description, location, dtstart, dtend, rrule, dtstamp, sequence, due, status, alarm)
			if err != nil {
				return fmt.Errorf("error adding event: %w", err)
			}
		}

		if err == io.EOF {
			break
		}
	}

	return nil
}
