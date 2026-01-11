package persistence

import (
	"context"
	"fmt"
	"sync"

	"github.com/neatflowcv/proxmoxer/internal/domain/cluster"
	"github.com/neatflowcv/proxmoxer/internal/domain/common"
)

// MemoryRepository is an in-memory implementation of cluster.Repository
// Suitable for MVP and development. Use a proper database for production.
type MemoryRepository struct {
	mu       sync.RWMutex
	clusters map[string]*cluster.Cluster
}

// NewMemoryRepository creates a new in-memory cluster repository.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		mu:       sync.RWMutex{},
		clusters: make(map[string]*cluster.Cluster),
	}
}

// Save creates or updates a cluster in memory.
func (r *MemoryRepository) Save(ctx context.Context, c *cluster.Cluster) error {
	if c == nil {
		return common.ErrClusterNil
	}

	err := c.Validate()
	if err != nil {
		return fmt.Errorf("invalid cluster: %w", err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.clusters[c.ID] = c

	return nil
}

// FindByID retrieves a cluster by its ID.
func (r *MemoryRepository) FindByID(ctx context.Context, id string) (*cluster.Cluster, error) {
	if id == "" {
		return nil, fmt.Errorf("cluster id cannot be empty: %w", common.ErrInvalidClusterID)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	c, ok := r.clusters[id]
	if !ok {
		return nil, fmt.Errorf("cluster with id %s not found: %w", id, common.ErrClusterNotFound)
	}

	return c, nil
}

// FindByName retrieves a cluster by its name.
func (r *MemoryRepository) FindByName(ctx context.Context, name string) (*cluster.Cluster, error) {
	if name == "" {
		return nil, common.ErrClusterNameEmpty
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, c := range r.clusters {
		if c.Name == name {
			return c, nil
		}
	}

	return nil, fmt.Errorf("cluster with name %s not found: %w", name, common.ErrClusterNotFound)
}

// List retrieves all registered clusters.
func (r *MemoryRepository) List(ctx context.Context) ([]*cluster.Cluster, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	clusters := make([]*cluster.Cluster, 0, len(r.clusters))
	for _, c := range r.clusters {
		clusters = append(clusters, c)
	}

	return clusters, nil
}

// Delete removes a cluster by its ID.
func (r *MemoryRepository) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("cluster id cannot be empty: %w", common.ErrInvalidClusterID)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.clusters[id]; !ok {
		return fmt.Errorf("cluster with id %s not found: %w", id, common.ErrClusterNotFound)
	}

	delete(r.clusters, id)

	return nil
}

// Exists checks if a cluster with the given ID exists.
func (r *MemoryRepository) Exists(ctx context.Context, id string) (bool, error) {
	if id == "" {
		return false, nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.clusters[id]

	return ok, nil
}
