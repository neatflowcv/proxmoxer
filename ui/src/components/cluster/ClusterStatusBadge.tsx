import { Badge } from '../common'

interface ClusterStatusBadgeProps {
  status: string
}

const statusConfig: Record<string, { variant: 'success' | 'warning' | 'danger' | 'neutral'; label: string }> = {
  healthy: { variant: 'success', label: 'Healthy' },
  degraded: { variant: 'warning', label: 'Degraded' },
  unhealthy: { variant: 'danger', label: 'Unhealthy' },
  unknown: { variant: 'neutral', label: 'Unknown' },
}

export function ClusterStatusBadge({ status }: ClusterStatusBadgeProps) {
  const config = statusConfig[status] || statusConfig.unknown

  return <Badge variant={config.variant}>{config.label}</Badge>
}
