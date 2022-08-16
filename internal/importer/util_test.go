package importer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringExpect(t *testing.T) {
	assert.Equal(t, "False", stringExpect("false", "False", "other"))
	assert.Equal(t, "other", stringExpect("abc", "False", "other"))
	assert.Equal(t, "other", stringExpect("", "False", "other"))
}
