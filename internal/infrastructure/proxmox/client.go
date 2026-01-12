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

	"github.com/neatflowcv/proxmoxer/internal/domain/common"
)

// Client represents a Proxmox API client.
type Client struct {
	baseURL    string
	httpClient *http.Client
	timeout    time.Duration
}

// NewClient creates a new Proxmox API client
// insecureSkipVerify should be true for self-signed certificates (testing/development only).
func NewClient(baseURL string, timeout time.Duration, insecureSkipVerify bool) *Client {
	const defaultTimeout = 30 * time.Second
	if timeout == 0 {
		timeout = defaultTimeout
	}

	tlsConfig := createTLSConfig(insecureSkipVerify)
	transport := createHTTPTransport(tlsConfig)

	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Transport:     transport,
			CheckRedirect: nil,
			Jar:           nil,
			Timeout:       timeout,
		},
		timeout: timeout,
	}
}

// createTLSConfig creates a TLS configuration.
// InsecureSkipVerify is intentional for self-signed certificates in development.
func createTLSConfig(insecureSkipVerify bool) *tls.Config {
	return &tls.Config{
		Rand:                                nil,
		Time:                                nil,
		Certificates:                        nil,
		NameToCertificate:                   nil,
		GetCertificate:                      nil,
		GetClientCertificate:                nil,
		GetConfigForClient:                  nil,
		VerifyPeerCertificate:               nil,
		VerifyConnection:                    nil,
		RootCAs:                             nil,
		NextProtos:                          nil,
		ServerName:                          "",
		ClientAuth:                          0,
		ClientCAs:                           nil,
		InsecureSkipVerify:                  insecureSkipVerify, //nolint:gosec
		CipherSuites:                        nil,
		PreferServerCipherSuites:            false,
		SessionTicketsDisabled:              false,
		SessionTicketKey:                    [32]byte{},
		ClientSessionCache:                  nil,
		UnwrapSession:                       nil,
		WrapSession:                         nil,
		MinVersion:                          0,
		MaxVersion:                          0,
		CurvePreferences:                    nil,
		DynamicRecordSizingDisabled:         false,
		Renegotiation:                       0,
		KeyLogWriter:                        nil,
		EncryptedClientHelloConfigList:      nil,
		EncryptedClientHelloRejectionVerify: nil,
		GetEncryptedClientHelloKeys:         nil,
		EncryptedClientHelloKeys:            nil,
	}
}

// createHTTPTransport creates an HTTP transport with the given TLS configuration.
func createHTTPTransport(tlsConfig *tls.Config) *http.Transport {
	return &http.Transport{
		Proxy:                  nil,
		OnProxyConnectResponse: nil,
		DialContext:            nil,
		Dial:                   nil,
		DialTLSContext:         nil,
		DialTLS:                nil,
		TLSClientConfig:        tlsConfig,
		TLSHandshakeTimeout:    0,
		DisableKeepAlives:      false,
		DisableCompression:     false,
		MaxIdleConns:           0,
		MaxIdleConnsPerHost:    0,
		MaxConnsPerHost:        0,
		IdleConnTimeout:        0,
		ResponseHeaderTimeout:  0,
		ExpectContinueTimeout:  0,
		TLSNextProto:           nil,
		ProxyConnectHeader:     nil,
		GetProxyConnectHeader:  nil,
		MaxResponseHeaderBytes: 0,
		WriteBufferSize:        0,
		ReadBufferSize:         0,
		ForceAttemptHTTP2:      false,
		HTTP2:                  nil,
		Protocols:              nil,
	}
}

// AuthenticateResponse represents the response from Proxmox authentication.
type AuthenticateResponse struct {
	Data struct {
		Ticket string `json:"ticket"`
		CSRF   string `json:"csrf"`
	} `json:"data"`
	RequestID string `json:"requestid"`
}

// NodeInfo represents basic node information.
type NodeInfo struct {
	Node   string `json:"node"`
	Status string `json:"status"`
}

// ListNodesResponse represents the response from Proxmox nodes endpoint.
type ListNodesResponse struct {
	Data      []NodeInfo `json:"data"`
	RequestID string     `json:"requestid"`
}

// DiskInfo represents disk information from Proxmox API.
type DiskInfo struct {
	DevPath string `json:"devpath"`
	Type    string `json:"type"`
	Size    int64  `json:"size"`
	Model   string `json:"model"`
	Serial  string `json:"serial"`
	Vendor  string `json:"vendor"`
	Wearout any    `json:"wearout"`
	Health  string `json:"health"`
	Used    string `json:"used"`
	GPT     int    `json:"gpt"`
}

// ListDisksResponse represents the response from Proxmox disks endpoint.
type ListDisksResponse struct {
	Data []DiskInfo `json:"data"`
}

// GetVersionResponse represents the response from Proxmox version endpoint.
type GetVersionResponse struct {
	Data struct {
		Release string `json:"release"`
		Version string `json:"version"`
	} `json:"data"`
	RequestID string `json:"requestid"`
}

// NodeStatusData represents the status data of a single node.
type NodeStatusData struct {
	CPU     float64 `json:"cpu"`
	Memory  MemoryStatus
	Swap    SwapStatus
	Uptime  int64     `json:"uptime"`
	LoadAvg []float64 `json:"loadavg"`
}

// MemoryStatus represents memory information.
type MemoryStatus struct {
	Used  int64 `json:"used"`
	Total int64 `json:"total"`
	Free  int64 `json:"free"`
}

// SwapStatus represents swap information.
type SwapStatus struct {
	Used  int64 `json:"used"`
	Total int64 `json:"total"`
	Free  int64 `json:"free"`
}

// nodeStatusAPIResponse represents the raw API response for node status.
type nodeStatusAPIResponse struct {
	Data struct {
		CPU     float64 `json:"cpu"`
		Memory  struct {
			Used  int64 `json:"used"`
			Total int64 `json:"total"`
			Free  int64 `json:"free"`
		} `json:"memory"`
		Swap struct {
			Used  int64 `json:"used"`
			Total int64 `json:"total"`
			Free  int64 `json:"free"`
		} `json:"swap"`
		Uptime  int64     `json:"uptime"`
		LoadAvg []float64 `json:"loadavg"`
	} `json:"data"`
}

// ClusterResource represents a resource from the cluster resources endpoint.
type ClusterResource struct {
	ID      string  `json:"id"`
	Node    string  `json:"node"`
	Type    string  `json:"type"`
	Status  string  `json:"status"`
	CPU     float64 `json:"cpu"`
	MaxCPU  int     `json:"maxcpu"`
	Mem     int64   `json:"mem"`
	MaxMem  int64   `json:"maxmem"`
	Disk    int64   `json:"disk"`
	MaxDisk int64   `json:"maxdisk"`
	Uptime  int64   `json:"uptime"`
	Name    string  `json:"name"`
}

// clusterResourcesResponse represents the response from cluster resources endpoint.
type clusterResourcesResponse struct {
	Data []ClusterResource `json:"data"`
}

// Authenticate authenticates with the Proxmox API and returns ticket and CSRF token.
// This validates the credentials by attempting an actual API call.
func (c *Client) Authenticate(ctx context.Context, username, password string) (string, string, error) {
	authURL := c.baseURL + "/api2/json/access/ticket"

	// Prepare form data
	data := url.Values{}
	data.Set("username", username)
	data.Set("password", password)

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, authURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return "", "", fmt.Errorf("failed to create authentication request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("authentication request failed: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to read authentication response: %w", err)
	}

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf(
			"authentication failed with status %d: %s: %w",
			resp.StatusCode, string(body), common.ErrAuthenticationFailed)
	}

	// Parse response
	var authResp AuthenticateResponse

	err = json.Unmarshal(body, &authResp)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse authentication response: %w", err)
	}

	if authResp.Data.Ticket == "" {
		return "", "", common.ErrNoAuthenticationTicket
	}

	return authResp.Data.Ticket, authResp.Data.CSRF, nil
}

// GetVersion retrieves the Proxmox version.
func (c *Client) GetVersion(ctx context.Context, ticket string) (string, error) {
	versionURL := c.baseURL + "/api2/json/version"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, versionURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create version request: %w", err)
	}

	c.setAuthHeaders(req, ticket)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("version request failed: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read version response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get version with status %d: %w", resp.StatusCode, common.ErrProxmoxConnectionFailed)
	}

	var versionResp GetVersionResponse

	err = json.Unmarshal(body, &versionResp)
	if err != nil {
		return "", fmt.Errorf("failed to parse version response: %w", err)
	}

	return versionResp.Data.Version, nil
}

// GetNodeCount retrieves the number of nodes in the cluster.
func (c *Client) GetNodeCount(ctx context.Context, ticket string) (int, error) {
	nodesURL := c.baseURL + "/api2/json/nodes"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, nodesURL, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create nodes request: %w", err)
	}

	c.setAuthHeaders(req, ticket)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("nodes request failed: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read nodes response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to get nodes with status %d: %w", resp.StatusCode, common.ErrProxmoxConnectionFailed)
	}

	var nodesResp ListNodesResponse

	err = json.Unmarshal(body, &nodesResp)
	if err != nil {
		return 0, fmt.Errorf("failed to parse nodes response: %w", err)
	}

	return len(nodesResp.Data), nil
}

// ListNodes retrieves the list of nodes in the cluster.
func (c *Client) ListNodes(ctx context.Context, ticket string) ([]NodeInfo, error) {
	nodesURL := c.baseURL + "/api2/json/nodes"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, nodesURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create nodes request: %w", err)
	}

	c.setAuthHeaders(req, ticket)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("nodes request failed: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read nodes response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get nodes with status %d: %w", resp.StatusCode, common.ErrProxmoxConnectionFailed)
	}

	var nodesResp ListNodesResponse

	err = json.Unmarshal(body, &nodesResp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse nodes response: %w", err)
	}

	return nodesResp.Data, nil
}

// ListNodeDisks retrieves disk information for a specific node.
func (c *Client) ListNodeDisks(ctx context.Context, ticket string, nodeName string) ([]DiskInfo, error) {
	disksURL := fmt.Sprintf("%s/api2/json/nodes/%s/disks/list", c.baseURL, nodeName)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, disksURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create disks request: %w", err)
	}

	c.setAuthHeaders(req, ticket)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("disks request failed: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read disks response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get disks with status %d: %w", resp.StatusCode, common.ErrDiskQueryFailed)
	}

	var disksResp ListDisksResponse

	err = json.Unmarshal(body, &disksResp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse disks response: %w", err)
	}

	return disksResp.Data, nil
}

// setAuthHeaders sets the authentication headers for API requests using Cookie-based authentication.
func (c *Client) setAuthHeaders(req *http.Request, ticket string) {
	req.Header.Set("Cookie", "PVEAuthCookie="+ticket)
}

// GetNodeStatus retrieves status information for a specific node.
func (c *Client) GetNodeStatus(ctx context.Context, ticket string, nodeName string) (*NodeStatusData, error) {
	statusURL := fmt.Sprintf("%s/api2/json/nodes/%s/status", c.baseURL, nodeName)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, statusURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create node status request: %w", err)
	}

	c.setAuthHeaders(req, ticket)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("node status request failed: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read node status response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"failed to get node status with status %d: %w", resp.StatusCode, common.ErrProxmoxConnectionFailed)
	}

	var statusResp nodeStatusAPIResponse

	err = json.Unmarshal(body, &statusResp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse node status response: %w", err)
	}

	return &NodeStatusData{
		CPU: statusResp.Data.CPU,
		Memory: MemoryStatus{
			Used:  statusResp.Data.Memory.Used,
			Total: statusResp.Data.Memory.Total,
			Free:  statusResp.Data.Memory.Free,
		},
		Swap: SwapStatus{
			Used:  statusResp.Data.Swap.Used,
			Total: statusResp.Data.Swap.Total,
			Free:  statusResp.Data.Swap.Free,
		},
		Uptime:  statusResp.Data.Uptime,
		LoadAvg: statusResp.Data.LoadAvg,
	}, nil
}

// GetClusterResources retrieves all cluster resources (nodes, VMs, containers, storage).
func (c *Client) GetClusterResources(ctx context.Context, ticket string) ([]ClusterResource, error) {
	resourcesURL := c.baseURL + "/api2/json/cluster/resources"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, resourcesURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cluster resources request: %w", err)
	}

	c.setAuthHeaders(req, ticket)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cluster resources request failed: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read cluster resources response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"failed to get cluster resources with status %d: %w", resp.StatusCode, common.ErrProxmoxConnectionFailed)
	}

	var resourcesResp clusterResourcesResponse

	err = json.Unmarshal(body, &resourcesResp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse cluster resources response: %w", err)
	}

	return resourcesResp.Data, nil
}
