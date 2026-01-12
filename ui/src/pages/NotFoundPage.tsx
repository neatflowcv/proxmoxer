import { Link } from 'react-router-dom'
import { Button } from '../components/common'

export default function NotFoundPage() {
  return (
    <div className="flex flex-col items-center justify-center py-24">
      <h1 className="text-6xl font-bold text-gray-300">404</h1>
      <h2 className="mt-4 text-xl font-semibold text-gray-900">Page not found</h2>
      <p className="mt-2 text-gray-500">The page you're looking for doesn't exist.</p>
      <Link to="/" className="mt-6">
        <Button>Go to Dashboard</Button>
      </Link>
    </div>
  )
}
