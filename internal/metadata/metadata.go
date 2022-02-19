package metadata

import (
	"io"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/podops/podops/internal"
	"github.com/tcolgate/mp3"
	"gopkg.in/yaml.v3"
)

const (
	defaultContentType = "application/octet-stream"
)

type (
	// Metadata keeps basic metadata of a cdn resource
	Metadata struct {
		Name        string `json:"name" yaml:"name"`
		Size        int64  `json:"size" yaml:"size"`
		Duration    int64  `json:"duration,omitempty" yaml:"duration,omitempty"`
		ContentType string `json:"type" yaml:"type"`
		Timestamp   int64  `json:"timestamp" yaml:"timestamp"`
		ETag        string `json:"etag" yaml:"etag"`
	}
)

// ExtractMetadataFromHeader extracts the metadata from http.Response
func ExtractMetadataFromHeader(header http.Header) *Metadata {
	if header == nil {
		return nil
	}

	meta := Metadata{
		ContentType: header.Get("content-type"),
		ETag:        cleanETag(header.Get("etag")),
	}
	l, err := strconv.ParseInt(header.Get("content-length"), 10, 64)
	if err == nil {
		meta.Size = l
	}
	// expects 'Wed, 30 Dec 2020 14:14:26 GM'
	t, err := time.Parse(time.RFC1123, header.Get("date"))
	if err == nil {
		meta.Timestamp = t.Unix()
	}
	return &meta
}

func ExtractMetadataFromFile(path string) (*Metadata, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		return nil, err
	}

	// the basics
	meta := CreateMetadata(fi)

	// try to detect the media type
	// thanks to https://gist.github.com/rayrutjes/db9b9ea8e02255d62ce2
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return nil, err
	}
	meta.ContentType = http.DetectContentType(buffer)
	// reset the read pointer
	file.Seek(0, 0)

	// in case it is a .mp3, calculate the play time.
	// thanks to https://stackoverflow.com/questions/60281655/how-to-find-the-length-of-mp3-file-in-golang
	if meta.IsAudio() {
		d := mp3.NewDecoder(file)

		var f mp3.Frame
		skipped := 0
		t := 0.0

		for {
			if err := d.Decode(&f, &skipped); err != nil {
				if err == io.EOF {
					break
				}
				return nil, err
			}
			t = t + f.Duration().Seconds()
		}
		meta.Duration = int64(t) // duration in seconds
	}
	return &meta, nil
}

func CreateMetadata(fi fs.FileInfo) Metadata {
	meta := Metadata{
		Name:        fi.Name(),
		Size:        fi.Size(),
		ContentType: defaultContentType,
		Timestamp:   fi.ModTime().Unix(),
	}
	meta.ETag = internal.CreateETag(meta.Name, meta.Size, meta.Timestamp)

	return meta
}

func LoadMetadataResource(path string) (*Metadata, error) {
	var meta Metadata

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal([]byte(data), &meta)
	if err != nil {
		return nil, err
	}

	return &meta, nil
}

func (m *Metadata) IsAudio() bool {
	return m.ContentType == "audio/mpeg" || m.ContentType == "application/octet-stream" // FIXME to include other types also
}

func (m *Metadata) IsImage() bool {
	return !m.IsAudio()
}

// CalculateLength returns the play duration of a media file like a .mp3
func CalculateLength(path string) (int64, error) {
	m, err := ExtractMetadataFromFile(path)
	if err != nil {
		return 0, err
	}
	return m.Duration, nil
}

// etag can start/end with "" that we need to strip
func cleanETag(etag string) string {
	et := strings.TrimSuffix(etag, "\"")
	return strings.TrimPrefix(et, "\"")
}
