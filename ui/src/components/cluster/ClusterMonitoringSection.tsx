import { Card, CardBody, CardHeader, Spinner, Button } from '../common'
import { NodeStatusCard } from './NodeStatusCard'
import { useClusterStatus } from '../../hooks'
import type { ResourceSummary } from '../../api'

interface ClusterMonitoringSectionProps {
  clusterId: string
}

function ResourceSummaryCard({ summary }: { summary: ResourceSummary }) {
  return (
    <Card>
      <CardHeader>
        <h3 className="text-lg font-semibold">Resource Summary</h3>
      </CardHeader>
      <CardBody>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div className="text-center">
            <div className="text-2xl font-bold text-blue-600">{summary.total_vms}</div>
            <div className="text-sm text-gray-500">Total VMs</div>
          </div>
          <div className="text-center">
            <div className="text-2xl font-bold text-green-600">{summary.running_vms}</div>
            <div className="text-sm text-gray-500">Running VMs</div>
          </div>
          <div className="text-center">
            <div className="text-2xl font-bold text-blue-600">{summary.total_containers}</div>
            <div className="text-sm text-gray-500">Total Containers</div>
          </div>
          <div className="text-center">
            <div className="text-2xl font-bold text-green-600">{summary.running_containers}</div>
            <div className="text-sm text-gray-500">Running Containers</div>
          </div>
        </div>
      </CardBody>
    </Card>
  )
}

function formatTime(date: Date): string {
  return date.toLocaleTimeString()
}

export function ClusterMonitoringSection({ clusterId }: ClusterMonitoringSectionProps) {
  const { status, loading, error, refetch, lastUpdated } = useClusterStatus(clusterId)

  if (loading && !status) {
    return (
      <div className="flex justify-center items-center py-12">
        <Spinner size="lg" />
      </div>
    )
  }

  if (error && !status) {
    return (
      <Card>
        <CardBody>
          <div className="text-center py-8">
            <p className="text-red-600 mb-4">Failed to load monitoring data: {error.message}</p>
            <Button onClick={refetch}>Retry</Button>
          </div>
        </CardBody>
      </Card>
    )
  }

  if (!status) {
    return null
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold text-gray-900">Cluster Monitoring</h2>
        <div className="flex items-center gap-4">
          {lastUpdated && (
            <span className="text-sm text-gray-500">
              Last updated: {formatTime(lastUpdated)}
            </span>
          )}
          <Button variant="secondary" size="sm" onClick={refetch} disabled={loading}>
            {loading ? 'Refreshing...' : 'Refresh'}
          </Button>
        </div>
      </div>

      <ResourceSummaryCard summary={status.resource_summary} />

      <div>
        <h3 className="text-lg font-semibold text-gray-900 mb-4">Node Status</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {status.nodes.map((node) => (
            <NodeStatusCard key={node.node_name} node={node} />
          ))}
        </div>
      </div>
    </div>
  )
}
