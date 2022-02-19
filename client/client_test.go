package client

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/podops/podops/config"
	"github.com/stretchr/testify/assert"
)

const (
	guid        = "aaa94297acfc"
	invalidGuid = "aaa94297acfcXXX"

	rootDir                  = "../example"
	storageDir               = "../data/cdn"
	adminCredentialsLocation = "../cmd/cli/.podops/config"
	tmpCredentialLocation    = ".podops/config"

	testMP3FilePath = "../example/clip1.mp3"
	testPNGFilePath = "../example/cover.png"
	testPNGFile     = "cover.png"
)

func init() {
	err := os.RemoveAll(filepath.Join(storageDir, guid))
	if err != nil {
		log.Fatal(err)
	}
}

func TestInit(t *testing.T) {
	cfg := config.UpdateClientSettings(adminCredentialsLocation)

	assert.NotNil(t, cfg)
	assert.NotEmpty(t, config.Settings().Credentials.Token)
	assert.NotEmpty(t, config.Settings().Endpoint)
	assert.Equal(t, "http://localhost:8080", config.Settings().Endpoint)

	tmp, err := Init(cfg.Credentials.UserID, guid)

	assert.NoError(t, err)
	assert.NotNil(t, tmp)

	err = tmp.WriteToFile(tmpCredentialLocation)

	assert.NoError(t, err)
}

func TestUpload(t *testing.T) {
	config.UpdateClientSettings(tmpCredentialLocation)

	err := Upload(guid, testMP3FilePath)
	assert.NoError(t, err)

	err = Upload(guid, testPNGFilePath)
	assert.NoError(t, err)
}

func TestFailedUpload(t *testing.T) {
	config.UpdateClientSettings(tmpCredentialLocation)

	err := Upload(invalidGuid, testMP3FilePath)
	assert.Error(t, err)
}

func TestList(t *testing.T) {
	config.UpdateClientSettings(tmpCredentialLocation)

	metadata, err := List(guid)

	assert.NoError(t, err)
	assert.NotNil(t, metadata)
	assert.Equal(t, 2, len(*metadata))
}

func TestDelete(t *testing.T) {
	config.UpdateClientSettings(tmpCredentialLocation)

	err := Delete(guid, testPNGFile)
	assert.NoError(t, err)

	metadata, err := List(guid)
	assert.NoError(t, err)
	assert.NotNil(t, metadata)
	assert.Equal(t, 1, len(*metadata))
}

func TestSync(t *testing.T) {
	config.UpdateClientSettings(tmpCredentialLocation)

	err := Sync(guid, rootDir, false)
	assert.NoError(t, err)
}

func TestSyncWithPurge(t *testing.T) {
	config.UpdateClientSettings(tmpCredentialLocation)

	err := Sync(guid, rootDir, true)
	assert.NoError(t, err)

	metadata, err := List(guid)
	assert.NoError(t, err)
	assert.NotNil(t, metadata)
	assert.Equal(t, 5, len(*metadata))
}
