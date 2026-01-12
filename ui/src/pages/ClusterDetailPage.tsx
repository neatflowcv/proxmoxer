import { useState } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { useCluster, useClusterDisks } from '../hooks'
import { deleteCluster } from '../api'
import { ClusterStatusBadge, DiskTable, ClusterMonitoringSection } from '../components/cluster'
import { Button, Card, CardBody, CardHeader, Spinner, Modal } from '../components/common'

export default function ClusterDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { cluster, loading: clusterLoading, error: clusterError } = useCluster(id!)
  const { disks, loading: disksLoading } = useClusterDisks(id!)
  const [showDeleteModal, setShowDeleteModal] = useState(false)
  const [deleting, setDeleting] = useState(false)

  const handleDelete = async () => {
    if (!id) return

    setDeleting(true)
    try {
      await deleteCluster(id)
      navigate('/')
    } catch (err) {
      console.error('Failed to delete cluster:', err)
    } finally {
      setDeleting(false)
      setShowDeleteModal(false)
    }
  }

  if (clusterLoading) {
    return (
      <div className="flex justify-center items-center py-24">
        <Spinner size="lg" />
      </div>
    )
  }

  if (clusterError || !cluster) {
    return (
      <div className="bg-red-50 text-red-700 px-6 py-4 rounded-lg">
        <h2 className="font-semibold">Error loading cluster</h2>
        <p className="mt-1 text-sm">{clusterError?.message || 'Cluster not found'}</p>
        <Link to="/" className="mt-4 inline-block">
          <Button variant="secondary">Back to Dashboard</Button>
        </Link>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Link to="/">
            <Button variant="ghost" className="gap-2">
              <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M10 19l-7-7m0 0l7-7m-7 7h18"
                />
              </svg>
              Back
            </Button>
          </Link>
          <h1 className="text-2xl font-bold text-gray-900">{cluster.name}</h1>
          <ClusterStatusBadge status={cluster.status} />
        </div>
        <Button variant="danger" onClick={() => setShowDeleteModal(true)}>
          Delete Cluster
        </Button>
      </div>

      <Card>
        <CardHeader>
          <h2 className="text-lg font-semibold">Cluster Information</h2>
        </CardHeader>
        <CardBody>
          <dl className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <dt className="text-sm text-gray-500">API Endpoint</dt>
              <dd className="mt-1 font-medium">{cluster.api_endpoint}</dd>
            </div>
            <div>
              <dt className="text-sm text-gray-500">Proxmox Version</dt>
              <dd className="mt-1 font-medium">{cluster.proxmox_version || 'N/A'}</dd>
            </div>
            <div>
              <dt className="text-sm text-gray-500">Node Count</dt>
              <dd className="mt-1 font-medium">{cluster.node_count}</dd>
            </div>
            <div>
              <dt className="text-sm text-gray-500">Created</dt>
              <dd className="mt-1 font-medium">
                {new Date(cluster.created_at).toLocaleDateString()}
              </dd>
            </div>
          </dl>
        </CardBody>
      </Card>

      <ClusterMonitoringSection clusterId={id!} />

      <Card>
        <CardHeader>
          <h2 className="text-lg font-semibold">Disk Information</h2>
        </CardHeader>
        <CardBody className="p-0">
          <DiskTable nodes={disks?.nodes ?? []} loading={disksLoading} />
        </CardBody>
      </Card>

      <Modal
        isOpen={showDeleteModal}
        onClose={() => setShowDeleteModal(false)}
        title="Delete Cluster"
        footer={
          <>
            <Button variant="ghost" onClick={() => setShowDeleteModal(false)}>
              Cancel
            </Button>
            <Button variant="danger" loading={deleting} onClick={handleDelete}>
              Delete
            </Button>
          </>
        }
      >
        <p className="text-gray-600">
          Are you sure you want to delete <strong>{cluster.name}</strong>? This action cannot be undone.
        </p>
      </Modal>
    </div>
  )
}
