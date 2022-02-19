package main

import (
	"context"
	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	httpapi "github.com/txsvc/httpservice/pkg/api"
	"github.com/txsvc/httpservice/pkg/httpserver"

	"github.com/podops/podops/auth"
	"github.com/podops/podops/config"
	"github.com/podops/podops/internal/api"
	"github.com/podops/podops/internal/cdn"
)

func init() {
	// register the main credentials
	cfg, _ := config.LoadClientSettings("")
	auth.RegisterAuthorization(cfg)
	// register all client crdentials
	if err := cdn.CreateCredentialsMappings(context.TODO(), config.StorageLocation); err != nil {
		log.Fatal(err)
	}
}

func setup() *echo.Echo {
	// create a new router instance
	e := echo.New()

	// add and configure the middlewares
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.DefaultCORSConfig))
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "ts=${time_unix}, m=${method}, uri=${uri}, ua=${user_agent}, st=${status}, dt=${latency}\n"}))

	// API endpoint namespace
	apiEndpoints := e.Group(api.NamespacePrefix)

	// asset related routes
	apiEndpoints.POST(api.AssetRoute, api.AssetUploadEndpoint)
	apiEndpoints.GET(api.AssetRoute, api.AssetListEndpoint)
	apiEndpoints.DELETE(api.AssetDeleteRoute, api.AssetDeleteEndpoint)

	// admin endpoints
	apiEndpoints.PUT(api.InitRoute, api.InitEndpoint)
	apiEndpoints.POST(api.WebhookRoute, api.WebhookGithubEndpoint)

	// default endpoint to catch random requests
	e.GET("/", httpapi.DefaultEndpoint)

	return e
}

func shutdown(*echo.Echo) {
	// TODO: implement your own stuff here
}

func main() {
	service, err := httpserver.New(setup, shutdown, nil)
	if err != nil {
		log.Fatal(err)
	}
	service.StartBlocking()
}
