package builder

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/podops/podops"
	"github.com/podops/podops/config"
	"github.com/podops/podops/internal/loader"
	"github.com/podops/podops/internal/metadata"
)

// ResolveResource imports or moves a resource into the local build location
func ResolveResource(ctx context.Context, parent, root string, force bool, encl *podops.AssetRef) error {
	switch encl.Rel {
	case podops.ResourceTypeExternal:
		return nil // FIXME verify that the resource exits or do nothing ?

	case podops.ResourceTypeImport:
		target := filepath.Join(root, config.BuildLocation, encl.MediaReference())

		ti, err := os.Stat(target)
		if os.IsNotExist(err) || force {
			return ImportResource(ctx, parent, root, encl)
		}

		if ti.Size() != int64(encl.Size) {
			return ImportResource(ctx, parent, root, encl)
		}

	case podops.ResourceTypeLocal:
		src := filepath.Join(root, encl.URI)
		target := filepath.Join(root, config.BuildLocation, encl.MediaReference())

		ti, err := os.Stat(target)
		if os.IsNotExist(err) || force {
			return MoveResource(ctx, src, target)
		}

		if ti.Size() != int64(encl.Size) {
			return MoveResource(ctx, src, target)
		}

	}

	return nil
}

// ImportResource imports a resource from src and places it into the local build location
func ImportResource(ctx context.Context, parent, root string, encl *podops.AssetRef) error {
	resp, err := http.Get(encl.URI)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf(podops.MsgResourceImportError, encl.URI)
		return err
	}

	// update the inventory
	meta := metadata.ExtractMetadataFromHeader(resp.Header.Clone())

	encl.Size = int(meta.Size)
	encl.Type = meta.ContentType
	encl.ETag = meta.ETag

	path := filepath.Join(root, config.BuildLocation, encl.MediaReference())

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	// transfer using a buffer
	buffer := make([]byte, 65536)
	l, err := io.CopyBuffer(out, resp.Body, buffer)

	if err != nil {
		return err
	}
	encl.Size = int(l)

	out.Close()

	// calculate the length of an audio file, if it is an audio file
	if meta.IsAudio() {
		duration, _ := metadata.CalculateLength(path)
		encl.Duration = int(duration)
	}

	return nil
}

func MoveResource(ctx context.Context, src, target string) error {
	sf, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sf.Close()

	tf, err := os.Create(target)
	if err != nil {
		return err
	}
	defer tf.Close()

	_, err = io.Copy(tf, sf)
	return err
}

// ValidateResource validates the existens of a local or remote resource
func ValidateResource(ctx context.Context, parent, root string, encl *podops.AssetRef) error {
	if encl.Rel == podops.ResourceTypeLocal {
		return validateLocal(parent, root, encl)
	}
	return validateRemote(ctx, parent, root, encl)
}

func LoadAssetRef(path string) (*podops.AssetRef, error) {
	var asset podops.AssetRef

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal([]byte(data), &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}

// validateLocal validates that the referenced resource exists on the filesystem
// and extracts additional metadata from the file. The metadata is written to the cache.
func validateLocal(parent, root string, encl *podops.AssetRef) error {
	path := filepath.Join(root, encl.URI)

	meta, err := metadata.ExtractMetadataFromFile(path)
	if err != nil {
		return err
	}

	encl.Size = int(meta.Size)
	encl.Type = meta.ContentType
	encl.ETag = meta.ETag
	encl.Timestamp = meta.Timestamp
	if meta.IsAudio() {
		encl.Duration = int(meta.Duration)
	}

	// asset file
	enclosurePath := filepath.Join(root, config.BuildLocation, fmt.Sprintf("%s.yaml", encl.AssetReference(parent)))
	return loader.WriteResource(context.TODO(), enclosurePath, encl)

}

// validateRemote validates that the referenced resource can be reached ("pinged")
// and extracts additional metadata from the http response. The metadata is written to the cache.
func validateRemote(ctx context.Context, parent, root string, encl *podops.AssetRef) error {
	head, err := pingURL(encl.URI)
	if err != nil {
		return err
	}

	meta := metadata.ExtractMetadataFromHeader(head)

	encl.Size = int(meta.Size)
	encl.Type = meta.ContentType
	encl.ETag = meta.ETag
	encl.Timestamp = meta.Timestamp
	// cant calculate the lenght of a remote asset just yet

	// asset file
	enclosurePath := filepath.Join(root, config.BuildLocation, fmt.Sprintf("%s.yaml", encl.AssetReference(parent)))
	return loader.WriteResource(context.TODO(), enclosurePath, encl)
}

// pingURL tries a HEAD or GET request to verify that 'url' exists and is reachable
func pingURL(url string) (http.Header, error) {

	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", config.UserAgentString)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp != nil {
		defer resp.Body.Close()
		// anything other than OK, Created, Accepted, NoContent is treated as an error
		if resp.StatusCode > http.StatusNoContent {
			return nil, fmt.Errorf(podops.MsgResourceIsInvalid, url)
		}
	}
	return resp.Header.Clone(), nil
}
