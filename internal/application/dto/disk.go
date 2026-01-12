package dto

// DiskResponse represents a single disk in the response.
type DiskResponse struct {
	// Device path (e.g., /dev/sda)
	Device string `json:"device"`
	// Disk type (hdd, ssd, nvme)
	Type string `json:"type"`
	// Size in bytes
	Size int64 `json:"size"`
	// Disk model
	Model string `json:"model"`
	// Serial number
	Serial string `json:"serial"`
	// Vendor name
	Vendor string `json:"vendor"`
	// SSD wear level (percentage, -1 for HDD)
	Wearout int `json:"wearout"`
	// S.M.A.R.T. health status
	Health string `json:"health"`
	// Usage type (LVM, ZFS, filesystem, etc.)
	Used string `json:"used"`
}

// NodeDisksResponse represents disks for a single node.
type NodeDisksResponse struct {
	// Node name
	NodeName string `json:"node_name"`
	// Node status (online, offline)
	Status string `json:"status"`
	// List of disks
	Disks []DiskResponse `json:"disks"`
	// Error message if disk query failed for this node
	Error string `json:"error,omitempty"`
}

// ClusterDisksResponse represents the full cluster disk information.
type ClusterDisksResponse struct {
	// Cluster ID
	ClusterID string `json:"cluster_id"`
	// Cluster name
	ClusterName string `json:"cluster_name"`
	// List of nodes with their disks
	Nodes []NodeDisksResponse `json:"nodes"`
	// Total number of disks across all nodes
	TotalDisks int `json:"total_disks"`
}
