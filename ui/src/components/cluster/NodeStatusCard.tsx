import { Card, CardBody, Badge, ProgressBar } from '../common'
import type { NodeStatusResponse } from '../../api'

interface NodeStatusCardProps {
  node: NodeStatusResponse
}

function formatBytes(bytes: number): string {
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let unitIndex = 0
  let size = bytes

  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024
    unitIndex++
  }

  return `${size.toFixed(1)} ${units[unitIndex]}`
}

function formatUptime(seconds: number): string {
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)

  if (days > 0) {
    return `${days}d ${hours}h ${minutes}m`
  }
  if (hours > 0) {
    return `${hours}h ${minutes}m`
  }
  return `${minutes}m`
}

export function NodeStatusCard({ node }: NodeStatusCardProps) {
  const isOnline = node.status === 'online'
  const hasError = !!node.error

  return (
    <Card className={hasError ? 'border-red-300' : ''}>
      <CardBody>
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold text-gray-900">{node.node_name}</h3>
          <Badge variant={isOnline ? 'success' : 'danger'}>
            {isOnline ? 'Online' : 'Offline'}
          </Badge>
        </div>

        {hasError ? (
          <div className="text-red-600 text-sm">{node.error}</div>
        ) : (
          <div className="space-y-4">
            <div>
              <ProgressBar
                value={node.cpu_usage}
                label="CPU"
                showPercentage
                size="md"
              />
            </div>

            <div>
              <ProgressBar
                value={node.memory_usage}
                label={`Memory (${formatBytes(node.memory_used)} / ${formatBytes(node.memory_total)})`}
                showPercentage
                size="md"
              />
            </div>

            {node.swap_total > 0 && (
              <div>
                <ProgressBar
                  value={node.swap_usage}
                  label={`Swap (${formatBytes(node.swap_used)} / ${formatBytes(node.swap_total)})`}
                  showPercentage
                  size="md"
                />
              </div>
            )}

            <div className="pt-2 border-t border-gray-200">
              <div className="grid grid-cols-2 gap-2 text-sm text-gray-600">
                <div>
                  <span className="text-gray-500">Uptime:</span>
                  <span className="ml-2 font-medium">{formatUptime(node.uptime)}</span>
                </div>
                {node.load_avg && node.load_avg.length >= 3 && (
                  <div>
                    <span className="text-gray-500">Load:</span>
                    <span className="ml-2 font-medium">
                      {node.load_avg.map(v => v.toFixed(2)).join(' / ')}
                    </span>
                  </div>
                )}
              </div>
            </div>
          </div>
        )}
      </CardBody>
    </Card>
  )
}
