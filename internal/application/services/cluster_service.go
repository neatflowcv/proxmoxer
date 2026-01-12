package services

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/neatflowcv/proxmoxer/internal/application/dto"
	"github.com/neatflowcv/proxmoxer/internal/domain/cluster"
	"github.com/neatflowcv/proxmoxer/internal/domain/common"
	"github.com/neatflowcv/proxmoxer/internal/infrastructure/proxmox"
	"golang.org/x/sync/errgroup"
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
	ListNodes(ctx context.Context, ticket string) ([]proxmox.NodeInfo, error)
	ListNodeDisks(ctx context.Context, ticket string, nodeName string) ([]proxmox.DiskInfo, error)
	GetNodeStatus(ctx context.Context, ticket string, nodeName string) (*proxmox.NodeStatusData, error)
	GetClusterResources(ctx context.Context, ticket string) ([]proxmox.ClusterResource, error)
}

// ProxmoxClientFactory defines the interface for creating new ProxmoxClient instances.
type ProxmoxClientFactory interface {
	NewClient(baseURL string) ProxmoxClient
}

// ClusterService handles cluster-related use cases.
type ClusterService struct {
	clusterRepo         cluster.Repository
	proxmoxClientFactory ProxmoxClientFactory
	logger              Logger
}

// NewClusterService creates a new ClusterService instance.
func NewClusterService(
	repo cluster.Repository,
	clientFactory ProxmoxClientFactory,
	logger Logger,
) *ClusterService {
	if logger == nil {
		logger = NewSimpleLogger(log.Default())
	}

	return &ClusterService{
		clusterRepo:          repo,
		proxmoxClientFactory: clientFactory,
		logger:               logger,
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

// ListClusterDisks retrieves disk information for all nodes in a cluster.
func (s *ClusterService) ListClusterDisks(ctx context.Context, clusterID string) (*dto.ClusterDisksResponse, error) {
	if clusterID == "" {
		return nil, fmt.Errorf("cluster id cannot be empty: %w", common.ErrInvalidClusterID)
	}

	// Get cluster from repository
	c, err := s.clusterRepo.FindByID(ctx, clusterID)
	if err != nil {
		s.logger.Error("Cluster not found", "cluster_id", clusterID)

		return nil, fmt.Errorf("cluster not found: %w", common.ErrClusterNotFound)
	}

	// Create Proxmox client and authenticate
	proxmoxClient := s.proxmoxClientFactory.NewClient(c.APIEndpoint)

	ticket, _, err := proxmoxClient.Authenticate(ctx, c.Username, c.Password)
	if err != nil {
		s.logger.Error("Proxmox authentication failed", "error", err.Error())

		return nil, fmt.Errorf("authentication failed: %w", common.ErrAuthenticationFailed)
	}

	// Get list of nodes
	nodes, err := proxmoxClient.ListNodes(ctx, ticket)
	if err != nil {
		s.logger.Error("Failed to get nodes", "error", err.Error())

		return nil, fmt.Errorf("failed to get nodes: %w", err)
	}

	// Fetch disks for all nodes in parallel
	nodeDisks, totalDisks, err := s.fetchNodeDisksParallel(ctx, proxmoxClient, ticket, nodes)
	if err != nil {
		s.logger.Error("Error fetching disks", "error", err.Error())

		return nil, fmt.Errorf("failed to fetch disks: %w", err)
	}

	s.logger.Info("Cluster disks retrieved successfully", "cluster_id", clusterID, "total_disks", totalDisks)

	return &dto.ClusterDisksResponse{
		ClusterID:   c.ID,
		ClusterName: c.Name,
		Nodes:       nodeDisks,
		TotalDisks:  totalDisks,
	}, nil
}

// fetchNodeDisksParallel fetches disk information for all nodes in parallel.
func (s *ClusterService) fetchNodeDisksParallel(
	ctx context.Context,
	proxmoxClient ProxmoxClient,
	ticket string,
	nodes []proxmox.NodeInfo,
) ([]dto.NodeDisksResponse, int, error) {
	nodeDisks := make([]dto.NodeDisksResponse, len(nodes))

	var totalDisks int

	var mu sync.Mutex

	// Get disks for each node in parallel
	g, gctx := errgroup.WithContext(ctx)

	for i, node := range nodes {
		idx := i
		n := node

		g.Go(func() error {
			nodeResponse := dto.NodeDisksResponse{
				NodeName: n.Node,
				Status:   n.Status,
				Disks:    []dto.DiskResponse{},
				Error:    "",
			}

			disks, diskErr := proxmoxClient.ListNodeDisks(gctx, ticket, n.Node)
			if diskErr != nil {
				s.logger.Warn("Failed to get disks for node", "node", n.Node, "error", diskErr.Error())
				nodeResponse.Error = diskErr.Error()
			} else {
				for _, disk := range disks {
					nodeResponse.Disks = append(nodeResponse.Disks, s.diskInfoToResponse(disk))
				}
			}

			mu.Lock()

			nodeDisks[idx] = nodeResponse
			totalDisks += len(nodeResponse.Disks)

			mu.Unlock()

			return nil // Don't fail on individual node errors
		})
	}

	// Wait for all goroutines to complete
	err := g.Wait()
	if err != nil {
		return nil, 0, fmt.Errorf("error fetching node disks: %w", err)
	}

	return nodeDisks, totalDisks, nil
}

// diskInfoToResponse converts a proxmox DiskInfo to a DTO response.
func (s *ClusterService) diskInfoToResponse(disk proxmox.DiskInfo) dto.DiskResponse {
	wearout := -1
	if w, ok := disk.Wearout.(float64); ok {
		wearout = int(w)
	}

	return dto.DiskResponse{
		Device:  disk.DevPath,
		Type:    disk.Type,
		Size:    disk.Size,
		Model:   disk.Model,
		Serial:  disk.Serial,
		Vendor:  disk.Vendor,
		Wearout: wearout,
		Health:  disk.Health,
		Used:    disk.Used,
	}
}

// createClusterFromRequest creates a cluster entity by authenticating with Proxmox.
func (s *ClusterService) createClusterFromRequest(ctx context.Context,
	req *dto.RegisterClusterRequest) (*cluster.Cluster, error) {
	// Create client using factory with the endpoint from the request
	proxmoxClient := s.proxmoxClientFactory.NewClient(req.APIEndpoint)

	// Authenticate with Proxmox API to validate credentials
	ticket, _, err := proxmoxClient.Authenticate(ctx, req.Username, req.Password)
	if err != nil {
		s.logger.Error("Proxmox authentication failed", "error", err.Error())

		return nil, fmt.Errorf("authentication failed: %w", common.ErrAuthenticationFailed)
	}

	// Get cluster version
	version, err := proxmoxClient.GetVersion(ctx, ticket)
	if err != nil {
		s.logger.Error("Failed to get Proxmox version", "error", err.Error())

		version = "unknown"
	}

	// Get node count
	nodeCount, err := proxmoxClient.GetNodeCount(ctx, ticket)
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

// GetClusterStatus retrieves monitoring status for all nodes in a cluster.
func (s *ClusterService) GetClusterStatus(ctx context.Context, clusterID string) (*dto.ClusterStatusResponse, error) {
	if clusterID == "" {
		return nil, fmt.Errorf("cluster id cannot be empty: %w", common.ErrInvalidClusterID)
	}

	// Get cluster from repository
	c, err := s.clusterRepo.FindByID(ctx, clusterID)
	if err != nil {
		s.logger.Error("Cluster not found", "cluster_id", clusterID)

		return nil, fmt.Errorf("cluster not found: %w", common.ErrClusterNotFound)
	}

	// Create Proxmox client and authenticate
	proxmoxClient := s.proxmoxClientFactory.NewClient(c.APIEndpoint)

	ticket, _, err := proxmoxClient.Authenticate(ctx, c.Username, c.Password)
	if err != nil {
		s.logger.Error("Proxmox authentication failed", "error", err.Error())

		return nil, fmt.Errorf("authentication failed: %w", common.ErrAuthenticationFailed)
	}

	// Get list of nodes
	nodes, err := proxmoxClient.ListNodes(ctx, ticket)
	if err != nil {
		s.logger.Error("Failed to get nodes", "error", err.Error())

		return nil, fmt.Errorf("failed to get nodes: %w", err)
	}

	// Get cluster resources for VM/container counts
	resources, err := proxmoxClient.GetClusterResources(ctx, ticket)
	if err != nil {
		s.logger.Warn("Failed to get cluster resources", "error", err.Error())
		resources = []proxmox.ClusterResource{}
	}

	// Fetch status for all nodes in parallel
	nodeStatuses := s.fetchNodeStatusesParallel(ctx, proxmoxClient, ticket, nodes)

	// Calculate resource summary
	resourceSummary := s.calculateResourceSummary(resources)

	s.logger.Info("Cluster status retrieved successfully", "cluster_id", clusterID)

	return &dto.ClusterStatusResponse{
		ClusterID:       c.ID,
		ClusterName:     c.Name,
		Nodes:           nodeStatuses,
		ResourceSummary: resourceSummary,
		FetchedAt:       time.Now(),
	}, nil
}

// fetchNodeStatusesParallel fetches status for all nodes in parallel.
func (s *ClusterService) fetchNodeStatusesParallel(
	ctx context.Context,
	proxmoxClient ProxmoxClient,
	ticket string,
	nodes []proxmox.NodeInfo,
) []dto.NodeStatusResponse {
	nodeStatuses := make([]dto.NodeStatusResponse, len(nodes))

	var mu sync.Mutex

	g, gctx := errgroup.WithContext(ctx)

	for i, node := range nodes {
		idx := i
		n := node

		g.Go(func() error {
			nodeResponse := dto.NodeStatusResponse{
				NodeName: n.Node,
				Status:   n.Status,
				LoadAvg:  []float64{},
			}

			status, statusErr := proxmoxClient.GetNodeStatus(gctx, ticket, n.Node)
			if statusErr != nil {
				s.logger.Warn("Failed to get status for node", "node", n.Node, "error", statusErr.Error())
				nodeResponse.Error = statusErr.Error()
			} else {
				nodeResponse.CPUUsage = status.CPU * 100
				nodeResponse.MemoryUsed = status.Memory.Used
				nodeResponse.MemoryTotal = status.Memory.Total
				if status.Memory.Total > 0 {
					nodeResponse.MemoryUsage = float64(status.Memory.Used) / float64(status.Memory.Total) * 100
				}
				nodeResponse.SwapUsed = status.Swap.Used
				nodeResponse.SwapTotal = status.Swap.Total
				if status.Swap.Total > 0 {
					nodeResponse.SwapUsage = float64(status.Swap.Used) / float64(status.Swap.Total) * 100
				}
				nodeResponse.Uptime = status.Uptime
				nodeResponse.LoadAvg = status.LoadAvg
			}

			mu.Lock()
			nodeStatuses[idx] = nodeResponse
			mu.Unlock()

			return nil
		})
	}

	_ = g.Wait()

	return nodeStatuses
}

// calculateResourceSummary calculates VM and container counts from cluster resources.
func (s *ClusterService) calculateResourceSummary(resources []proxmox.ClusterResource) dto.ResourceSummary {
	summary := dto.ResourceSummary{}

	for _, r := range resources {
		switch r.Type {
		case "qemu":
			summary.TotalVMs++
			if r.Status == "running" {
				summary.RunningVMs++
			}
		case "lxc":
			summary.TotalContainers++
			if r.Status == "running" {
				summary.RunningContainers++
			}
		}
	}

	return summary
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
