package main

import (
	caddycmd "github.com/caddyserver/caddy/v2/cmd"

	// plug in Caddy modules here
	_ "github.com/caddyserver/caddy/v2/modules/standard"
	_ "github.com/podops/podops/internal/cdn/modules"
)

func init() {
	/*
		// init the Google Observer if a setup is detected
		if env.Exists("PROJECT_ID") {
			if env.Exists("GOOGLE_APPLICATION_CREDENTIALS") {
				loggerConfig := provider.WithProvider("google.cloud.logging", observer.TypeLogger, stackdriver.NewGoogleStackdriverProvider)
				errorConfig := provider.WithProvider("google.cloud.error", observer.TypeErrorReporter, stackdriver.NewGoogleStackdriverProvider)
				metricsConfig := provider.WithProvider("google.cloud.metrics", observer.TypeMetrics, stackdriver.NewGoogleStackdriverProvider)
				observer.NewConfig(loggerConfig, errorConfig, metricsConfig)
			}
		}
	*/
}

func main() {
	caddycmd.Main()
}
