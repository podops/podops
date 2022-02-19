package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfigPath(t *testing.T) {
	path := DefaultConfigPath()
	assert.NotEmpty(t, path)
	assert.True(t, strings.HasSuffix(path, DefaultConfigFileLocation))
}

func TestResolveConfigPath(t *testing.T) {
	paths := ResolveConfigPath("")
	assert.Equal(t, 2, len(paths))

	paths = ResolveConfigPath(".")
	assert.Equal(t, 3, len(paths))
}
