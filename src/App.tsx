import { Navigate, Route, Routes } from 'react-router-dom'
import {
  getPostLoginPath,
  getRole,
  isAuthenticated,
} from './auth/authStorage'
import {
  AdminAppLayout,
  MerchantAppLayout,
  UserAppLayout,
} from './components/AppLayout'
import ProtectedRoute from './components/ProtectedRoute'
import RoleProtectedRoute from './components/RoleProtectedRoute'
import AdminDashboard from './pages/AdminDashboard'
import AdminMerchantsPage from './pages/AdminMerchantsPage'
import AdminPaymentsPage from './pages/AdminPaymentsPage'
import AdminPurchasesPage from './pages/AdminPurchasesPage'
import AdminUsersPage from './pages/AdminUsersPage'
import LoginPage from './pages/LoginPage'
import MerchantDashboard from './pages/MerchantDashboard'
import MerchantRegisterPage from './pages/MerchantRegisterPage'
import MerchantTransactionsPage from './pages/MerchantTransactionsPage'
import PaymentHistoryPage from './pages/PaymentHistoryPage'
import PaymentPage from './pages/PaymentPage'
import PurchasePage from './pages/PurchasePage'
import RegisterPage from './pages/RegisterPage'
import UserDashboard from './pages/UserDashboard'

function HomeRedirect() {
  if (!isAuthenticated()) {
    return <Navigate to="/login" replace />
  }

  const role = getRole()
  if (!role) {
    return <Navigate to="/login" replace />
  }

  return <Navigate to={getPostLoginPath(role)} replace />
}

function App() {
  return (
    <Routes>
      <Route path="/" element={<HomeRedirect />} />
      <Route path="/login" element={<LoginPage />} />
      <Route path="/register" element={<RegisterPage />} />
      <Route path="/merchant/register" element={<MerchantRegisterPage />} />

      <Route
        element={
          <ProtectedRoute>
            <UserAppLayout />
          </ProtectedRoute>
        }
      >
        <Route path="/dashboard" element={<UserDashboard />} />
        <Route path="/purchases" element={<PurchasePage />} />
        <Route path="/payments" element={<PaymentPage />} />
        <Route path="/payments/history" element={<PaymentHistoryPage />} />
      </Route>

      <Route
        element={
          <ProtectedRoute>
            <MerchantAppLayout />
          </ProtectedRoute>
        }
      >
        <Route path="/merchant/dashboard" element={<MerchantDashboard />} />
        <Route
          path="/merchant/transactions"
          element={<MerchantTransactionsPage />}
        />
      </Route>

      <Route
        element={
          <RoleProtectedRoute requiredRole="admin">
            <AdminAppLayout />
          </RoleProtectedRoute>
        }
      >
        <Route path="/admin/dashboard" element={<AdminDashboard />} />
        <Route path="/admin/users" element={<AdminUsersPage />} />
        <Route path="/admin/merchants" element={<AdminMerchantsPage />} />
        <Route path="/admin/purchases" element={<AdminPurchasesPage />} />
        <Route path="/admin/payments" element={<AdminPaymentsPage />} />
      </Route>

      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}

export default App
