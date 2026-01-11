package services

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/neatflowcv/proxmoxer/internal/application/dto"
	"github.com/neatflowcv/proxmoxer/internal/domain/cluster"
	"github.com/neatflowcv/proxmoxer/internal/domain/common"
)

// Logger interface for dependency injection.
type Logger interface {
	Info(msg string, fields ...any)
	Error(msg string, fields ...any)
	Warn(msg string, fields ...any)
}

// SimpleLogger is a basic implementation of Logger for MVP.
type SimpleLogger struct {
	logger *log.Logger
}

// NewSimpleLogger creates a new simple logger.
func NewSimpleLogger(logger *log.Logger) *SimpleLogger {
	return &SimpleLogger{logger: logger}
}

func (sl *SimpleLogger) Info(msg string, fields ...any) {
	sl.logger.Printf("[INFO] %s %v\n", msg, fields)
}

func (sl *SimpleLogger) Error(msg string, fields ...any) {
	sl.logger.Printf("[ERROR] %s %v\n", msg, fields)
}

func (sl *SimpleLogger) Warn(msg string, fields ...any) {
	sl.logger.Printf("[WARN] %s %v\n", msg, fields)
}

// ProxmoxClient defines the interface for Proxmox API operations.
type ProxmoxClient interface {
	Authenticate(ctx context.Context, username string, password string) (ticket string, csrf string, err error)
	GetVersion(ctx context.Context, ticket string) (version string, err error)
	GetNodeCount(ctx context.Context, ticket string) (count int, err error)
}

// ClusterService handles cluster-related use cases.
type ClusterService struct {
	clusterRepo   cluster.Repository
	proxmoxClient ProxmoxClient
	logger        Logger
}

// NewClusterService creates a new ClusterService instance.
func NewClusterService(
	repo cluster.Repository,
	client ProxmoxClient,
	logger Logger,
) *ClusterService {
	if logger == nil {
		logger = NewSimpleLogger(log.Default())
	}

	return &ClusterService{
		clusterRepo:   repo,
		proxmoxClient: client,
		logger:        logger,
	}
}

// RegisterCluster registers a new Proxmox cluster.
func (s *ClusterService) RegisterCluster(
	ctx context.Context,
	req *dto.RegisterClusterRequest,
) (*dto.ClusterResponse, error) {
	// Validate request
	err := s.validateRegisterRequest(req)
	if err != nil {
		s.logger.Error("Invalid register request", "error", err.Error())

		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if cluster with this name already exists
	_, findErr := s.clusterRepo.FindByName(ctx, req.Name)
	if findErr == nil {
		s.logger.Warn("Cluster name already exists", "name", req.Name)

		return nil, fmt.Errorf("cluster with name %s already exists: %w", req.Name, common.ErrClusterAlreadyExists)
	}

	s.logger.Info("Attempting to authenticate with Proxmox cluster", "endpoint", req.APIEndpoint)

	// Authenticate and fetch cluster info
	newCluster, err := s.createClusterFromRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	// Save to repository
	err = s.clusterRepo.Save(ctx, newCluster)
	if err != nil {
		s.logger.Error("Failed to save cluster", "error", err.Error())

		return nil, fmt.Errorf("failed to save cluster: %w", common.ErrInternalError)
	}

	s.logger.Info("Cluster registered successfully", "cluster_id", newCluster.ID, "name", req.Name)

	return s.clusterToResponse(newCluster), nil
}

// DeregisterCluster removes a registered cluster.
func (s *ClusterService) DeregisterCluster(ctx context.Context, clusterID string) error {
	if clusterID == "" {
		s.logger.Error("Empty cluster ID provided")

		return fmt.Errorf("cluster id cannot be empty: %w", common.ErrInvalidClusterID)
	}

	// Check if cluster exists
	_, err := s.clusterRepo.FindByID(ctx, clusterID)
	if err != nil {
		s.logger.Error("Cluster not found", "cluster_id", clusterID)

		return fmt.Errorf("cluster not found: %w", common.ErrClusterNotFound)
	}

	// Delete the cluster
	err = s.clusterRepo.Delete(ctx, clusterID)
	if err != nil {
		s.logger.Error("Failed to delete cluster", "cluster_id", clusterID, "error", err.Error())

		return fmt.Errorf("failed to delete cluster: %w", err)
	}

	s.logger.Info("Cluster deregistered successfully", "cluster_id", clusterID)

	return nil
}

// ListClusters returns all registered clusters.
func (s *ClusterService) ListClusters(ctx context.Context) (*dto.ListClustersResponse, error) {
	clusters, err := s.clusterRepo.List(ctx)
	if err != nil {
		s.logger.Error("Failed to list clusters", "error", err.Error())

		return nil, fmt.Errorf("failed to list clusters: %w", err)
	}

	responses := make([]dto.ClusterResponse, len(clusters))
	for i, c := range clusters {
		responses[i] = *s.clusterToResponse(c)
	}

	s.logger.Info("Listed clusters", "count", len(clusters))

	return &dto.ListClustersResponse{
		Clusters: responses,
		Total:    len(clusters),
	}, nil
}

// GetCluster retrieves a specific cluster by ID.
func (s *ClusterService) GetCluster(ctx context.Context, clusterID string) (*dto.ClusterResponse, error) {
	if clusterID == "" {
		return nil, fmt.Errorf("cluster id cannot be empty: %w", common.ErrInvalidClusterID)
	}

	c, err := s.clusterRepo.FindByID(ctx, clusterID)
	if err != nil {
		s.logger.Error("Cluster not found", "cluster_id", clusterID)

		return nil, fmt.Errorf("cluster not found: %w", common.ErrClusterNotFound)
	}

	return s.clusterToResponse(c), nil
}

// createClusterFromRequest creates a cluster entity by authenticating with Proxmox.
func (s *ClusterService) createClusterFromRequest(ctx context.Context,
	req *dto.RegisterClusterRequest) (*cluster.Cluster, error) {
	// Authenticate with Proxmox API to validate credentials
	ticket, _, err := s.proxmoxClient.Authenticate(ctx, req.Username, req.Password)
	if err != nil {
		s.logger.Error("Proxmox authentication failed", "error", err.Error())

		return nil, fmt.Errorf("authentication failed: %w", common.ErrAuthenticationFailed)
	}

	// Get cluster version
	version, err := s.proxmoxClient.GetVersion(ctx, ticket)
	if err != nil {
		s.logger.Error("Failed to get Proxmox version", "error", err.Error())

		version = "unknown"
	}

	// Get node count
	nodeCount, err := s.proxmoxClient.GetNodeCount(ctx, ticket)
	if err != nil {
		s.logger.Error("Failed to get node count", "error", err.Error())

		nodeCount = 0
	}

	// Create cluster entity with unique ID
	clusterID := uuid.New().String()
	newCluster := cluster.NewCluster(
		clusterID,
		req.Name,
		req.APIEndpoint,
		req.Username,
		req.Password,
	)

	newCluster.UpdateProxmoxVersion(version)
	newCluster.UpdateNodeCount(nodeCount)
	newCluster.UpdateStatus(cluster.StatusHealthy)

	return newCluster, nil
}

// validateRegisterRequest validates the register cluster request.
func (s *ClusterService) validateRegisterRequest(req *dto.RegisterClusterRequest) error {
	if req == nil {
		return common.ErrRequestNil
	}

	if req.Name == "" {
		return common.ErrClusterNameRequired
	}

	const maxNameLength = 255
	if len(req.Name) > maxNameLength {
		return common.ErrClusterNameTooLong
	}

	if req.APIEndpoint == "" {
		return common.ErrAPIEndpointRequired
	}

	if req.Username == "" {
		return common.ErrUsernameRequired
	}

	if req.Password == "" {
		return common.ErrPasswordRequired
	}

	return nil
}

// clusterToResponse converts a domain cluster entity to a response DTO.
func (s *ClusterService) clusterToResponse(c *cluster.Cluster) *dto.ClusterResponse {
	return &dto.ClusterResponse{
		ID:             c.ID,
		Name:           c.Name,
		APIEndpoint:    c.APIEndpoint,
		Status:         string(c.Status),
		ProxmoxVersion: c.ProxmoxVersion,
		NodeCount:      c.NodeCount,
		CreatedAt:      c.CreatedAt,
		UpdatedAt:      c.UpdatedAt,
	}
}
