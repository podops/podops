package api

import (
	"context"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/go-github/github"
	"github.com/labstack/echo/v4"

	"github.com/txsvc/httpservice/pkg/api"
	"github.com/txsvc/stdlib/v2/settings"

	"github.com/podops/podops"
	"github.com/podops/podops/auth"
	"github.com/podops/podops/config"
	"github.com/podops/podops/internal"
	"github.com/podops/podops/internal/cdn"
)

const (
	// NamespacePrefix namespace for the client and CLI
	NamespacePrefix = "/a/v1"

	// InitRoute route to InitEndpoint
	InitRoute = "/init/:userid/:parent"

	// AssetRoute route to asset related andpoints: AssetUploadEndpoint, AssetListEndpoint, AssetDeleteEndpoint
	AssetRoute       = "/asset/:parent"
	AssetDeleteRoute = "/asset/:parent/:asset"

	// WebhookRoute route to recieve call-back notifications
	WebhookRoute = "/static/:parent"

	UploadFormName = "asset"
)

func MeterAPIRequest(ctx context.Context, req *http.Request, parent, api string) {
	if parent == "" {
		parent = "unknown"
	}
	// metrics for analytics
	// FIXME observer.Meter(context.TODO(), api, "production", parent, "uri", req.RequestURI, "user-agent", req.UserAgent(), "remote_addr", req.RemoteAddr)
}

// InitEndpoints creates a new show namespace on the CDN
func InitEndpoint(c echo.Context) error {
	ctx := context.Background()

	// basic auth validation
	_, err := auth.CheckAuthorization(ctx, c, config.ScopeContentAdmin)
	if err != nil {
		return api.ErrorResponse(c, http.StatusUnauthorized, err)
	}

	userid := c.Param("userid")
	if userid == "" {
		return api.ErrorResponse(c, http.StatusBadRequest, podops.ErrInvalidRoute)
	}
	parent := c.Param("parent")
	if parent == "" {
		return api.ErrorResponse(c, http.StatusBadRequest, podops.ErrInvalidRoute)
	}

	MeterAPIRequest(ctx, c.Request(), parent, "api.show.init")

	// validate the non-existence of the target location first
	path := filepath.Join(config.StorageLocation, parent)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// create the location
		err = os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return api.ErrorResponse(c, http.StatusInternalServerError, podops.ErrInternalError)
		}
	}

	// create a new secret token
	secretPath := filepath.Join(path, config.DefaultMasterKeyFileLocation)
	cfg := settings.DialSettings{
		Endpoint:        config.Settings().Endpoint,
		DefaultEndpoint: config.Settings().DefaultEndpoint,
		Scopes:          []string{config.ScopeContentRead, config.ScopeContentWrite},
		Credentials: &settings.Credentials{
			ProjectID: parent,
			UserID:    userid,
			Token:     internal.CreateSimpleToken(),
			Expires:   0,
		},
	}
	config.WithServiceEndpoint(config.Settings().GetOption(config.PodopsServiceEndpointEnv)).Apply(&cfg)
	config.WithContentEndpoint(config.Settings().GetOption(config.PodopsContentEndpointEnv)).Apply(&cfg)

	if err := cfg.WriteToFile(secretPath); err != nil {
		return api.ErrorResponse(c, http.StatusInternalServerError, podops.ErrInternalError)
	}

	cdn.MarkStorageChanged() // make sure we earmark the content direcory as "changed"
	auth.RegisterAuthorization(&cfg)

	return api.StandardResponse(c, http.StatusOK, cfg)
}

// AssetUploadEndpoint implements content upload to the CDN
func AssetUploadEndpoint(c echo.Context) error {
	ctx := context.Background()

	// basic auth validation
	cfg, err := auth.CheckAuthorization(ctx, c, config.ScopeContentWrite)
	if err != nil {
		return api.ErrorResponse(c, http.StatusUnauthorized, err)
	}

	parent := c.Param("parent")
	if parent == "" {
		return api.ErrorResponse(c, http.StatusBadRequest, podops.ErrInvalidRoute)
	}
	if !canAccessContent(cfg.Credentials.ProjectID, parent, config.ScopeContentWrite) {
		return api.ErrorResponse(c, http.StatusUnauthorized, nil)
	}

	MeterAPIRequest(ctx, c.Request(), parent, "api.asset.upload")

	// validate the existence of the target location first in order to prevent random uploads
	contentRoot := filepath.Join(config.StorageLocation, parent)
	if _, err := os.Stat(contentRoot); os.IsNotExist(err) {
		return api.ErrorResponse(c, http.StatusBadRequest, podops.ErrInvalidGUID)
	}

	_, err = cdn.ReceiveFileUpload(ctx, c.Request(), contentRoot, UploadFormName)
	if err != nil {
		return api.ErrorResponse(c, http.StatusInternalServerError, err)
	}

	cdn.MarkStorageChanged() // make sure we earmark the content direcory as "changed"

	return api.StandardResponse(c, http.StatusOK, nil)
}

// AssetListEndpoint returns a list of media assets for a given show
func AssetListEndpoint(c echo.Context) error {
	ctx := context.Background()

	// basic auth validation
	cfg, err := auth.CheckAuthorization(ctx, c, config.ScopeContentRead)
	if err != nil {
		return api.ErrorResponse(c, http.StatusUnauthorized, err)
	}

	parent := c.Param("parent")
	if parent == "" {
		return api.ErrorResponse(c, http.StatusBadRequest, podops.ErrInvalidRoute)
	}
	if !canAccessContent(cfg.Credentials.ProjectID, parent, config.ScopeContentRead) {
		return api.ErrorResponse(c, http.StatusUnauthorized, nil)
	}

	MeterAPIRequest(ctx, c.Request(), parent, "api.asset.list")

	// validate the existence of the target location
	contentRoot := filepath.Join(config.StorageLocation, parent)
	if _, err := os.Stat(contentRoot); os.IsNotExist(err) {
		return api.ErrorResponse(c, http.StatusBadRequest, podops.ErrInvalidGUID)
	}

	r, err := cdn.ListResources(ctx, contentRoot)
	if err != nil {
		return api.ErrorResponse(c, http.StatusInternalServerError, podops.ErrInternalError)
	}

	return api.StandardResponse(c, http.StatusOK, r)
}

// AssetDeleteEndpoint removes a media asset from the CDN
func AssetDeleteEndpoint(c echo.Context) error {
	ctx := context.Background()

	// basic auth validation
	cfg, err := auth.CheckAuthorization(ctx, c, config.ScopeContentWrite)
	if err != nil {
		return api.ErrorResponse(c, http.StatusUnauthorized, err)
	}

	parent := c.Param("parent")
	if parent == "" {
		return api.ErrorResponse(c, http.StatusBadRequest, podops.ErrInvalidRoute)
	}
	asset := c.Param("asset")
	if parent == "" {
		return api.ErrorResponse(c, http.StatusBadRequest, podops.ErrInvalidRoute)
	}
	if !canAccessContent(cfg.Credentials.ProjectID, parent, config.ScopeContentWrite) {
		return api.ErrorResponse(c, http.StatusUnauthorized, nil)
	}

	MeterAPIRequest(ctx, c.Request(), parent, "api.asset.delete")

	// validate the existence of the target location
	assetPath := filepath.Join(config.StorageLocation, parent, asset)
	if _, err := os.Stat(assetPath); os.IsNotExist(err) {
		return api.ErrorResponse(c, http.StatusBadRequest, podops.ErrInvalidResourceName)
	}

	err = os.RemoveAll(assetPath)
	if err != nil {
		return api.ErrorResponse(c, http.StatusInternalServerError, podops.ErrInternalError)
	}

	return nil
}

func WebhookGithubEndpoint(c echo.Context) error {
	ctx := context.Background()

	// no basic auth here, we use the github secret

	parent := c.Param("parent")
	if parent == "" {
		return api.ErrorResponse(c, http.StatusBadRequest, podops.ErrInvalidRoute)
	}

	MeterAPIRequest(ctx, c.Request(), parent, "api.webhook.push")

	// find the secret key
	cfg, err := settings.ReadSettingsFromFile(filepath.Join(config.StorageLocation, parent, config.DefaultMasterKeyFileLocation))
	if err != nil {
		return api.ErrorResponse(c, http.StatusInternalServerError, err)
	}

	// parse and validate the github payload
	payload, err := github.ValidatePayload(c.Request(), []byte(cfg.Credentials.Token))
	if err != nil {
		return api.ErrorResponse(c, http.StatusInternalServerError, err)
	}

	event, err := github.ParseWebHook(github.WebHookType(c.Request()), payload)
	if err != nil {
		return api.ErrorResponse(c, http.StatusInternalServerError, err)
	}

	switch e := event.(type) {
	case *github.PushEvent:
		repo := *e.Repo.URL
		if err := cdn.CloneOrPullRepo(repo, parent); err != nil {
			return api.ErrorResponse(c, http.StatusInternalServerError, err)
		}
	default:
		return api.ErrorResponse(c, http.StatusInternalServerError, podops.ErrUnsupportedWebhookEvent)
	}

	// do something
	return api.StandardResponse(c, http.StatusOK, nil)
}
