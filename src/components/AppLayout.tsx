import { NavLink, Outlet, useLocation } from 'react-router-dom'
import { useEffect, useState } from 'react'
import { AccountProvider, useAccount } from '../auth/AccountContext'
import { useLogoutConfirmation } from '../hooks/useLogoutConfirmation'
import './AppLayout.css'

export type NavItem = {
  to: string
  label: string
  icon: string
}

type AppLayoutProps = {
  title?: string
  navItems: NavItem[]
}

function AccountMenu() {
  const { displayName, email, role } = useAccount()
  const { requestLogout, logoutModal } = useLogoutConfirmation()
  const [open, setOpen] = useState(false)

  useEffect(() => {
    function onDocClick(event: MouseEvent) {
      const target = event.target as HTMLElement | null
      if (!target?.closest('.app-account')) {
        setOpen(false)
      }
    }
    document.addEventListener('click', onDocClick)
    return () => document.removeEventListener('click', onDocClick)
  }, [])

  return (
    <div className="app-account">
      {logoutModal}
      <button
        type="button"
        className="app-account__trigger"
        aria-expanded={open}
        aria-haspopup="menu"
        onClick={() => setOpen((value) => !value)}
      >
        <span className="app-account__avatar" aria-hidden="true">
          {displayName.slice(0, 1).toUpperCase()}
        </span>
        <span className="app-account__meta">
          <span className="app-account__name">{displayName}</span>
          <span className="app-account__role">{role ?? 'account'}</span>
        </span>
      </button>
      {open ? (
        <div className="app-account__menu" role="menu">
          <p className="app-account__menu-name">{displayName}</p>
          {email ? <p className="app-account__menu-email">{email}</p> : null}
          <button
            type="button"
            className="app-account__logout"
            role="menuitem"
            onClick={() => {
              setOpen(false)
              requestLogout()
            }}
          >
            Logout
          </button>
        </div>
      ) : null}
    </div>
  )
}

function LayoutShell({ navItems, title }: { navItems: NavItem[]; title?: string }) {
  const location = useLocation()
  const [mobileOpen, setMobileOpen] = useState(false)
  const { requestLogout, logoutModal } = useLogoutConfirmation()

  useEffect(() => {
    setMobileOpen(false)
  }, [location.pathname])

  const pageTitle =
    title ||
    navItems.find((item) => location.pathname === item.to)?.label ||
    'Dashboard'

  return (
    <div className={`app-layout${mobileOpen ? ' app-layout--nav-open' : ''}`}>
      {logoutModal}
      <aside className="app-sidebar" aria-label="Primary">
        <div className="app-sidebar__brand">
          <span className="app-sidebar__logo">PL</span>
          <span>PayLater</span>
        </div>
        <nav className="app-sidebar__nav">
          {navItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              className={({ isActive }) =>
                isActive
                  ? 'app-sidebar__link app-sidebar__link--active'
                  : 'app-sidebar__link'
              }
              end
            >
              <span className="app-sidebar__icon" aria-hidden="true">
                {item.icon}
              </span>
              {item.label}
            </NavLink>
          ))}
        </nav>
        <button
          type="button"
          className="app-sidebar__logout"
          onClick={requestLogout}
        >
          Logout
        </button>
      </aside>

      {mobileOpen ? (
        <button
          type="button"
          className="app-layout__backdrop"
          aria-label="Close navigation"
          onClick={() => setMobileOpen(false)}
        />
      ) : null}

      <div className="app-layout__content">
        <header className="app-topbar">
          <div className="app-topbar__left">
            <button
              type="button"
              className="app-topbar__menu"
              aria-label="Open navigation"
              onClick={() => setMobileOpen(true)}
            >
              Menu
            </button>
            <h1 className="app-topbar__title">{pageTitle}</h1>
          </div>
          <AccountMenu />
        </header>
        <main className="app-main">
          <Outlet />
        </main>
      </div>
    </div>
  )
}

export const USER_NAV: NavItem[] = [
  { to: '/dashboard', label: 'Dashboard', icon: '⌂' },
  { to: '/purchases', label: 'Make a Purchase', icon: '₹' },
  { to: '/payments', label: 'Make a Payment', icon: '↩' },
  { to: '/payments/history', label: 'Payment History', icon: '≡' },
]

export const MERCHANT_NAV: NavItem[] = [
  { to: '/merchant/dashboard', label: 'Dashboard', icon: '⌂' },
  { to: '/merchant/transactions', label: 'Transactions', icon: '≡' },
]

export const ADMIN_NAV: NavItem[] = [
  { to: '/admin/dashboard', label: 'Dashboard', icon: '⌂' },
  { to: '/admin/users', label: 'Users', icon: 'U' },
  { to: '/admin/merchants', label: 'Merchants', icon: 'M' },
  { to: '/admin/purchases', label: 'Purchases', icon: 'P' },
  { to: '/admin/payments', label: 'Payments', icon: '$' },
]

function AppLayout({ navItems, title }: AppLayoutProps) {
  return (
    <AccountProvider>
      <LayoutShell navItems={navItems} title={title} />
    </AccountProvider>
  )
}

export function UserAppLayout() {
  return <AppLayout navItems={USER_NAV} />
}

export function MerchantAppLayout() {
  return <AppLayout navItems={MERCHANT_NAV} />
}

export function AdminAppLayout() {
  return <AppLayout navItems={ADMIN_NAV} />
}

export default AppLayout
