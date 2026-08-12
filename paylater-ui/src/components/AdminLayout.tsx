import type { ReactNode } from 'react'
import { useNavigate } from 'react-router-dom'
import { useLogoutConfirmation } from '../hooks/useLogoutConfirmation'
import './AdminShared.css'

type AdminLayoutProps = {
  title: string
  subtitle?: string
  children: ReactNode
}

const NAV_ITEMS = [
  { label: 'Dashboard', path: '/admin/dashboard' },
  { label: 'Users', path: '/admin/users' },
  { label: 'Merchants', path: '/admin/merchants' },
  { label: 'Purchases', path: '/admin/purchases' },
  { label: 'Payments', path: '/admin/payments' },
] as const

function AdminLayout({ title, subtitle, children }: AdminLayoutProps) {
  const navigate = useNavigate()
  const { requestLogout, logoutModal } = useLogoutConfirmation()

  return (
    <div className="admin-shell">
      {logoutModal}
      <header className="admin-shell__header">
        <div>
          <p className="admin-shell__brand">PayLater</p>
          <h1 className="admin-shell__title">{title}</h1>
          {subtitle ? <p className="admin-shell__subtitle">{subtitle}</p> : null}
        </div>
        <nav className="admin-shell__nav" aria-label="Admin navigation">
          {NAV_ITEMS.map((item) => (
            <button
              key={item.path}
              type="button"
              className="admin-shell__nav-link"
              onClick={() => navigate(item.path)}
            >
              {item.label}
            </button>
          ))}
          <button
            type="button"
            className="admin-shell__logout"
            onClick={requestLogout}
          >
            Logout
          </button>
        </nav>
      </header>
      <main className="admin-shell__main">{children}</main>
    </div>
  )
}

export default AdminLayout
