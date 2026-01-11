package services

import (
	"context"
	"log"
	"testing"

	"github.com/neatflowcv/proxmoxer/internal/application/dto"
	"github.com/neatflowcv/proxmoxer/internal/domain/cluster"
	"github.com/neatflowcv/proxmoxer/internal/domain/common"
	"github.com/neatflowcv/proxmoxer/internal/infrastructure/persistence"
)

// mockProxmoxClient is a mock implementation of Proxmox client for testing
type mockProxmoxClient struct {
	authenticateFn func(ctx context.Context, username string, password string) (ticket string, csrf string, err error)
	getVersionFn   func(ctx context.Context, ticket string) (version string, err error)
	getNodeCountFn func(ctx context.Context, ticket string) (count int, err error)
}

func (m *mockProxmoxClient) Authenticate(ctx context.Context, username string, password string) (ticket string, csrf string, err error) {
	if m.authenticateFn != nil {
		return m.authenticateFn(ctx, username, password)
	}
	return "test-ticket", "test-csrf", nil
}

func (m *mockProxmoxClient) GetVersion(ctx context.Context, ticket string) (version string, err error) {
	if m.getVersionFn != nil {
		return m.getVersionFn(ctx, ticket)
	}
	return "7.4-1", nil
}

func (m *mockProxmoxClient) GetNodeCount(ctx context.Context, ticket string) (count int, err error) {
	if m.getNodeCountFn != nil {
		return m.getNodeCountFn(ctx, ticket)
	}
	return 3, nil
}

func TestRegisterCluster_Success(t *testing.T) {
	ctx := context.Background()
	repo := persistence.NewMemoryRepository()
	mockClient := &mockProxmoxClient{}
	logger := NewSimpleLogger(log.Default())
	service := NewClusterService(repo, mockClient, logger)

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
	ctx := context.Background()
	repo := persistence.NewMemoryRepository()
	mockClient := &mockProxmoxClient{}
	logger := NewSimpleLogger(log.Default())
	service := NewClusterService(repo, mockClient, logger)

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
	ctx := context.Background()
	repo := persistence.NewMemoryRepository()
	mockClient := &mockProxmoxClient{}
	logger := NewSimpleLogger(log.Default())
	service := NewClusterService(repo, mockClient, logger)

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
	ctx := context.Background()
	repo := persistence.NewMemoryRepository()
	mockClient := &mockProxmoxClient{}
	logger := NewSimpleLogger(log.Default())
	service := NewClusterService(repo, mockClient, logger)

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

	service.RegisterCluster(ctx, req1)
	service.RegisterCluster(ctx, req2)

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
	ctx := context.Background()
	repo := persistence.NewMemoryRepository()
	mockClient := &mockProxmoxClient{}
	logger := NewSimpleLogger(log.Default())
	service := NewClusterService(repo, mockClient, logger)

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
	ctx := context.Background()
	repo := persistence.NewMemoryRepository()
	mockClient := &mockProxmoxClient{}
	logger := NewSimpleLogger(log.Default())
	service := NewClusterService(repo, mockClient, logger)

	// Try to deregister non-existent cluster
	err := service.DeregisterCluster(ctx, "non-existent-id")
	if err == nil {
		t.Fatal("expected error for non-existent cluster")
	}
}

func TestGetCluster(t *testing.T) {
	ctx := context.Background()
	repo := persistence.NewMemoryRepository()
	mockClient := &mockProxmoxClient{}
	logger := NewSimpleLogger(log.Default())
	service := NewClusterService(repo, mockClient, logger)

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
	ctx := context.Background()
	repo := persistence.NewMemoryRepository()
	mockClient := &mockProxmoxClient{
		authenticateFn: func(ctx context.Context, username string, password string) (ticket string, csrf string, err error) {
			return "", "", common.ErrAuthenticationFailed
		},
	}
	logger := NewSimpleLogger(log.Default())
	service := NewClusterService(repo, mockClient, logger)

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
