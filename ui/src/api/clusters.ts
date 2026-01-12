import { request } from './client'
import type {
  RegisterClusterRequest,
  ClusterResponse,
  ListClustersResponse,
  ClusterDisksResponse,
} from './types'

export async function listClusters(): Promise<ListClustersResponse> {
  return request<ListClustersResponse>('/api/v1/clusters')
}

export async function registerCluster(data: RegisterClusterRequest): Promise<ClusterResponse> {
  return request<ClusterResponse>('/api/v1/clusters', {
    method: 'POST',
    body: data,
  })
}

export async function getCluster(id: string): Promise<ClusterResponse> {
  return request<ClusterResponse>(`/api/v1/clusters/${id}`)
}

export async function deleteCluster(id: string): Promise<void> {
  return request<void>(`/api/v1/clusters/${id}`, {
    method: 'DELETE',
  })
}

export async function getClusterDisks(id: string): Promise<ClusterDisksResponse> {
  return request<ClusterDisksResponse>(`/api/v1/clusters/${id}/disks`)
}
