package internal

import (
	"fmt"
	"strconv"
	"strings"
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

// ConvStrToInt returns an int from a string and just "hides" all errors by returning -1
func ConvStrToInt(s string) int {
	if s == "" {
		return 0
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return -1
	}
	return i
}

// convTimeStringToSeconds decomposes a HH:MM:SS string and returns the seconds
func ConvTimeStringToSeconds(s string) int {
	hh, mm, ss := 0, 0, 0

	if s == "" {
		return 0
	}

	parts := strings.Split(s, ":")
	switch len(parts) {
	case 1:
		ss = ConvStrToInt(parts[0])
	case 2:
		mm = ConvStrToInt(parts[0])
		ss = ConvStrToInt(parts[1])
	case 3:
		hh = ConvStrToInt(parts[0])
		mm = ConvStrToInt(parts[1])
		ss = ConvStrToInt(parts[2])
	default:
		return -1 // should not happen !
	}
	return hh*3600 + mm*60 + ss
}
