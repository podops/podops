package config

import (
	"os"
	"os/user"
	"path/filepath"

	"github.com/txsvc/stdlib/v2/env"
)

const (
	defaultBuildLocation   = ".build"
	defaultStorageLocation = "/data/storage"
	defaultStaticLocation  = "/data/public/default"

	DefaultTouchFile             = ".updated"
	DefaultConfigFileLocation    = ".podops/config"
	DefaultMasterKeyFileLocation = "master.key"
)

var (
	// StorageLocation is the root location for the cdn on the server
	StorageLocation = env.GetString(PodopsStorageLocationEnv, defaultStorageLocation)
	// StaticLocation is the root location for the static content on the server
	StaticLocation = env.GetString(PodopsStaticLocationEnv, defaultStaticLocation)
	// BuildLocation is the default location for podcast assets
	BuildLocation = env.GetString(PodopsBuildLocationEnv, defaultBuildLocation)
)

func DefaultConfigPath() string {
	if usr, err := user.Current(); err == nil {
		return filepath.Join(usr.HomeDir, DefaultConfigFileLocation)
	}

	return filepath.Join(".", DefaultConfigFileLocation)
}

func ResolveConfigPath(path string) []string {
	count := 2 // at least 2 locations will be checked

	if path != "" {
		count++
	}
	if env.Exists(PodopsConfigPathEnv) {
		count++
	}

	paths := make([]string, count)
	n := 0

	// start with path if one was provided
	if path != "" {
		paths[n] = path
		n++
	}

	// ENV is second if set
	if env.Exists(PodopsConfigPathEnv) {
		paths[n] = env.GetString(PodopsConfigPathEnv, "")
		n++
	}

	// current working dir is next
	if wd, err := os.Getwd(); err == nil {
		paths[n] = filepath.Join(wd, DefaultConfigFileLocation)
	} else {
		paths[n] = DefaultConfigFileLocation
	}
	n++

	// last is the global location
	paths[n] = DefaultConfigPath()

	return paths

}
