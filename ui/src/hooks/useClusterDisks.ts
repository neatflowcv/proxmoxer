import { getClusterDisks } from '../api'
import { useApi } from './useApi'
import type { ClusterDisksResponse } from '../api'

interface UseClusterDisksResult {
  disks: ClusterDisksResponse | null
  loading: boolean
  error: Error | null
  refetch: () => void
}

export function useClusterDisks(id: string): UseClusterDisksResult {
  const { data, loading, error, refetch } = useApi(() => getClusterDisks(id), [id])

  return {
    disks: data,
    loading,
    error,
    refetch,
  }
}
