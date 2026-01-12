export interface RegisterClusterRequest {
  name: string
  api_endpoint: string
  username: string
  password: string
}

export interface ClusterResponse {
  id: string
  name: string
  api_endpoint: string
  status: string
  proxmox_version: string
  node_count: number
  created_at: string
  updated_at: string
}

export interface ListClustersResponse {
  clusters: ClusterResponse[]
  total: number
}

export interface DiskResponse {
  device: string
  type: string
  size: number
  model: string
  serial: string
  vendor: string
  wearout: number
  health: string
  used: string
}

export interface NodeDisksResponse {
  node_name: string
  status: string
  disks: DiskResponse[]
  error?: string
}

export interface ClusterDisksResponse {
  cluster_id: string
  cluster_name: string
  nodes: NodeDisksResponse[]
  total_disks: number
}

export interface ErrorResponse {
  code: string
  message: string
  details?: Record<string, unknown>
}

export interface NodeStatusResponse {
  node_name: string
  status: string
  cpu_usage: number
  memory_used: number
  memory_total: number
  memory_usage: number
  swap_used: number
  swap_total: number
  swap_usage: number
  uptime: number
  load_avg: number[]
  error?: string
}

export interface ResourceSummary {
  total_vms: number
  running_vms: number
  total_containers: number
  running_containers: number
}

export interface ClusterStatusResponse {
  cluster_id: string
  cluster_name: string
  nodes: NodeStatusResponse[]
  resource_summary: ResourceSummary
  fetched_at: string
}
