package modules

import (
	"fmt"
	weakrand "math/rand"
	"net/http"
	"os"
	"strconv"

	"github.com/caddyserver/caddy/v2/modules/caddyhttp"

	"github.com/podops/podops/config"
	"github.com/podops/podops/internal/metadata"
)

const (
	cacheControl           = "public, max-age=1800"
	minBackoff, maxBackoff = 2, 5
)

// serveAndCache returns the requested resource. The implementation borrows from Caddy's own implementation:
// https://github.com/caddyserver/caddy/blob/master/modules/caddyhttp/fileserver/staticfiles.go
func serveAndCache(parent, path, metric string, w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {

	// only continue if the file exits
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return next.ServeHTTP(w, r)
	}

	// extract metadata from the file and the headers
	meta, err := metadata.ExtractMetadataFromFile(path)
	if err != nil {
		return err
	}
	//headers := api.ExtractHeaders(r)
	//offset, length := headers.Ranges()

	// metrics for analytics
	// FIXME observer.Meter(context.TODO(), metric, "production", parent, "user-agent", headers.UserAgent, "remote_addr", r.RemoteAddr, "type", meta.ContentType, "method", r.Method, "name", meta.Name, "size", fmt.Sprintf("%d", meta.Size), "offset", fmt.Sprintf("%d", offset), "length", fmt.Sprintf("%d", length))

	// write our own set of headers for the response
	w.Header().Add("etag", meta.ETag)
	w.Header().Add("accept-ranges", "bytes")
	w.Header().Add("cache-control", cacheControl)
	w.Header().Add("content-type", meta.ContentType)
	w.Header().Add("content-length", fmt.Sprintf("%d", meta.Size))
	w.Header().Add("x-served-by", config.ServerString)

	file, err := openFile(path, w)
	if err != nil {
		return err
	}
	defer file.Close()

	// let the implementation in the standard library deal with this ...
	http.ServeContent(w, r, info.Name(), info.ModTime(), file)

	return nil
}

func openFile(filename string, w http.ResponseWriter) (*os.File, error) {
	file, err := os.Open(filename)

	if err != nil {

		if os.IsNotExist(err) {
			return nil, caddyhttp.Error(http.StatusNotFound, err)
		} else if os.IsPermission(err) {
			return nil, caddyhttp.Error(http.StatusForbidden, err)
		}

		// maybe the server is under load and ran out of file descriptors?
		// have client wait arbitrary seconds to help prevent a stampede
		//nolint:gosec
		backoff := weakrand.Intn(maxBackoff-minBackoff) + minBackoff
		w.Header().Set("Retry-After", strconv.Itoa(backoff))

		return nil, caddyhttp.Error(http.StatusServiceUnavailable, err)
	}
	return file, nil
}
