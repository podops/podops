package metadata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testAudioFilePath = "../../example/clip1.mp3"
	testImageFilePath = "../../example/cover.png"
)

func TestExtractMetadataFromAudioFile(t *testing.T) {

	meta, err := ExtractMetadataFromFile(testAudioFilePath)
	assert.NoError(t, err)
	assert.NotNil(t, meta)

	assert.Equal(t, meta.ContentType, "audio/mpeg")
	assert.Greater(t, meta.Duration, int64(0))
	assert.NotEmpty(t, meta.ETag, meta.Name)
	assert.Greater(t, meta.Size, int64(0))
	assert.Greater(t, meta.Timestamp, int64(0))

	assert.True(t, meta.IsAudio())
}

func TestExtractMetadataFromImageFile(t *testing.T) {

	meta, err := ExtractMetadataFromFile(testImageFilePath)
	assert.NoError(t, err)
	assert.NotNil(t, meta)

	assert.False(t, meta.IsAudio())
	assert.Equal(t, meta.ContentType, "image/png")
}

func TestExtractMetadataFromFileNotFound(t *testing.T) {
	meta, err := ExtractMetadataFromFile("fileNotFound.mp3")
	assert.Error(t, err)
	assert.Nil(t, meta)
}
