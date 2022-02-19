package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestENVConfigValues(t *testing.T) {
	assert.NotEmpty(t, StorageLocation)
	assert.Equal(t, defaultStorageLocation, StorageLocation)

	assert.NotEmpty(t, BuildLocation)
	assert.Equal(t, defaultBuildLocation, BuildLocation)
}

func TestDefaultClientSettings(t *testing.T) {
	settings1 := DefaultClientSettings()

	assert.NotNil(t, settings1)
	assert.NoError(t, settings1.Validate())

	assert.Equal(t, defaultAPIEndpoint, settings1.Endpoint)
	assert.Equal(t, defaultAPIEndpoint, settings1.DefaultEndpoint)
	assert.Equal(t, UserAgentString, settings1.UserAgent)
	assert.True(t, settings1.HasOption(PodopsServiceEndpointEnv))
	assert.Equal(t, defaultServiceEndpoint, settings1.GetOption(PodopsServiceEndpointEnv))
	assert.True(t, settings1.HasOption(PodopsContentEndpointEnv))
	assert.Equal(t, defaultContentEndpoint, settings1.GetOption(PodopsContentEndpointEnv))
}

/*
func TestWriteReadClientSettings(t *testing.T) {
	settings1 := DefaultClientSettings()

	assert.NotNil(t, settings1)
	err := settings1.WriteToFile(ResolveConfigPath("."))
	assert.NoError(t, err)

	settings2, err := settings.ReadSettingsFromFile(ResolveConfigPath("."))
	assert.NoError(t, err)
	assert.NotNil(t, settings2)
	assert.Equal(t, settings1, settings2)
}
*/
