package dto

import "time"

// NodeStatusResponse represents status information for a single node.
type NodeStatusResponse struct {
	// Node name
	NodeName string `json:"node_name"`
	// Node status (online, offline)
	Status string `json:"status"`
	// CPU usage percentage (0-100)
	CPUUsage float64 `json:"cpu_usage"`
	// Memory used in bytes
	MemoryUsed int64 `json:"memory_used"`
	// Total memory in bytes
	MemoryTotal int64 `json:"memory_total"`
	// Memory usage percentage (0-100)
	MemoryUsage float64 `json:"memory_usage"`
	// Swap used in bytes
	SwapUsed int64 `json:"swap_used"`
	// Total swap in bytes
	SwapTotal int64 `json:"swap_total"`
	// Swap usage percentage (0-100)
	SwapUsage float64 `json:"swap_usage"`
	// Uptime in seconds
	Uptime int64 `json:"uptime"`
	// Load average [1min, 5min, 15min]
	LoadAvg []float64 `json:"load_avg"`
	// Error message if status query failed
	Error string `json:"error,omitempty"`
}

// ResourceSummary represents aggregated resource counts for the cluster.
type ResourceSummary struct {
	// Total number of VMs
	TotalVMs int `json:"total_vms"`
	// Number of running VMs
	RunningVMs int `json:"running_vms"`
	// Total number of containers
	TotalContainers int `json:"total_containers"`
	// Number of running containers
	RunningContainers int `json:"running_containers"`
}

// ClusterStatusResponse represents the full cluster monitoring data.
type ClusterStatusResponse struct {
	// Cluster ID
	ClusterID string `json:"cluster_id"`
	// Cluster name
	ClusterName string `json:"cluster_name"`
	// Status of each node
	Nodes []NodeStatusResponse `json:"nodes"`
	// Aggregated resource counts
	ResourceSummary ResourceSummary `json:"resource_summary"`
	// When the data was fetched
	FetchedAt time.Time `json:"fetched_at"`
}
