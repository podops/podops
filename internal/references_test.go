package internal

import (
	"testing"

	"github.com/podops/podops"
	"github.com/stretchr/testify/assert"
)

func TestCreateRandomAssetGUID(t *testing.T) {
	guid := CreateRandomAssetGUID()

	assert.NotEmpty(t, guid)
	assert.Equal(t, 12, len(guid))
}

func TestValidateRandomAssetGUID(t *testing.T) {
	guid := CreateRandomAssetGUID()
	assert.True(t, podops.ValidGUID(guid))

	assert.False(t, podops.ValidGUID("c8d34e2"))              // too short
	assert.False(t, podops.ValidGUID("c8d34e24b2c8d34e24b2")) // too long
	assert.False(t, podops.ValidGUID("XYZ3_-24b230"))         // invalid characters
	assert.False(t, podops.ValidGUID("2ffc4 bda824"))         // no spaces allowd
}
