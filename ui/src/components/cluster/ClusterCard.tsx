import { Link } from 'react-router-dom'
import { Card, CardBody, Button } from '../common'
import { ClusterStatusBadge } from './ClusterStatusBadge'
import type { ClusterResponse } from '../../api'

interface ClusterCardProps {
  cluster: ClusterResponse
  onDelete: (id: string) => void
}

export function ClusterCard({ cluster, onDelete }: ClusterCardProps) {
  return (
    <Card>
      <CardBody>
        <div className="flex items-start justify-between">
          <div>
            <h3 className="text-lg font-semibold text-gray-900">{cluster.name}</h3>
            <p className="text-sm text-gray-500 mt-1">{cluster.api_endpoint}</p>
          </div>
          <ClusterStatusBadge status={cluster.status} />
        </div>

        <div className="mt-4 grid grid-cols-2 gap-4 text-sm">
          <div>
            <span className="text-gray-500">Nodes:</span>
            <span className="ml-2 font-medium">{cluster.node_count}</span>
          </div>
          <div>
            <span className="text-gray-500">Version:</span>
            <span className="ml-2 font-medium">{cluster.proxmox_version || 'N/A'}</span>
          </div>
        </div>

        <div className="mt-6 flex gap-3">
          <Link to={`/clusters/${cluster.id}`} className="flex-1">
            <Button variant="secondary" className="w-full">
              View Details
            </Button>
          </Link>
          <Button variant="danger" onClick={() => onDelete(cluster.id)}>
            Delete
          </Button>
        </div>
      </CardBody>
    </Card>
  )
}
