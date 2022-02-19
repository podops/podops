package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConstants(t *testing.T) {
	assert.NotEmpty(t, VersionString)
	assert.NotEmpty(t, UserAgentString)
	assert.NotEmpty(t, ServerString)
}
