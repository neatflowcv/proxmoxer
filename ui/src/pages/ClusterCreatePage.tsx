import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { registerCluster } from '../api'
import { ClusterForm } from '../components/cluster'
import { Button } from '../components/common'
import type { RegisterClusterRequest, ApiError } from '../api'

export default function ClusterCreatePage() {
  const navigate = useNavigate()
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleSubmit = async (data: RegisterClusterRequest) => {
    setLoading(true)
    setError(null)

    try {
      await registerCluster(data)
      navigate('/')
    } catch (err) {
      const apiError = err as ApiError
      setError(apiError.message || 'Failed to register cluster')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="max-w-xl mx-auto">
      <div className="mb-6">
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
            Back to Dashboard
          </Button>
        </Link>
      </div>

      <ClusterForm onSubmit={handleSubmit} loading={loading} error={error} />
    </div>
  )
}
