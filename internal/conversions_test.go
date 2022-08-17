package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDuration(t *testing.T) {
	assert.Equal(t, "00:00:00", ParseDuration(-1))
	assert.Equal(t, "00:00:00", ParseDuration(0))
	assert.Equal(t, "00:00:01", ParseDuration(1))
	assert.Equal(t, "00:00:10", ParseDuration(10))
	assert.Equal(t, "00:01:00", ParseDuration(60))
	assert.Equal(t, "00:01:01", ParseDuration(61))
	assert.Equal(t, "00:11:01", ParseDuration(661))
	assert.Equal(t, "01:00:00", ParseDuration(3600))
	assert.Equal(t, "01:00:01", ParseDuration(3601))
	assert.Equal(t, "01:01:01", ParseDuration(3661))
	assert.Equal(t, "11:01:01", ParseDuration(39661))
}

func TestConvStrToInt(t *testing.T) {
	assert.Equal(t, 0, ConvStrToInt(""))
	assert.Equal(t, 0, ConvStrToInt("0"))
	assert.Equal(t, 42, ConvStrToInt("42"))
	assert.Equal(t, -1, ConvStrToInt("42.0"))
	assert.Equal(t, -1, ConvStrToInt("abc"))
}

func TestConvTimeStringToSeconds(t *testing.T) {
	// inverse from TestParseDuration
	assert.Equal(t, 0, ConvTimeStringToSeconds("00:00:00"))
	assert.Equal(t, 1, ConvTimeStringToSeconds("00:00:01"))
	assert.Equal(t, 10, ConvTimeStringToSeconds("00:00:10"))
	assert.Equal(t, 60, ConvTimeStringToSeconds("00:00:60"))
	assert.Equal(t, 61, ConvTimeStringToSeconds("00:01:01"))
	assert.Equal(t, 661, ConvTimeStringToSeconds("00:11:01"))
	assert.Equal(t, 3600, ConvTimeStringToSeconds("01:00:00"))
	assert.Equal(t, 3601, ConvTimeStringToSeconds("01:00:01"))
	assert.Equal(t, 3661, ConvTimeStringToSeconds("01:01:01"))
	assert.Equal(t, 39661, ConvTimeStringToSeconds("11:01:01"))
	// edge cases
	assert.Equal(t, 42, ConvTimeStringToSeconds("42"))
	assert.Equal(t, 0, ConvTimeStringToSeconds(""))
	assert.Equal(t, -1, ConvTimeStringToSeconds("abc"))
	assert.Equal(t, -1, ConvTimeStringToSeconds("42:11:01:01"))
}
