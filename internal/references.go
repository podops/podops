package internal

import (
	"crypto/rand"
	"fmt"
	"io"

	"github.com/txsvc/stdlib/v2/id"
)

// All functions related to the creation of ID, GUIDs or other media/asset references
// are collected in this file in order to have one place where all the
// relevant implementations can be found.

// CreateRandomAssetGUID returns a random ID for assets references.
// The GUID is 6 bytes / 12 char long, covering 16^12 address space.
// The GUID is most likely unique but there is not guarantee.
func CreateRandomAssetGUID() string {
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
func CreateETag(name string, size, timestamp int64) string {
	return id.Fingerprint(fmt.Sprintf("%s%d%d", name, size, timestamp))
}

func CreateSimpleID() string {
	id, _ := id.ShortUUID()
	return id
}

func CreateSimpleToken() string {
	token, _ := id.UUID()
	return token
}
