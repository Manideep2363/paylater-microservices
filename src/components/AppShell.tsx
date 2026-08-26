import type { ReactNode } from 'react'
import Header from './Header'
import Sidebar from './Sidebar'
import './AppShell.css'

type AppShellProps = {
  children: ReactNode
}

function AppShell({ children }: AppShellProps) {
  return (
    <div className="app-shell">
      <Header />
      <div className="app-shell__body">
        <Sidebar />
        <main className="app-shell__content">{children}</main>
      </div>
    </div>
  )
}

export default AppShell
