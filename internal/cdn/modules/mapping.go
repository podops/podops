package modules

import (
	"context"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"

	"github.com/txsvc/stdlib/v2/timestamp"

	"github.com/podops/podops/config"
	"github.com/podops/podops/internal/cdn"
)

// see https://github.com/caddyserver/cache-handler/blob/master/httpcache.go
// see https://github.com/pquerna/cachecontrol

const (
	reloadInterval = 60 // seconds
)

type (
	ContentMapper struct {
		logger *zap.Logger
	}
)

var (
	// Interface guards
	_ caddy.Validator             = (*ContentMapper)(nil)
	_ caddy.Provisioner           = (*ContentMapper)(nil)
	_ caddyhttp.MiddlewareHandler = (*ContentMapper)(nil)
	_ caddyfile.Unmarshaler       = (*ContentMapper)(nil)
)

func init() {
	caddy.RegisterModule(ContentMapper{})
	httpcaddyfile.RegisterHandlerDirective("cdn_mapping", parseContentMapperConfig)
}

func (cr ContentMapper) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {

	// anything else than a GET/HEAD is handled elsewhere
	switch r.Method {
	case http.MethodGet:
	case http.MethodHead:
	default:
		return next.ServeHTTP(w, r)
	}

	if path, ok := cdn.Rewrite(r.RequestURI); ok {
		parts := strings.Split(r.RequestURI[1:], "/")
		guid, _ := cdn.ResolveName(parts[0])

		return serveAndCache(guid, path, "cdn.feed.get", w, r, next)
	}

	return next.ServeHTTP(w, r)
}

func (ContentMapper) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.mapping",
		New: func() caddy.Module { return new(ContentMapper) },
	}
}

func (cm *ContentMapper) Provision(ctx caddy.Context) error {

	cm.logger = ctx.Logger(cm)

	// setup a goroutine that periodically scans the StorageLocation for changes
	ticker := time.NewTicker(reloadInterval * time.Second)
	quit := make(chan struct{})
	go func() {
		ts := cm.reload(0)
		for {
			select {
			case <-ticker.C:
				ts = cm.reload(ts)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	return cdn.CreateInventoryMappings(context.Background(), config.StorageLocation)
}

func (cm *ContentMapper) reload(ts int64) int64 {
	if cdn.IsStorageChanged(ts) {
		err := cdn.CreateInventoryMappings(context.TODO(), config.StorageLocation)
		if err != nil {
			cm.logger.Error("error reloading name mapping", zap.Error(err))
			return ts
		}

		cm.logger.Info("reloaded the name mapping")
		return timestamp.Now()
	}
	return ts
}

func (cm *ContentMapper) Validate() error {
	return nil
}

func (cm *ContentMapper) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	return nil
}

func parseContentMapperConfig(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var m ContentMapper
	err := m.UnmarshalCaddyfile(h.Dispenser)
	return m, err
}
