import { useState, useEffect, useCallback, useRef } from 'react'
import { getClusterStatus } from '../api'
import type { ClusterStatusResponse } from '../api'

interface UseClusterStatusResult {
  status: ClusterStatusResponse | null
  loading: boolean
  error: Error | null
  refetch: () => void
  lastUpdated: Date | null
}

const DEFAULT_REFRESH_INTERVAL = 30000 // 30 seconds

export function useClusterStatus(
  id: string,
  refreshInterval: number = DEFAULT_REFRESH_INTERVAL
): UseClusterStatusResult {
  const [status, setStatus] = useState<ClusterStatusResponse | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<Error | null>(null)
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null)
  const intervalRef = useRef<number | null>(null)

  const fetchData = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const data = await getClusterStatus(id)
      setStatus(data)
      setLastUpdated(new Date())
    } catch (err) {
      setError(err as Error)
    } finally {
      setLoading(false)
    }
  }, [id])

  useEffect(() => {
    fetchData()

    if (refreshInterval > 0) {
      intervalRef.current = window.setInterval(fetchData, refreshInterval)
    }

    return () => {
      if (intervalRef.current !== null) {
        window.clearInterval(intervalRef.current)
      }
    }
  }, [fetchData, refreshInterval])

  return {
    status,
    loading,
    error,
    refetch: fetchData,
    lastUpdated,
  }
}
