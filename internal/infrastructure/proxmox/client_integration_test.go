//go:build integration

package proxmox

import (
	"context"
	"testing"
	"time"
)

const (
	// Real Proxmox server configuration
	proxmoxURL      = "https://192.168.122.127:8006"
	proxmoxUsername = "root@pam"
	proxmoxPassword = "rootroot"
	testTimeout     = 15 * time.Second
)

// TestProxmoxClient_Authenticate_Integration tests authentication with the real Proxmox server
func TestProxmoxClient_Authenticate_Integration(t *testing.T) {
	client := NewClient(proxmoxURL, testTimeout, true)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	ticket, csrf, err := client.Authenticate(ctx, proxmoxUsername, proxmoxPassword)
	if err != nil {
		t.Fatalf("Authentication failed: %v", err)
	}

	if ticket == "" {
		t.Fatal("Authentication succeeded but ticket is empty")
	}

	t.Logf("Successfully authenticated. Ticket length: %d, CSRF length: %d", len(ticket), len(csrf))
}

// TestProxmoxClient_GetVersion_Integration tests retrieving version from the real Proxmox server
func TestProxmoxClient_GetVersion_Integration(t *testing.T) {
	client := NewClient(proxmoxURL, testTimeout, true)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	// First authenticate to get ticket
	ticket, _, err := client.Authenticate(ctx, proxmoxUsername, proxmoxPassword)
	if err != nil {
		t.Fatalf("Authentication failed: %v", err)
	}

	// Then get version
	version, err := client.GetVersion(ctx, ticket)
	if err != nil {
		t.Fatalf("GetVersion failed: %v", err)
	}

	if version == "" {
		t.Fatal("GetVersion succeeded but version is empty")
	}

	t.Logf("Successfully retrieved version: %s", version)
}

// TestProxmoxClient_GetNodeCount_Integration tests retrieving node count from the real Proxmox server
func TestProxmoxClient_GetNodeCount_Integration(t *testing.T) {
	client := NewClient(proxmoxURL, testTimeout, true)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	// First authenticate to get ticket
	ticket, _, err := client.Authenticate(ctx, proxmoxUsername, proxmoxPassword)
	if err != nil {
		t.Fatalf("Authentication failed: %v", err)
	}

	// Then get node count
	count, err := client.GetNodeCount(ctx, ticket)
	if err != nil {
		t.Fatalf("GetNodeCount failed: %v", err)
	}

	if count < 0 {
		t.Fatalf("GetNodeCount returned invalid count: %d", count)
	}

	t.Logf("Successfully retrieved node count: %d", count)
}

// TestProxmoxClient_InvalidCredentials_Integration tests that authentication fails with invalid credentials
func TestProxmoxClient_InvalidCredentials_Integration(t *testing.T) {
	client := NewClient(proxmoxURL, testTimeout, true)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	_, _, err := client.Authenticate(ctx, proxmoxUsername, "invalidpassword")
	if err == nil {
		t.Fatal("Expected authentication to fail with invalid password, but it succeeded")
	}

	t.Logf("Authentication correctly failed with invalid credentials: %v", err)
}

// TestProxmoxClient_AuthenticationFlow_Integration tests the complete authentication flow
func TestProxmoxClient_AuthenticationFlow_Integration(t *testing.T) {
	client := NewClient(proxmoxURL, testTimeout, true)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	// Step 1: Authenticate
	ticket, csrf, err := client.Authenticate(ctx, proxmoxUsername, proxmoxPassword)
	if err != nil {
		t.Fatalf("Step 1 - Authentication failed: %v", err)
	}

	if ticket == "" {
		t.Fatal("Step 1 - No ticket received")
	}

	t.Logf("Step 1 - Authentication successful. Ticket length: %d, CSRF length: %d",
		len(ticket), len(csrf))

	// Step 2: Get version using ticket
	version, err := client.GetVersion(ctx, ticket)
	if err != nil {
		t.Fatalf("Step 2 - GetVersion failed: %v", err)
	}

	if version == "" {
		t.Fatal("Step 2 - Version is empty")
	}

	t.Logf("Step 2 - Retrieved version: %s", version)

	// Step 3: Get node count using ticket
	count, err := client.GetNodeCount(ctx, ticket)
	if err != nil {
		t.Fatalf("Step 3 - GetNodeCount failed: %v", err)
	}

	t.Logf("Step 3 - Retrieved node count: %d", count)

	t.Log("Complete authentication flow test passed successfully")
}
