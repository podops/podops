package client

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/txsvc/stdlib/v2/settings"

	"github.com/podops/podops"
	"github.com/podops/podops/config"
	"github.com/podops/podops/feed"
	"github.com/podops/podops/internal/api"
	"github.com/podops/podops/internal/metadata"
)

const (
	initRoute  = "/init"
	assetRoute = "/asset"
)

func Init(userid, parent string) (*settings.DialSettings, error) {
	if userid == "" {
		return nil, podops.ErrInvalidParameters
	}
	if parent == "" {
		return nil, podops.ErrInvalidGUID
	}

	cfg := settings.DialSettings{}
	cmd := fmt.Sprintf("%s%s/%s/%s", api.NamespacePrefix, initRoute, userid, parent)
	_, err := Put(config.Settings().Endpoint, cmd, config.Settings().Credentials.Token, nil, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Upload moves a file from the local file system to the CDN. The file is placed in
// a location specified by parent. The location has to exist beforehand otherwise the
// API endpoint will return an error.
func Upload(parent, path string) error {
	if parent == "" {
		return podops.ErrInvalidGUID
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return podops.ErrResourceNotFound
	}

	url := fmt.Sprintf("%s%s%s/%s", config.Settings().Endpoint, api.NamespacePrefix, assetRoute, parent)
	req, err := UploadRequest(url, config.Settings().Credentials.Token, api.UploadFormName, path)
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	body := &bytes.Buffer{}
	_, err = body.ReadFrom(resp.Body)
	if err != nil {
		return err
	}
	resp.Body.Close()

	if resp.StatusCode > http.StatusNoContent {
		return fmt.Errorf(podops.MsgResourceUploadError, fmt.Sprintf("%s:%s", path, resp.Status))
	}

	return nil
}

func Delete(parent, asset string) error {
	if parent == "" {
		return podops.ErrInvalidGUID
	}
	cmd := fmt.Sprintf("%s%s/%s/%s", api.NamespacePrefix, assetRoute, parent, asset)
	_, err := Del(config.Settings().Endpoint, cmd, config.Settings().Credentials.Token, nil)
	if err != nil {
		return err
	}

	return nil
}

// List returns a list of media resources on the remote content endpoint.
func List(parent string) (*[]metadata.Metadata, error) {
	var rsrc []metadata.Metadata

	if parent == "" {
		return nil, podops.ErrInvalidGUID
	}

	cmd := fmt.Sprintf("%s%s/%s", api.NamespacePrefix, assetRoute, parent)

	_, err := Get(config.Settings().Endpoint, cmd, config.Settings().Credentials.Token, &rsrc)
	if err != nil {
		return nil, err
	}

	return &rsrc, nil
}

// Sync synchronizes a local repository against the remote content endpoint. All local resources that
// are missing on the remote site will be uploaded, already existing ones are ignored.
// Only media files (e.g. .mp3, .png) are synchronized.
func Sync(parent, root string, purge bool) error {
	if parent == "" {
		return podops.ErrInvalidGUID
	}

	// check that the cache dir exists
	assetPath := filepath.Join(root, config.BuildLocation)
	if _, err := os.Stat(assetPath); os.IsNotExist(err) {
		return podops.ErrResourceNotFound
	}

	// get a list of resources and build a look-up table
	r, err := List(parent)
	if err != nil {
		return err
	}

	rsrc := make(map[string]metadata.Metadata)
	for _, m := range *r {
		rsrc[m.ETag] = m
	}

	// now iterate over all the resource definitions and upload missing ones
	err = filepath.Walk(assetPath, func(path string, info os.FileInfo, e error) error {
		fi, err := os.Stat(path)
		if err != nil {
			return err
		}
		if fi.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		if ext != ".yaml" {
			return nil // skip e.g feed.xml
		}

		ar, err := feed.LoadAssetRef(path)
		if err != nil {
			return err
		}

		_, ok := rsrc[ar.ETag]
		if !ok {
			parts := strings.Split(ar.URI, ".")
			if len(parts) < 2 {
				return podops.ErrInvalidResourceName
			}
			assetFileName := fmt.Sprintf("%s.%s", ar.ETag, parts[len(parts)-1])
			assetPath := filepath.Join(root, config.BuildLocation, assetFileName)

			return Upload(parent, assetPath)
		} else {
			// remove the processed asset from the map, for purgeing assets later
			delete(rsrc, ar.ETag)
		}

		return nil
	})

	if err != nil {
		return err
	}

	// upload feed.xml last
	feedFilePath := filepath.Join(root, config.BuildLocation, feed.DefaultFeedName)
	err = Upload(parent, feedFilePath)
	if err != nil {
		return err
	}

	// purge obsolete assets
	if purge {
		for _, r := range rsrc {
			if err := Delete(parent, r.Name); err != nil {
				return err
			}
		}
	}

	return nil
}
