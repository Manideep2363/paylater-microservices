import type { ReactNode } from 'react'
import { Navigate } from 'react-router-dom'
import { getRole, isAuthenticated, type AuthRole } from '../auth/authStorage'

type RoleProtectedRouteProps = {
  children: ReactNode
  requiredRole: AuthRole
}

function RoleProtectedRoute({
  children,
  requiredRole,
}: RoleProtectedRouteProps) {
  if (!isAuthenticated()) {
    return <Navigate to="/login" replace />
  }

  const role = getRole()
  if (role !== requiredRole) {
    if (role === 'user') {
      return <Navigate to="/dashboard" replace />
    }
    if (role === 'merchant') {
      return <Navigate to="/merchant/dashboard" replace />
    }
    return <Navigate to="/login" replace />
  }

  return children
}

export default RoleProtectedRoute
