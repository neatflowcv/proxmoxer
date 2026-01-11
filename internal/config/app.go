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

// AppConfig holds the application configuration
type AppConfig struct {
	ServerPort      string
	ProxmoxTimeout  time.Duration
	Logger          *log.Logger
}

// NewAppConfig creates default app configuration
func NewAppConfig() *AppConfig {
	return &AppConfig{
		ServerPort:     getEnv("SERVER_PORT", "8080"),
		ProxmoxTimeout: 30 * time.Second,
		Logger:         log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile),
	}
}

// InitializeApp initializes all application components
func InitializeApp(config *AppConfig) (*http.Router, error) {
	config.Logger.Println("Initializing application components...")

	// Initialize repository (in-memory for MVP)
	clusterRepo := persistence.NewMemoryRepository()
	config.Logger.Println("✓ Cluster repository initialized (in-memory)")

	// Initialize Proxmox client (will be initialized with actual URL per request)
	// For now, we create a dummy client - the actual endpoint comes from the request
	// insecureSkipVerify=true to support self-signed certificates
	proxmoxClient := proxmox.NewClient("https://pve.local", config.ProxmoxTimeout, true)
	config.Logger.Println("✓ Proxmox client initialized")

	// Initialize services
	clusterService := services.NewClusterService(clusterRepo, proxmoxClient, nil)
	config.Logger.Println("✓ Cluster service initialized")

	// Initialize router with all handlers
	router := http.NewRouter(clusterService, config.Logger)
	config.Logger.Println("✓ HTTP router initialized")

	config.Logger.Println("Application initialization completed successfully!")
	return router, nil
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
