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
