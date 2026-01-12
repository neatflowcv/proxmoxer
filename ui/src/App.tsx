import { Routes, Route } from 'react-router-dom'
import MainLayout from './components/layout/MainLayout'
import DashboardPage from './pages/DashboardPage'
import ClusterCreatePage from './pages/ClusterCreatePage'
import ClusterDetailPage from './pages/ClusterDetailPage'
import NotFoundPage from './pages/NotFoundPage'

function App() {
  return (
    <MainLayout>
      <Routes>
        <Route path="/" element={<DashboardPage />} />
        <Route path="/clusters/new" element={<ClusterCreatePage />} />
        <Route path="/clusters/:id" element={<ClusterDetailPage />} />
        <Route path="*" element={<NotFoundPage />} />
      </Routes>
    </MainLayout>
  )
}

export default App
