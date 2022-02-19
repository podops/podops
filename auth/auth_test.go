package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	scopeProductionRead  = "production:read"
	scopeProductionWrite = "production:write"
	scopeProductionBuild = "production:build"
	scopeResourceRead    = "resource:read"
	scopeResourceWrite   = "resource:write"
)

func TestScope(t *testing.T) {
	assert.True(t, hasScope([]string{scopeProductionWrite, scopeProductionRead, scopeResourceRead}, scopeProductionRead))
	assert.False(t, hasScope([]string{scopeProductionWrite, scopeProductionRead, scopeResourceRead}, scopeResourceWrite))

	assert.True(t, hasScope([]string{scopeProductionWrite, scopeProductionRead, scopeResourceRead}, scopeProductionRead+","+scopeProductionWrite))
	assert.False(t, hasScope([]string{scopeProductionWrite, scopeProductionRead, scopeResourceRead}, scopeProductionRead+","+scopeResourceWrite))
}
