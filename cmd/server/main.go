// Package main is the entrypoint for the GoFar application server.
// It initialises the application and starts listening on the configured address.
//
// Usage:
//
//	go run ./cmd/server
//
// The server address defaults to ":3000" and can be overridden via the APP_ADDR
// environment variable or the config system.
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/subhasundardas/gofar/framework/app"
	_ "github.com/subhasundardas/gofar/modules/all"
)

func main() {

	// Build and bootstrap the application (loads config, fiber, events, modules).
	application, err := app.New()
	if err != nil {
		log.Fatalf("failed to initialise application: %v", err)
	}

	// Determine the listen address. Config key "APP_ADDR" overrides the default.
	addr := application.Config().GetWithDefault("APP_ADDR", ":3000")

	// Graceful shutdown: listen for OS termination signals in a goroutine.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		application.Logger().Info("shutdown signal received, stopping server...")
		if err := application.Shutdown(); err != nil {
			log.Printf("error during shutdown: %v", err)
		}
	}()

	application.Logger().Infof("GoFar server starting on %s", addr)

	// Run blocks until the server is shut down.
	if err := application.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
