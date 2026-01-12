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
