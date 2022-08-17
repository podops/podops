package cdn

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/mmcdole/gofeed"

	"github.com/txsvc/stdlib/v2/settings"
	"github.com/txsvc/stdlib/v2/timestamp"

	"github.com/podops/podops/auth"
	"github.com/podops/podops/config"
	"github.com/podops/podops/internal/metadata"
)

var (
	feedPathMapping map[string]string // e.g. /minimalpodcast/feed or /minimalpodcast/feed.xml -> /a7c94297acfc/feed.xml
	nameMapping     map[string]string // e.g. minimalpodcast -> a7c94297acfc

	mu sync.Mutex // used to protect the above maps
)

// Rewrite takes the canonical name of a podcast feed and returns its path in the cdn.
// If no mapping exists, the function just returns an empty string/false as status code.
func Rewrite(path string) (string, bool) {
	mu.Lock()
	defer mu.Unlock()

	if rw, ok := feedPathMapping[path]; ok {
		return rw, true
	}
	return "", false
}

// Resolve name returns the matching cdn GUID for a podcast's canonical name.
// If no mapping exists, the function just returns an empty string/false as status code.
func ResolveName(name string) (string, bool) {
	mu.Lock()
	defer mu.Unlock()

	if guid, ok := nameMapping[name]; ok {
		return guid, true
	}
	return "", false
}

// CreateInventoryMappings scans a file location for podcast feeds and creates
// the canonical names mappings to their file location on the cdn.
func CreateInventoryMappings(ctx context.Context, root string) error {

	mu.Lock()
	defer mu.Unlock()

	// re-initialize the maps i.e. forget old stuff
	feedPathMapping = make(map[string]string)
	nameMapping = make(map[string]string)

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {

		name := filepath.Base(path)
		if name == config.DefaultFeedName {
			parts := strings.Split(path, "/")
			parent := parts[len(parts)-2]

			// parse feed.xml and extract the name & uri
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			fp := gofeed.NewParser()
			feed, err := fp.Parse(file)
			if err != nil {
				return err
			}

			link, err := url.Parse(feed.Link)
			if err != nil {
				return err
			}

			// FIXME what if the mapping already exists? should not happen but ... ?

			// map path/feed and path/feed.xml to the location in the cdn storage
			feedPathMapping[link.Path+"/feed"] = path
			feedPathMapping[link.Path+"/feed.xml"] = path

			// map podcast name to GUID
			nameMapping[link.Path[1:]] = parent
		}

		return nil
	})

	return err
}

// CreateCredentialsMappings scans a file location for repo specific client credentials.
func CreateCredentialsMappings(ctx context.Context, root string) error {

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {

		name := filepath.Base(path)
		if name == config.DefaultMasterKeyFileLocation {
			if cfg, err := settings.ReadSettingsFromFile(path); err == nil {
				auth.RegisterAuthorization(cfg)
			}
		}

		return nil
	})

	return err
}

// ListResources collects all the media resources in the specified location
// and returns the metadata for these resources.
func ListResources(ctx context.Context, root string) (*[]metadata.Metadata, error) {

	var rsrc []metadata.Metadata
	rsrc = make([]metadata.Metadata, 0)

	err := filepath.Walk(root, func(path string, info os.FileInfo, e error) error {
		fi, err := os.Stat(path)
		if err != nil {
			return err
		}
		if fi.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)

		if ext == ".xml" || ext == ".yaml" || ext == ".key" {
			return nil // skip e.g feed.xml
		}

		m, err := metadata.ExtractMetadataFromFile(path)
		if err != nil {
			return err
		}

		parts := strings.Split(m.Name, ".")
		if len(parts) == 2 {
			// relace the etag as it does not reflect the etag on the client / producer side
			m.ETag = parts[0]
		} else {
			m.ETag = ""
		}
		rsrc = append(rsrc, *m)

		return nil
	})

	if err != nil {
		return nil, err
	}
	return &rsrc, nil
}

func MarkStorageChanged() {
	path := filepath.Join(config.StorageLocation, config.DefaultTouchFile)
	data := []byte(fmt.Sprintf("%d", timestamp.Now()))
	os.WriteFile(path, data, 0644)
}

func IsStorageChanged(ts int64) bool {
	path := filepath.Join(config.StorageLocation, config.DefaultTouchFile)
	body, err := ioutil.ReadFile(path)
	if err != nil {
		return true // better save than sorry, i.e. reload
	}
	i, err := strconv.ParseInt(string(body), 10, 0)
	if err != nil {
		return true // better save than sorry, i.e. reload
	}
	return ts < i
}
