package config

import (
	"log"
	"os"
	"time"

	"github.com/neatflowcv/proxmoxer/internal/api/http"
	"github.com/neatflowcv/proxmoxer/internal/application/services"
	"github.com/neatflowcv/proxmoxer/internal/infrastructure/persistence"
	"github.com/neatflowcv/proxmoxer/internal/infrastructure/proxmox"
)

// proxmoxClientFactory implements services.ProxmoxClientFactory.
type proxmoxClientFactory struct {
	timeout            time.Duration
	insecureSkipVerify bool
}

//nolint:ireturn // Factory pattern requires returning interface for dependency injection and testability
func (f *proxmoxClientFactory) NewClient(baseURL string) services.ProxmoxClient {
	return proxmox.NewClient(baseURL, f.timeout, f.insecureSkipVerify)
}

// AppConfig holds the application configuration.
type AppConfig struct {
	ServerPort     string
	ProxmoxTimeout time.Duration
	Logger         *log.Logger
}

// NewAppConfig creates default app configuration.
func NewAppConfig() *AppConfig {
	const defaultProxmoxTimeout = 30 * time.Second

	return &AppConfig{
		ServerPort:     getEnv("SERVER_PORT", "8080"),
		ProxmoxTimeout: defaultProxmoxTimeout,
		Logger:         log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile),
	}
}

// InitializeApp initializes all application components.
func InitializeApp(config *AppConfig) (*http.Router, error) {
	config.Logger.Println("Initializing application components...")

	// Initialize repository (in-memory for MVP)
	clusterRepo := persistence.NewMemoryRepository()

	config.Logger.Println("✓ Cluster repository initialized (in-memory)")

	// Create Proxmox client factory
	// The factory creates a new client for each endpoint dynamically
	// insecureSkipVerify=true to support self-signed certificates
	clientFactory := &proxmoxClientFactory{
		timeout:            config.ProxmoxTimeout,
		insecureSkipVerify: true,
	}
	config.Logger.Println("✓ Proxmox client factory initialized")

	// Initialize services
	clusterService := services.NewClusterService(clusterRepo, clientFactory, nil)

	config.Logger.Println("✓ Cluster service initialized")

	// Initialize router with all handlers
	router := http.NewRouter(clusterService, config.Logger)
	config.Logger.Println("✓ HTTP router initialized")

	config.Logger.Println("Application initialization completed successfully!")

	return router, nil
}

// getEnv retrieves an environment variable or returns a default value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}
