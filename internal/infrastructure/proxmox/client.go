package proxmox

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client represents a Proxmox API client
type Client struct {
	baseURL    string
	httpClient *http.Client
	timeout    time.Duration
}

// NewClient creates a new Proxmox API client
// insecureSkipVerify should be true for self-signed certificates (testing/development only)
func NewClient(baseURL string, timeout time.Duration, insecureSkipVerify bool) *Client {
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	// Create TLS configuration
	tlsConfig := &tls.Config{
		InsecureSkipVerify: insecureSkipVerify,
	}

	// Create HTTP transport with TLS config
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
		timeout: timeout,
	}
}

// AuthenticateResponse represents the response from Proxmox authentication
type AuthenticateResponse struct {
	Data struct {
		Ticket string `json:"ticket"`
		CSRF   string `json:"csrf"`
	} `json:"data"`
	RequestID string `json:"requestid"`
}

// GetNodesResponse represents the response from Proxmox nodes endpoint
type GetNodesResponse struct {
	Data []struct {
		Node   string `json:"node"`
		Status string `json:"status"`
	} `json:"data"`
	RequestID string `json:"requestid"`
}

// GetVersionResponse represents the response from Proxmox version endpoint
type GetVersionResponse struct {
	Data struct {
		Release string `json:"release"`
		Version string `json:"version"`
	} `json:"data"`
	RequestID string `json:"requestid"`
}

// Authenticate authenticates with the Proxmox API and returns ticket and CSRF token
// This validates the credentials by attempting an actual API call
func (c *Client) Authenticate(ctx context.Context, username string, password string) (ticket string, csrf string, err error) {
	authURL := fmt.Sprintf("%s/api2/json/access/ticket", c.baseURL)

	// Prepare form data
	data := url.Values{}
	data.Set("username", username)
	data.Set("password", password)

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", authURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return "", "", fmt.Errorf("failed to create authentication request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("authentication request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to read authentication response: %w", err)
	}

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("authentication failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var authResp AuthenticateResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		return "", "", fmt.Errorf("failed to parse authentication response: %w", err)
	}

	if authResp.Data.Ticket == "" {
		return "", "", fmt.Errorf("no authentication ticket received")
	}

	return authResp.Data.Ticket, authResp.Data.CSRF, nil
}

// GetVersion retrieves the Proxmox version
func (c *Client) GetVersion(ctx context.Context, ticket string) (version string, err error) {
	versionURL := fmt.Sprintf("%s/api2/json/version", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", versionURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create version request: %w", err)
	}

	c.setAuthHeaders(req, ticket)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("version request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read version response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get version with status %d", resp.StatusCode)
	}

	var versionResp GetVersionResponse
	if err := json.Unmarshal(body, &versionResp); err != nil {
		return "", fmt.Errorf("failed to parse version response: %w", err)
	}

	return versionResp.Data.Version, nil
}

// GetNodeCount retrieves the number of nodes in the cluster
func (c *Client) GetNodeCount(ctx context.Context, ticket string) (count int, err error) {
	nodesURL := fmt.Sprintf("%s/api2/json/nodes", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", nodesURL, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create nodes request: %w", err)
	}

	c.setAuthHeaders(req, ticket)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("nodes request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read nodes response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to get nodes with status %d", resp.StatusCode)
	}

	var nodesResp GetNodesResponse
	if err := json.Unmarshal(body, &nodesResp); err != nil {
		return 0, fmt.Errorf("failed to parse nodes response: %w", err)
	}

	return len(nodesResp.Data), nil
}

// setAuthHeaders sets the authentication headers for API requests using Cookie-based authentication
func (c *Client) setAuthHeaders(req *http.Request, ticket string) {
	req.Header.Set("Cookie", fmt.Sprintf("PVEAuthCookie=%s", ticket))
}
