package cluster

import (
	"errors"
	"time"
)

// ClusterStatus represents the health status of a cluster
type ClusterStatus string

const (
	StatusHealthy   ClusterStatus = "healthy"
	StatusDegraded  ClusterStatus = "degraded"
	StatusUnhealthy ClusterStatus = "unhealthy"
	StatusUnknown   ClusterStatus = "unknown"
)

// Cluster represents a Proxmox cluster managed by the system
type Cluster struct {
	// Unique identifier for the cluster
	ID string
	// Human-readable name of the cluster
	Name string
	// Proxmox cluster API URL
	APIEndpoint string
	// Proxmox username for authentication
	Username string
	// Proxmox password (stored in memory, no encryption in MVP)
	Password string
	// Current health status of the cluster
	Status ClusterStatus
	// Proxmox version running on the cluster
	ProxmoxVersion string
	// Number of nodes in the cluster
	NodeCount int
	// When the cluster was registered
	CreatedAt time.Time
	// Last time the cluster information was updated
	UpdatedAt time.Time
}

// NewCluster creates a new Cluster instance
func NewCluster(
	id string,
	name string,
	apiEndpoint string,
	username string,
	password string,
) *Cluster {
	now := time.Now()
	return &Cluster{
		ID:          id,
		Name:        name,
		APIEndpoint: apiEndpoint,
		Username:    username,
		Password:    password,
		Status:      StatusUnknown,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// UpdateStatus updates the cluster status and timestamp
func (c *Cluster) UpdateStatus(status ClusterStatus) {
	c.Status = status
	c.UpdatedAt = time.Now()
}

// UpdateNodeCount updates the node count
func (c *Cluster) UpdateNodeCount(count int) {
	c.NodeCount = count
	c.UpdatedAt = time.Now()
}

// UpdateProxmoxVersion updates the Proxmox version
func (c *Cluster) UpdateProxmoxVersion(version string) {
	c.ProxmoxVersion = version
	c.UpdatedAt = time.Now()
}

// IsHealthy returns true if the cluster is in a healthy state
func (c *Cluster) IsHealthy() bool {
	return c.Status == StatusHealthy
}

// Validate performs basic validation on the cluster entity
func (c *Cluster) Validate() error {
	if c.ID == "" {
		return errors.New("cluster id cannot be empty")
	}
	if c.Name == "" {
		return errors.New("cluster name cannot be empty")
	}
	if c.APIEndpoint == "" {
		return errors.New("api endpoint cannot be empty")
	}
	if c.Username == "" {
		return errors.New("username cannot be empty")
	}
	if c.Password == "" {
		return errors.New("password cannot be empty")
	}
	return nil
}
