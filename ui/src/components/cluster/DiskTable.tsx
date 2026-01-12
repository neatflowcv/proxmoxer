import { Table, Badge } from '../common'
import type { NodeDisksResponse, DiskResponse } from '../../api'

interface DiskTableProps {
  nodes: NodeDisksResponse[]
  loading?: boolean
}

function formatSize(bytes: number): string {
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let size = bytes
  let unitIndex = 0

  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024
    unitIndex++
  }

  return `${size.toFixed(1)} ${units[unitIndex]}`
}

function getHealthBadge(health: string) {
  const variant = health === 'PASSED' ? 'success' : health === 'UNKNOWN' ? 'neutral' : 'danger'
  return <Badge variant={variant}>{health || 'Unknown'}</Badge>
}

function getTypeBadge(type: string) {
  const variant = type === 'ssd' || type === 'nvme' ? 'info' : 'neutral'
  return <Badge variant={variant}>{type?.toUpperCase() || 'Unknown'}</Badge>
}

const diskColumns = [
  { key: 'device', header: 'Device' },
  {
    key: 'type',
    header: 'Type',
    render: (value: unknown) => getTypeBadge(value as string),
  },
  {
    key: 'size',
    header: 'Size',
    render: (value: unknown) => formatSize(value as number),
  },
  { key: 'model', header: 'Model' },
  { key: 'vendor', header: 'Vendor' },
  {
    key: 'health',
    header: 'Health',
    render: (value: unknown) => getHealthBadge(value as string),
  },
  {
    key: 'wearout',
    header: 'Wearout',
    render: (value: unknown) => {
      const wearout = value as number
      if (wearout < 0) return 'N/A'
      const variant = wearout > 80 ? 'danger' : wearout > 50 ? 'warning' : 'success'
      return <Badge variant={variant}>{wearout}%</Badge>
    },
  },
  { key: 'used', header: 'Used By' },
]

export function DiskTable({ nodes, loading }: DiskTableProps) {
  if (loading) {
    return (
      <div className="flex justify-center py-12">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600" />
      </div>
    )
  }

  if (nodes.length === 0) {
    return (
      <div className="text-center py-12 text-gray-500">
        No nodes found
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {nodes.map((node) => (
        <div key={node.node_name} className="bg-white rounded-lg shadow border">
          <div className="px-6 py-4 border-b bg-gray-50 flex items-center justify-between">
            <h3 className="font-semibold text-gray-900">{node.node_name}</h3>
            <Badge variant={node.status === 'online' ? 'success' : 'danger'}>
              {node.status}
            </Badge>
          </div>

          {node.error ? (
            <div className="px-6 py-4 text-red-600 text-sm">
              Error: {node.error}
            </div>
          ) : (
            <Table<DiskResponse>
              columns={diskColumns}
              data={node.disks}
              keyExtractor={(disk) => disk.device}
              emptyMessage="No disks found on this node"
            />
          )}
        </div>
      ))}
    </div>
  )
}
