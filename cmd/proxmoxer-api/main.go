package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/neatflowcv/proxmoxer/internal/config"
)

func main() {
	// Initialize configuration
	appConfig := config.NewAppConfig()

	appConfig.Logger.Println("==============================================")
	appConfig.Logger.Println("Proxmoxer API Server Starting")
	appConfig.Logger.Println("==============================================")
	appConfig.Logger.Printf("Version: 0.1.0-mvp\n")
	appConfig.Logger.Printf("Server Port: %s\n", appConfig.ServerPort)

	// Initialize application components
	router, err := config.InitializeApp(appConfig)
	if err != nil {
		appConfig.Logger.Fatalf("Failed to initialize application: %v", err)
		os.Exit(1)
	}

	// Create HTTP server
	addr := fmt.Sprintf(":%s", appConfig.ServerPort)
	server := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: appConfig.ProxmoxTimeout,
		WriteTimeout:      appConfig.ProxmoxTimeout,
		IdleTimeout:       30 * appConfig.ProxmoxTimeout,
	}

	appConfig.Logger.Println("==============================================")
	appConfig.Logger.Printf("Starting server on %s\n", addr)
	appConfig.Logger.Println("==============================================")

	// Channel to notify when server has shut down
	serverErrors := make(chan error, 1)

	// Start server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- fmt.Errorf("server error: %w", err)
		}
	}()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Wait for either a signal or server error
	select {
	case err := <-serverErrors:
		appConfig.Logger.Fatalf("%v", err)
		os.Exit(1)
	case sig := <-sigChan:
		appConfig.Logger.Printf("\nReceived signal: %v\n", sig)
		appConfig.Logger.Println("Initiating graceful shutdown...")
	}

	// Create a context with 30-second timeout for graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Gracefully shutdown the server
	if err := server.Shutdown(shutdownCtx); err != nil {
		appConfig.Logger.Printf("Error during graceful shutdown: %v\n", err)
		os.Exit(1)
	}

	appConfig.Logger.Println("Server shutdown completed successfully")
}
