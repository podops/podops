package podops

import (
	"crypto/rand"
	"fmt"
	"io"
	"strings"

	"github.com/txsvc/stdlib/v2/id"
)

// All functions related to the creation of ID, GUIDs or other media/asset references
// are collected in this file in order to have one place where all the
// relevant implementations can be found.

// CreateRandomAssetGUID returns a random ID for assets references.
// The GUID is 6 bytes / 12 char long, covering 16^12 address space.
// The GUID is most likely unique but there is not guarantee.
func CreateRandomAssetGUID2() string {
	uuid := make([]byte, 6)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return ""
	}
	uuid[4] = uuid[4]&^0xc0 | 0x80
	uuid[2] = uuid[2]&^0xf0 | 0x40

	return fmt.Sprintf("%x", uuid[0:6])
}

// CreateETag calculates an etag based on a file's name, size and timestamp.
// It does not inspect the actual content of the file though.
func CreateETag2(name string, size, timestamp int64) string {
	return id.Fingerprint(fmt.Sprintf("%s%d%d", name, size, timestamp))
}

// AssetReference creates a unique asset reference based on the
// assets parent GUID and its URI. The reference is a CRC32 checksum
// and assumed to be static once the asset has been created.
// The media file the asset refers to might change over time.
func (r *AssetRef) AssetReference(parent string) string {
	return id.Checksum(parent + r.URI)
}

// MediaReference creates reference to a media file based on its current ETag.
// The MediaReference can change over time as the referenced file changes.
func (r *AssetRef) MediaReference() string {
	parts := strings.Split(r.URI, ".")
	if len(parts) == 0 {
		return r.ETag
	}
	return fmt.Sprintf("%s.%s", r.ETag, parts[len(parts)-1])
}

// CanonicalReference creates the full URI for the asset, as it can be found in the CDN
func (r *AssetRef) CanonicalReference(cdn, parent string) string {
	if r.Rel == ResourceTypeExternal {
		return r.URI
	}
	return fmt.Sprintf("%s/%s/%s", cdn, parent, r.MediaReference())
}

func CreateSimpleID2() string {
	id, _ := id.ShortUUID()
	return id
}

func CreateSimpleToken2() string {
	token, _ := id.UUID()
	return token
}
