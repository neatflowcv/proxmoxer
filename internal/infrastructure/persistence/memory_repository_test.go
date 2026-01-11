package persistence_test

import (
	"context"
	"testing"

	"github.com/neatflowcv/proxmoxer/internal/domain/cluster"
	"github.com/neatflowcv/proxmoxer/internal/infrastructure/persistence"
)

func TestMemoryRepository_Save(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := persistence.NewMemoryRepository()

	c := cluster.NewCluster(
		"test-id",
		"test-cluster",
		"https://pve.example.com:8006",
		"root@pam",
		"password",
	)

	err := repo.Save(ctx, c)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify it was saved
	saved, err := repo.FindByID(ctx, "test-id")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if saved.Name != "test-cluster" {
		t.Errorf("expected name test-cluster, got %s", saved.Name)
	}
}

func TestMemoryRepository_FindByID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := persistence.NewMemoryRepository()

	// Try to find non-existent cluster
	_, err := repo.FindByID(ctx, "non-existent")
	if err == nil {
		t.Fatal("expected error for non-existent cluster")
	}

	// Save and find
	c := cluster.NewCluster(
		"test-id",
		"test-cluster",
		"https://pve.example.com:8006",
		"root@pam",
		"password",
	)
	_ = repo.Save(ctx, c)

	found, err := repo.FindByID(ctx, "test-id")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if found.ID != "test-id" {
		t.Errorf("expected ID test-id, got %s", found.ID)
	}
}

func TestMemoryRepository_FindByName(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := persistence.NewMemoryRepository()

	c := cluster.NewCluster(
		"test-id",
		"test-cluster",
		"https://pve.example.com:8006",
		"root@pam",
		"password",
	)
	_ = repo.Save(ctx, c)

	found, err := repo.FindByName(ctx, "test-cluster")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if found.Name != "test-cluster" {
		t.Errorf("expected name test-cluster, got %s", found.Name)
	}
}

func TestMemoryRepository_List(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := persistence.NewMemoryRepository()

	// Save multiple clusters
	for i := 1; i <= 3; i++ {
		c := cluster.NewCluster(
			"cluster-"+string(rune(i)),
			"cluster-"+string(rune(i)),
			"https://pve.example.com:8006",
			"root@pam",
			"password",
		)
		_ = repo.Save(ctx, c)
	}

	clusters, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(clusters) != 3 {
		t.Errorf("expected 3 clusters, got %d", len(clusters))
	}
}

func TestMemoryRepository_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := persistence.NewMemoryRepository()

	c := cluster.NewCluster(
		"test-id",
		"test-cluster",
		"https://pve.example.com:8006",
		"root@pam",
		"password",
	)
	_ = repo.Save(ctx, c)

	// Delete the cluster
	err := repo.Delete(ctx, "test-id")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify it was deleted
	_, err = repo.FindByID(ctx, "test-id")
	if err == nil {
		t.Fatal("expected error for deleted cluster")
	}
}

func TestMemoryRepository_Exists(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := persistence.NewMemoryRepository()

	c := cluster.NewCluster(
		"test-id",
		"test-cluster",
		"https://pve.example.com:8006",
		"root@pam",
		"password",
	)
	_ = repo.Save(ctx, c)

	exists, err := repo.Exists(ctx, "test-id")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !exists {
		t.Error("expected cluster to exist")
	}

	exists, err = repo.Exists(ctx, "non-existent")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if exists {
		t.Error("expected cluster to not exist")
	}
}
