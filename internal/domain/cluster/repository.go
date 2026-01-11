package cluster

import "context"

// Repository defines the interface for cluster persistence operations
// Implementations should handle in-memory storage, databases, or other persistence mechanisms.
type Repository interface {
	// Save creates or updates a cluster
	Save(ctx context.Context, cluster *Cluster) error

	// FindByID retrieves a cluster by its ID
	FindByID(ctx context.Context, id string) (*Cluster, error)

	// FindByName retrieves a cluster by its name
	FindByName(ctx context.Context, name string) (*Cluster, error)

	// List retrieves all registered clusters
	List(ctx context.Context) ([]*Cluster, error)

	// Delete removes a cluster by its ID
	Delete(ctx context.Context, id string) error

	// Exists checks if a cluster with the given ID exists
	Exists(ctx context.Context, id string) (bool, error)
}
