package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/neatflowcv/proxmoxer/internal/config"
)

func main() {
	appConfig := config.NewAppConfig()

	logStartup(appConfig)

	router, err := config.InitializeApp(appConfig)
	if err != nil {
		appConfig.Logger.Fatalf("Failed to initialize application: %v", err)
		os.Exit(1)
	}

	addr := ":" + appConfig.ServerPort
	server := createServer(appConfig, addr, router)

	appConfig.Logger.Println("==============================================")
	appConfig.Logger.Printf("Starting server on %s\n", addr)
	appConfig.Logger.Println("==============================================")

	// Channel to notify when server has shut down
	serverErrors := make(chan error, 1)

	// Start server in a goroutine
	go func() {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
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

		return
	case sig := <-sigChan:
		appConfig.Logger.Printf("\nReceived signal: %v\n", sig)
		appConfig.Logger.Println("Initiating graceful shutdown...")
	}

	shutdownServer(appConfig, server)
}

func logStartup(appConfig *config.AppConfig) {
	appConfig.Logger.Println("==============================================")
	appConfig.Logger.Println("Proxmoxer API Server Starting")
	appConfig.Logger.Println("==============================================")
	appConfig.Logger.Printf("Version: 0.1.0-mvp\n")
	appConfig.Logger.Printf("Server Port: %s\n", appConfig.ServerPort)
}

func createServer(appConfig *config.AppConfig, addr string, router http.Handler) *http.Server {
	const idleTimeoutMultiplier = 30

	return &http.Server{
		Addr:                         addr,
		Handler:                      router,
		DisableGeneralOptionsHandler: false,
		TLSConfig:                    nil,
		ReadTimeout:                  0,
		ReadHeaderTimeout:            appConfig.ProxmoxTimeout,
		WriteTimeout:                 appConfig.ProxmoxTimeout,
		IdleTimeout:                  idleTimeoutMultiplier * appConfig.ProxmoxTimeout,
		MaxHeaderBytes:               0,
		TLSNextProto:                 nil,
		ConnState:                    nil,
		ErrorLog:                     nil,
		BaseContext:                  nil,
		ConnContext:                  nil,
		HTTP2:                        nil,
		Protocols:                    nil,
	}
}

func shutdownServer(appConfig *config.AppConfig, server *http.Server) {
	const shutdownTimeout = 30 * time.Second

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	err := server.Shutdown(shutdownCtx)
	if err != nil {
		appConfig.Logger.Printf("Error during graceful shutdown: %v\n", err)
		os.Exit(1) //nolint:gocritic // Defer cancel() not needed as os.Exit stops execution
	}

	appConfig.Logger.Println("Server shutdown completed successfully")
}
