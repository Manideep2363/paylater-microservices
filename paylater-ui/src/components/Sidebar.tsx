import './Sidebar.css'

const navItems = [
  'Dashboard',
  'Users',
  'Merchants',
  'Transactions',
  'Payments',
  'Reports',
  'Profile',
  'Logout',
] as const

function Sidebar() {
  return (
    <aside className="app-sidebar" aria-label="Main navigation">
      <nav>
        <ul className="app-sidebar__list">
          {navItems.map((item) => (
            <li key={item}>
              <button type="button" className="app-sidebar__link">
                {item}
              </button>
            </li>
          ))}
        </ul>
      </nav>
    </aside>
  )
}

export default Sidebar
