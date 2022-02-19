package loader

import (
	"context"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/podops/podops"
)

const (
	filePerm fs.FileMode = 0644
	dirPerm  fs.FileMode = 0644
)

type (
	// ResourceLoaderFunc implements loading of resources
	ResourceLoaderFunc func(data []byte) (interface{}, string, error)
)

var (
	resourceLoaders map[string]ResourceLoaderFunc
)

func init() {
	resourceLoaders = make(map[string]ResourceLoaderFunc)
	resourceLoaders[podops.ResourceShow] = loadShowResource
	resourceLoaders[podops.ResourceEpisode] = loadEpisodeResource
	//resourceLoaders[podops.ResourceAsset] = loadAssetResource
}

func WriteResource(ctx context.Context, path string, rsrc interface{}) error {
	data, err := yaml.Marshal(rsrc)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, data, filePerm)
}

func ReadResource(ctx context.Context, path string) (interface{}, string, string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, "", "", err
	}

	r, kind, guid, err := UnmarshalResource(data)
	if err != nil {
		return nil, "", "", err
	}

	return r, kind, guid, nil
}

func ReadEnclosure(ctx context.Context, path string) (*podops.AssetRef, error) {
	var r podops.AssetRef

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal([]byte(data), &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func DeleteResource(ctx context.Context, path string) error {
	return os.Remove(path)
}

// UnmarshalResource takes a byte array and determines its kind before unmarshalling it into its struct form
func UnmarshalResource(data []byte) (interface{}, string, string, error) {

	r, _ := LoadGenericResource(data)
	loader := resourceLoaders[r.Kind]
	if loader == nil {
		return nil, "", "", fmt.Errorf(podops.MsgResourceIsInvalid, r.Kind)
	}

	resource, guid, err := loader(data)
	if err != nil {
		return nil, "", "", err
	}
	return resource, r.Kind, guid, nil
}

// LoadGenericResource reads only the metadata of a resource
func LoadGenericResource(data []byte) (*podops.GenericResource, error) {
	var r podops.GenericResource

	err := yaml.Unmarshal([]byte(data), &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func loadShowResource(data []byte) (interface{}, string, error) {
	var show podops.Show

	err := yaml.Unmarshal([]byte(data), &show)
	if err != nil {
		return nil, "", err
	}

	return &show, show.GUID(), nil
}

func loadEpisodeResource(data []byte) (interface{}, string, error) {
	var episode podops.Episode

	err := yaml.Unmarshal([]byte(data), &episode)
	if err != nil {
		return nil, "", err
	}

	return &episode, episode.GUID(), nil
}

/*
func loadAssetResource(data []byte) (interface{}, string, error) {
	var asset podops.Asset

	err := yaml.Unmarshal([]byte(data), &asset)
	if err != nil {
		return nil, "", err
	}

	return &asset, asset.Metadata.GUID, nil
}
*/
