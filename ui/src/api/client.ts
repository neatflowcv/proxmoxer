import type { ErrorResponse } from './types'

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'

export class ApiError extends Error {
  constructor(
    public status: number,
    message: string,
    public response?: ErrorResponse
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

interface RequestOptions {
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE'
  body?: unknown
  headers?: Record<string, string>
}

export async function request<T>(endpoint: string, options?: RequestOptions): Promise<T> {
  const url = `${API_BASE_URL}${endpoint}`

  const response = await fetch(url, {
    method: options?.method || 'GET',
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
    body: options?.body ? JSON.stringify(options.body) : undefined,
  })

  if (!response.ok) {
    let errorResponse: ErrorResponse | undefined
    try {
      errorResponse = await response.json()
    } catch {
      // ignore JSON parse error
    }
    throw new ApiError(
      response.status,
      errorResponse?.message || `HTTP error ${response.status}`,
      errorResponse
    )
  }

  if (response.status === 204) {
    return undefined as T
  }

  return response.json()
}
