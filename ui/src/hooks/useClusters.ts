import { listClusters } from '../api'
import { useApi } from './useApi'
import type { ClusterResponse } from '../api'

interface UseClustersResult {
  clusters: ClusterResponse[]
  total: number
  loading: boolean
  error: Error | null
  refetch: () => void
}

export function useClusters(): UseClustersResult {
  const { data, loading, error, refetch } = useApi(() => listClusters(), [])

  return {
    clusters: data?.clusters ?? [],
    total: data?.total ?? 0,
    loading,
    error,
    refetch,
  }
}
