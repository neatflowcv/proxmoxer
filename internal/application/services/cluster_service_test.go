package services_test

import (
	"context"
	"log"
	"testing"

	"github.com/neatflowcv/proxmoxer/internal/application/dto"
	"github.com/neatflowcv/proxmoxer/internal/application/services"
	"github.com/neatflowcv/proxmoxer/internal/domain/cluster"
	"github.com/neatflowcv/proxmoxer/internal/domain/common"
	"github.com/neatflowcv/proxmoxer/internal/infrastructure/persistence"
	"github.com/neatflowcv/proxmoxer/internal/infrastructure/proxmox"
)

// mockProxmoxClient is a mock implementation of Proxmox client for testing.
type mockProxmoxClient struct {
	authenticateFn        func(ctx context.Context, username, password string) (ticket, csrf string, err error)
	getVersionFn          func(ctx context.Context, ticket string) (string, error)
	getNodeCountFn        func(ctx context.Context, ticket string) (int, error)
	getNodesFn            func(ctx context.Context, ticket string) ([]proxmox.NodeInfo, error)
	getNodeDisksFn        func(ctx context.Context, ticket string, nodeName string) ([]proxmox.DiskInfo, error)
	getNodeStatusFn       func(ctx context.Context, ticket string, nodeName string) (*proxmox.NodeStatusData, error)
	getClusterResourcesFn func(ctx context.Context, ticket string) ([]proxmox.ClusterResource, error)
}

func (m *mockProxmoxClient) Authenticate(ctx context.Context, username, password string) (
	string, string, error) {
	if m.authenticateFn != nil {
		return m.authenticateFn(ctx, username, password)
	}

	return "test-ticket", "test-csrf", nil
}

func (m *mockProxmoxClient) GetVersion(ctx context.Context, ticket string) (string, error) {
	if m.getVersionFn != nil {
		return m.getVersionFn(ctx, ticket)
	}

	return "7.4-1", nil
}

func (m *mockProxmoxClient) GetNodeCount(ctx context.Context, ticket string) (int, error) {
	if m.getNodeCountFn != nil {
		return m.getNodeCountFn(ctx, ticket)
	}

	return 3, nil
}

func (m *mockProxmoxClient) ListNodes(ctx context.Context, ticket string) ([]proxmox.NodeInfo, error) {
	if m.getNodesFn != nil {
		return m.getNodesFn(ctx, ticket)
	}

	return []proxmox.NodeInfo{
		{Node: "pve1", Status: "online"},
		{Node: "pve2", Status: "online"},
	}, nil
}

func (m *mockProxmoxClient) ListNodeDisks(
	ctx context.Context,
	ticket string,
	nodeName string,
) ([]proxmox.DiskInfo, error) {
	if m.getNodeDisksFn != nil {
		return m.getNodeDisksFn(ctx, ticket, nodeName)
	}

	return []proxmox.DiskInfo{
		{
			DevPath: "/dev/sda",
			Type:    "ssd",
			Size:    1000204886016,
			Model:   "Samsung SSD 870",
			Serial:  "S5VUNG0N123456",
			Vendor:  "ATA",
			Wearout: float64(98),
			Health:  "PASSED",
			Used:    "LVM",
			GPT:     0,
		},
	}, nil
}

func (m *mockProxmoxClient) GetNodeStatus(
	ctx context.Context,
	ticket string,
	nodeName string,
) (*proxmox.NodeStatusData, error) {
	if m.getNodeStatusFn != nil {
		return m.getNodeStatusFn(ctx, ticket, nodeName)
	}

	return &proxmox.NodeStatusData{
		CPU: 0.25,
		Memory: proxmox.MemoryStatus{
			Used:  8589934592,
			Total: 17179869184,
			Free:  8589934592,
		},
		Swap: proxmox.SwapStatus{
			Used:  0,
			Total: 4294967296,
			Free:  4294967296,
		},
		Uptime:  86400,
		LoadAvg: []float64{0.5, 0.7, 0.9},
	}, nil
}

func (m *mockProxmoxClient) GetClusterResources(
	ctx context.Context,
	ticket string,
) ([]proxmox.ClusterResource, error) {
	if m.getClusterResourcesFn != nil {
		return m.getClusterResourcesFn(ctx, ticket)
	}

	return []proxmox.ClusterResource{
		{ID: "qemu/100", Node: "pve1", Type: "qemu", Status: "running"},
		{ID: "qemu/101", Node: "pve1", Type: "qemu", Status: "stopped"},
		{ID: "lxc/200", Node: "pve2", Type: "lxc", Status: "running"},
	}, nil
}

// mockProxmoxClientFactory implements services.ProxmoxClientFactory for testing.
type mockProxmoxClientFactory struct {
	client services.ProxmoxClient
}

//nolint:ireturn // Factory pattern requires returning interface for dependency injection and testability
func (f *mockProxmoxClientFactory) NewClient(baseURL string) services.ProxmoxClient {
	return f.client
}

func TestRegisterCluster_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := persistence.NewMemoryRepository()
	mockClient := &mockProxmoxClient{
		authenticateFn: nil,
		getVersionFn:   nil,
		getNodeCountFn: nil,
		getNodesFn:     nil,
		getNodeDisksFn: nil,
	}
	mockFactory := &mockProxmoxClientFactory{client: mockClient}
	logger := services.NewSimpleLogger(log.Default())
	service := services.NewClusterService(repo, mockFactory, logger)

	req := &dto.RegisterClusterRequest{
		Name:        "test-cluster",
		APIEndpoint: "https://pve.example.com:8006",
		Username:    "root@pam",
		Password:    "password",
	}

	response, err := service.RegisterCluster(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if response == nil {
		t.Fatal("expected non-nil response")
	}

	if response.Name != req.Name {
		t.Errorf("expected name %s, got %s", req.Name, response.Name)
	}

	if response.APIEndpoint != req.APIEndpoint {
		t.Errorf("expected endpoint %s, got %s", req.APIEndpoint, response.APIEndpoint)
	}

	if response.Status != string(cluster.StatusHealthy) {
		t.Errorf("expected status %s, got %s", cluster.StatusHealthy, response.Status)
	}
}

func TestRegisterCluster_DuplicateName(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := persistence.NewMemoryRepository()
	mockClient := &mockProxmoxClient{
		authenticateFn: nil,
		getVersionFn:   nil,
		getNodeCountFn: nil,
		getNodesFn:     nil,
		getNodeDisksFn: nil,
	}
	mockFactory := &mockProxmoxClientFactory{client: mockClient}
	logger := services.NewSimpleLogger(log.Default())
	service := services.NewClusterService(repo, mockFactory, logger)

	req := &dto.RegisterClusterRequest{
		Name:        "test-cluster",
		APIEndpoint: "https://pve.example.com:8006",
		Username:    "root@pam",
		Password:    "password",
	}

	// Register first cluster
	_, err := service.RegisterCluster(ctx, req)
	if err != nil {
		t.Fatalf("first registration failed: %v", err)
	}

	// Try to register with same name
	_, err = service.RegisterCluster(ctx, req)
	if err == nil {
		t.Fatal("expected error for duplicate name")
	}
}

func TestRegisterCluster_InvalidRequest(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := persistence.NewMemoryRepository()
	mockClient := &mockProxmoxClient{
		authenticateFn: nil,
		getVersionFn:   nil,
		getNodeCountFn: nil,
		getNodesFn:     nil,
		getNodeDisksFn: nil,
	}
	mockFactory := &mockProxmoxClientFactory{client: mockClient}
	logger := services.NewSimpleLogger(log.Default())
	service := services.NewClusterService(repo, mockFactory, logger)

	// Test with empty name
	req := &dto.RegisterClusterRequest{
		Name:        "",
		APIEndpoint: "https://pve.example.com:8006",
		Username:    "root@pam",
		Password:    "password",
	}

	_, err := service.RegisterCluster(ctx, req)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestListClusters(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := persistence.NewMemoryRepository()
	mockClient := &mockProxmoxClient{
		authenticateFn: nil,
		getVersionFn:   nil,
		getNodeCountFn: nil,
		getNodesFn:     nil,
		getNodeDisksFn: nil,
	}
	mockFactory := &mockProxmoxClientFactory{client: mockClient}
	logger := services.NewSimpleLogger(log.Default())
	service := services.NewClusterService(repo, mockFactory, logger)

	// Register multiple clusters
	req1 := &dto.RegisterClusterRequest{
		Name:        "cluster-1",
		APIEndpoint: "https://pve1.example.com:8006",
		Username:    "root@pam",
		Password:    "password",
	}

	req2 := &dto.RegisterClusterRequest{
		Name:        "cluster-2",
		APIEndpoint: "https://pve2.example.com:8006",
		Username:    "root@pam",
		Password:    "password",
	}

	_, _ = service.RegisterCluster(ctx, req1)
	_, _ = service.RegisterCluster(ctx, req2)

	// List clusters
	response, err := service.ListClusters(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if response.Total != 2 {
		t.Errorf("expected 2 clusters, got %d", response.Total)
	}
}

func TestDeregisterCluster_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := persistence.NewMemoryRepository()
	mockClient := &mockProxmoxClient{
		authenticateFn: nil,
		getVersionFn:   nil,
		getNodeCountFn: nil,
		getNodesFn:     nil,
		getNodeDisksFn: nil,
	}
	mockFactory := &mockProxmoxClientFactory{client: mockClient}
	logger := services.NewSimpleLogger(log.Default())
	service := services.NewClusterService(repo, mockFactory, logger)

	// Register a cluster
	req := &dto.RegisterClusterRequest{
		Name:        "test-cluster",
		APIEndpoint: "https://pve.example.com:8006",
		Username:    "root@pam",
		Password:    "password",
	}

	response, err := service.RegisterCluster(ctx, req)
	if err != nil {
		t.Fatalf("registration failed: %v", err)
	}

	// Deregister the cluster
	err = service.DeregisterCluster(ctx, response.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify it's deleted
	_, err = service.GetCluster(ctx, response.ID)
	if err == nil {
		t.Fatal("expected error for deleted cluster")
	}
}

func TestDeregisterCluster_NotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := persistence.NewMemoryRepository()
	mockClient := &mockProxmoxClient{
		authenticateFn: nil,
		getVersionFn:   nil,
		getNodeCountFn: nil,
		getNodesFn:     nil,
		getNodeDisksFn: nil,
	}
	mockFactory := &mockProxmoxClientFactory{client: mockClient}
	logger := services.NewSimpleLogger(log.Default())
	service := services.NewClusterService(repo, mockFactory, logger)

	// Try to deregister non-existent cluster
	err := service.DeregisterCluster(ctx, "non-existent-id")
	if err == nil {
		t.Fatal("expected error for non-existent cluster")
	}
}

func TestGetCluster(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := persistence.NewMemoryRepository()
	mockClient := &mockProxmoxClient{
		authenticateFn: nil,
		getVersionFn:   nil,
		getNodeCountFn: nil,
		getNodesFn:     nil,
		getNodeDisksFn: nil,
	}
	mockFactory := &mockProxmoxClientFactory{client: mockClient}
	logger := services.NewSimpleLogger(log.Default())
	service := services.NewClusterService(repo, mockFactory, logger)

	// Register a cluster
	req := &dto.RegisterClusterRequest{
		Name:        "test-cluster",
		APIEndpoint: "https://pve.example.com:8006",
		Username:    "root@pam",
		Password:    "password",
	}

	registerResp, err := service.RegisterCluster(ctx, req)
	if err != nil {
		t.Fatalf("registration failed: %v", err)
	}

	// Get the cluster
	getResp, err := service.GetCluster(ctx, registerResp.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if getResp.ID != registerResp.ID {
		t.Errorf("expected ID %s, got %s", registerResp.ID, getResp.ID)
	}

	if getResp.Name != req.Name {
		t.Errorf("expected name %s, got %s", req.Name, getResp.Name)
	}
}

func TestAuthenticationFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := persistence.NewMemoryRepository()
	mockClient := &mockProxmoxClient{
		authenticateFn: func(ctx context.Context, username, password string) (string, string, error) {
			return "", "", common.ErrAuthenticationFailed
		},
		getVersionFn:   nil,
		getNodeCountFn: nil,
		getNodesFn:     nil,
		getNodeDisksFn: nil,
	}
	mockFactory := &mockProxmoxClientFactory{client: mockClient}
	logger := services.NewSimpleLogger(log.Default())
	service := services.NewClusterService(repo, mockFactory, logger)

	req := &dto.RegisterClusterRequest{
		Name:        "test-cluster",
		APIEndpoint: "https://pve.example.com:8006",
		Username:    "root@pam",
		Password:    "wrongpassword",
	}

	_, err := service.RegisterCluster(ctx, req)
	if err == nil {
		t.Fatal("expected authentication error")
	}
}
