package modules

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/podops/podops/config"
)

// see https://github.com/caddyserver/cache-handler/blob/master/httpcache.go
// see https://github.com/pquerna/cachecontrol

// see https://github.com/podops/podops.legacy/blob/43c7df67e38056b9a08b4da3f484ec8aecfb994a/internal/cdn/cdn.go

type (
	ContentStorage struct {
	}
)

var (
	// Interface guards
	_ caddy.Validator             = (*ContentStorage)(nil)
	_ caddy.Provisioner           = (*ContentStorage)(nil)
	_ caddyhttp.MiddlewareHandler = (*ContentStorage)(nil)
	_ caddyfile.Unmarshaler       = (*ContentStorage)(nil)
)

func init() {
	caddy.RegisterModule(ContentStorage{})
	httpcaddyfile.RegisterHandlerDirective("cdn_server", parseContentStorageConfig)
}

func (cs ContentStorage) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {

	// anything else than a GET/HEAD is handled elsewhere
	switch r.Method {
	case http.MethodGet:
	case http.MethodHead:
	default:
		return next.ServeHTTP(w, r)
	}

	uri := r.RequestURI // expected is e.g. /a7c94297acfc/86124f7f9cf.mp3
	parts := strings.Split(uri[1:], "/")
	if len(parts) == 2 {
		return serveAndCache(parts[0], filepath.Join(config.StorageLocation, uri), "cdn.content.get", w, r, next)
	}

	return next.ServeHTTP(w, r)
}

func (ContentStorage) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.storage",
		New: func() caddy.Module { return new(ContentStorage) },
	}
}

func (cs *ContentStorage) Provision(ctx caddy.Context) error {
	return nil
}

func (cs *ContentStorage) Validate() error {
	return nil
}

func (cs *ContentStorage) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	return nil
}

func parseContentStorageConfig(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var m ContentStorage
	err := m.UnmarshalCaddyfile(h.Dispenser)
	return m, err
}
