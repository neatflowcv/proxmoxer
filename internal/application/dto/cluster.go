package dto

import (
	"time"
)

// RegisterClusterRequest is the request DTO for registering a new cluster
type RegisterClusterRequest struct {
	// Human-readable name for the cluster
	Name string `json:"name" binding:"required,max=255"`
	// Proxmox API endpoint URL (e.g., https://pve.example.com:8006)
	APIEndpoint string `json:"api_endpoint" binding:"required,url"`
	// Proxmox username for authentication
	Username string `json:"username" binding:"required,max=255"`
	// Proxmox password for authentication
	Password string `json:"password" binding:"required,min=1"`
}

// DeregisterClusterRequest is the request DTO for deregistering a cluster
type DeregisterClusterRequest struct {
	// Cluster ID to deregister
	ClusterID string `json:"cluster_id" binding:"required"`
}

// ListClustersResponse is the response DTO for listing clusters
type ListClustersResponse struct {
	// List of clusters
	Clusters []ClusterResponse `json:"clusters"`
	// Total count of clusters
	Total int `json:"total"`
}

// ClusterResponse is the response DTO for a single cluster
type ClusterResponse struct {
	// Unique identifier
	ID string `json:"id"`
	// Human-readable name
	Name string `json:"name"`
	// Proxmox API endpoint
	APIEndpoint string `json:"api_endpoint"`
	// Current status of the cluster
	Status string `json:"status"`
	// Proxmox version
	ProxmoxVersion string `json:"proxmox_version"`
	// Number of nodes in the cluster
	NodeCount int `json:"node_count"`
	// When the cluster was registered
	CreatedAt time.Time `json:"created_at"`
	// Last update time
	UpdatedAt time.Time `json:"updated_at"`
}

// ErrorResponse is the standard error response DTO
type ErrorResponse struct {
	// Error code
	Code string `json:"code"`
	// Error message
	Message string `json:"message"`
	// Additional details
	Details map[string]interface{} `json:"details,omitempty"`
}
