import { getCluster } from '../api'
import { useApi } from './useApi'
import type { ClusterResponse } from '../api'

interface UseClusterResult {
  cluster: ClusterResponse | null
  loading: boolean
  error: Error | null
  refetch: () => void
}

export function useCluster(id: string): UseClusterResult {
  const { data, loading, error, refetch } = useApi(() => getCluster(id), [id])

  return {
    cluster: data,
    loading,
    error,
    refetch,
  }
}
