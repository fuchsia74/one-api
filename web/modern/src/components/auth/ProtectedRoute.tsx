import { Navigate, Outlet, useLocation } from 'react-router-dom'
import { useAuthStore } from '@/lib/stores/auth'

export function ProtectedRoute() {
  const { user } = useAuthStore()
  const location = useLocation()

  // Redirect to login with current path as redirect_to parameter
  if (!user) {
    const redirectTo = encodeURIComponent(location.pathname + location.search)
    return <Navigate to={`/login?redirect_to=${redirectTo}`} replace />
  }

  return <Outlet />
}
