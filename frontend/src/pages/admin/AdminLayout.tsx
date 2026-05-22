import { NavLink, Outlet, useNavigate } from 'react-router-dom'
import { adminLogout, isAdminLoggedIn } from '../../api'
import { useEffect } from 'react'

const NAV = [
  { to: '/admin/courses',   label: 'Courses',          icon: '📚' },
  { to: '/admin/languages', label: 'Languages',         icon: '🌐' },
  { to: '/admin/grammar',   label: 'Grammar Coverage',  icon: '📝' },
]

export default function AdminLayout() {
  const navigate = useNavigate()

  useEffect(() => {
    if (!isAdminLoggedIn()) navigate('/admin/login')
  }, [navigate])

  function handleLogout() {
    adminLogout()
    navigate('/admin/login')
  }

  return (
    <div className="admin-shell">
      <aside className="admin-sidebar">
        <div className="admin-sidebar-section">Admin</div>
        <nav>
          {NAV.map(item => (
            <NavLink
              key={item.to}
              to={item.to}
              className={({ isActive }) => `admin-nav-item${isActive ? ' active' : ''}`}
            >
              <span style={{ fontSize: '1rem' }}>{item.icon}</span>
              {item.label}
            </NavLink>
          ))}
        </nav>
        <div className="admin-sidebar-footer">
          <button className="btn btn-ghost btn-sm" onClick={handleLogout} style={{ width: '100%', justifyContent: 'flex-start' }}>
            Sign out
          </button>
        </div>
      </aside>
      <main className="admin-main">
        <Outlet />
      </main>
    </div>
  )
}
