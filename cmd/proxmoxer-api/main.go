package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

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

	// Start server
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
		os.Exit(1)
	}
}
