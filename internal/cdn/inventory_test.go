package cdn

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/podops/podops/config"
	"github.com/stretchr/testify/assert"
)

const (
	rootDir      = "../../data/cdn"
	inventoryDir = "../../example"
)

func TestCreateInventoryMappings(t *testing.T) {

	err := CreateInventoryMappings(context.TODO(), rootDir)
	if err != nil {
		fmt.Println(err)
	}

	assert.NoError(t, err)
}

func TestInventory(t *testing.T) {

	root := filepath.Join(inventoryDir, config.BuildLocation)
	rsrc, err := ListResources(context.TODO(), root)
	if err != nil {
		fmt.Println(err)
	}

	assert.NoError(t, err)
	assert.NotNil(t, rsrc)
	assert.NotEmpty(t, rsrc)
}
