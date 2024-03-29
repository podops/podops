package builder

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	rootDir = "../../example/"
)

func TestBuildOnly(t *testing.T) {

	name, err := Build(context.TODO(), rootDir, false, false, true, false, true)
	if err != nil {
		fmt.Println(err)
	}

	assert.NoError(t, err)
	assert.NotEmpty(t, name)
}

func TestAssembleAfterBuildOnly(t *testing.T) {

	err := Assemble(context.TODO(), rootDir, false, false, false)
	if err != nil {
		fmt.Println(err)
	}

	assert.NoError(t, err)
}

func TestBuildAndAssemble(t *testing.T) {

	name, err := Build(context.TODO(), rootDir, false, false, false, false, true)
	if err != nil {
		fmt.Println(err)
	}

	assert.NoError(t, err)
	assert.NotEmpty(t, name)
}

/*
func _TestAssemble(t *testing.T) {

	err := Assemble(context.TODO(), rootDir, false)
	if err != nil {
		fmt.Println(err)
	}

	assert.NoError(t, err)
}
*/
