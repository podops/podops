package internal

import (
	"fmt"
	"time"
)

// ParseDuration converts seconds duration into HH:MM:SS format
func ParseDuration(duration int64) string {
	if duration <= 0 {
		return "00:00:00"
	}

	h := duration / 3600
	duration = duration % 3600

	m := duration / 60
	duration = duration % 60

	s := duration

	// HH:MM:SS
	if h > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	}

	// 00:MM:SS
	return fmt.Sprintf("00:%02d:%02d", m, s)
}

// ParseDateRFC1123Z returns a RFC1123Z formatted string
func ParseDateRFC1123Z(t *time.Time) string {
	if t != nil && !t.IsZero() {
		return t.Format(time.RFC1123Z)
	}
	return time.Now().UTC().Format(time.RFC1123Z)
}
