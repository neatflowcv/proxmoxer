import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useClusters } from '../hooks'
import { deleteCluster } from '../api'
import { ClusterCard } from '../components/cluster'
import { Button, Spinner, Modal } from '../components/common'

export default function DashboardPage() {
  const { clusters, loading, error, refetch } = useClusters()
  const [deleteId, setDeleteId] = useState<string | null>(null)
  const [deleting, setDeleting] = useState(false)

  const handleDelete = async () => {
    if (!deleteId) return

    setDeleting(true)
    try {
      await deleteCluster(deleteId)
      refetch()
    } catch (err) {
      console.error('Failed to delete cluster:', err)
    } finally {
      setDeleting(false)
      setDeleteId(null)
    }
  }

  if (loading) {
    return (
      <div className="flex justify-center items-center py-24">
        <Spinner size="lg" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="bg-red-50 text-red-700 px-6 py-4 rounded-lg">
        <h2 className="font-semibold">Error loading clusters</h2>
        <p className="mt-1 text-sm">{error.message}</p>
        <Button variant="secondary" onClick={refetch} className="mt-4">
          Retry
        </Button>
      </div>
    )
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Clusters</h1>
        <Link to="/clusters/new">
          <Button>Add Cluster</Button>
        </Link>
      </div>

      {clusters.length === 0 ? (
        <div className="bg-white rounded-lg shadow border border-gray-200 p-12 text-center">
          <svg
            className="mx-auto h-12 w-12 text-gray-400"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={1}
              d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2"
            />
          </svg>
          <h3 className="mt-4 text-lg font-medium text-gray-900">No clusters</h3>
          <p className="mt-2 text-gray-500">Get started by adding your first Proxmox cluster.</p>
          <Link to="/clusters/new" className="mt-6 inline-block">
            <Button>Add Cluster</Button>
          </Link>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {clusters.map((cluster) => (
            <ClusterCard
              key={cluster.id}
              cluster={cluster}
              onDelete={setDeleteId}
            />
          ))}
        </div>
      )}

      <Modal
        isOpen={!!deleteId}
        onClose={() => setDeleteId(null)}
        title="Delete Cluster"
        footer={
          <>
            <Button variant="ghost" onClick={() => setDeleteId(null)}>
              Cancel
            </Button>
            <Button variant="danger" loading={deleting} onClick={handleDelete}>
              Delete
            </Button>
          </>
        }
      >
        <p className="text-gray-600">
          Are you sure you want to delete this cluster? This action cannot be undone.
        </p>
      </Modal>
    </div>
  )
}
