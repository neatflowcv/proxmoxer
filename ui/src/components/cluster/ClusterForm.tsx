import { useState } from 'react'
import { Button, Input, Card, CardBody, CardHeader } from '../common'
import type { RegisterClusterRequest } from '../../api'

interface ClusterFormProps {
  onSubmit: (data: RegisterClusterRequest) => Promise<void>
  loading?: boolean
  error?: string | null
}

export function ClusterForm({ onSubmit, loading = false, error }: ClusterFormProps) {
  const [formData, setFormData] = useState<RegisterClusterRequest>({
    name: '',
    api_endpoint: '',
    username: '',
    password: '',
  })

  const [errors, setErrors] = useState<Partial<Record<keyof RegisterClusterRequest, string>>>({})

  const validate = (): boolean => {
    const newErrors: Partial<Record<keyof RegisterClusterRequest, string>> = {}

    if (!formData.name.trim()) {
      newErrors.name = 'Name is required'
    }

    if (!formData.api_endpoint.trim()) {
      newErrors.api_endpoint = 'API endpoint is required'
    } else {
      try {
        new URL(formData.api_endpoint)
      } catch {
        newErrors.api_endpoint = 'Invalid URL format'
      }
    }

    if (!formData.username.trim()) {
      newErrors.username = 'Username is required'
    }

    if (!formData.password) {
      newErrors.password = 'Password is required'
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!validate()) return

    await onSubmit(formData)
  }

  const handleChange = (field: keyof RegisterClusterRequest) => (
    e: React.ChangeEvent<HTMLInputElement>
  ) => {
    setFormData((prev) => ({ ...prev, [field]: e.target.value }))
    if (errors[field]) {
      setErrors((prev) => ({ ...prev, [field]: undefined }))
    }
  }

  return (
    <Card>
      <CardHeader>
        <h2 className="text-lg font-semibold">Register New Cluster</h2>
      </CardHeader>
      <CardBody>
        <form onSubmit={handleSubmit} className="space-y-4">
          {error && (
            <div className="bg-red-50 text-red-700 px-4 py-3 rounded-md text-sm">
              {error}
            </div>
          )}

          <Input
            id="name"
            label="Cluster Name"
            placeholder="My Proxmox Cluster"
            value={formData.name}
            onChange={handleChange('name')}
            error={errors.name}
          />

          <Input
            id="api_endpoint"
            label="API Endpoint"
            placeholder="https://pve.example.com:8006"
            value={formData.api_endpoint}
            onChange={handleChange('api_endpoint')}
            error={errors.api_endpoint}
          />

          <Input
            id="username"
            label="Username"
            placeholder="root@pam"
            value={formData.username}
            onChange={handleChange('username')}
            error={errors.username}
          />

          <Input
            id="password"
            label="Password"
            type="password"
            value={formData.password}
            onChange={handleChange('password')}
            error={errors.password}
          />

          <div className="pt-4">
            <Button type="submit" loading={loading} className="w-full">
              Register Cluster
            </Button>
          </div>
        </form>
      </CardBody>
    </Card>
  )
}
