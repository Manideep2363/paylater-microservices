import { useEffect, useState, type FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import PasswordInput from '../components/PasswordInput'
import { clearAuth, getToken } from '../auth/authStorage'
import { ApiError } from '../services/http'
import {
  createAdminUser,
  getAdminUsers,
  type AdminUser,
} from '../services/adminService'
import { formatCurrency } from '../utils/money'
import '../components/AdminShared.css'

function AdminUsersPage() {
  const navigate = useNavigate()
  const [users, setUsers] = useState<AdminUser[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState<string | null>(null)
  const [showForm, setShowForm] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')

  async function loadUsers() {
    const token = getToken()
    if (!token) {
      clearAuth()
      navigate('/login', { replace: true })
      return
    }

    setLoading(true)
    setError(null)

    try {
      const list = await getAdminUsers(token)
      setUsers(list)
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        clearAuth()
        navigate('/login', { replace: true })
        return
      }
      setError(err instanceof ApiError ? err.message : 'Unable to load users.')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void loadUsers()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [navigate])

  async function handleCreate(event: FormEvent) {
    event.preventDefault()
    const token = getToken()
    if (!token) {
      clearAuth()
      navigate('/login', { replace: true })
      return
    }

    setSubmitting(true)
    setError(null)
    setSuccess(null)

    try {
      await createAdminUser(
        { name: name.trim(), email: email.trim(), password },
        token,
      )
      setSuccess('User created successfully.')
      setName('')
      setEmail('')
      setPassword('')
      setShowForm(false)
      await loadUsers()
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        clearAuth()
        navigate('/login', { replace: true })
        return
      }
      setError(err instanceof ApiError ? err.message : 'Unable to create user.')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="pl-page">
      <header className="pl-page__intro">
        <h2>Users</h2>
        <p>Manage PayLater users.</p>
      </header>

      <section className="pl-card">
        <div className="pl-card__header">
          <h3 className="pl-card__title">All users</h3>
          <button
            type="button"
            className="pl-btn pl-btn--primary"
            onClick={() => setShowForm((value) => !value)}
          >
            {showForm ? 'Hide form' : 'Create User'}
          </button>
        </div>

        {success ? <p className="pl-success">{success}</p> : null}
        {error ? (
          <p className="pl-error" role="alert">
            {error}
          </p>
        ) : null}

        {showForm ? (
          <form className="admin-form" onSubmit={handleCreate}>
            <div className="admin-field">
              <label htmlFor="admin-user-name">Name</label>
              <input
                id="admin-user-name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                required
                disabled={submitting}
              />
            </div>
            <div className="admin-field">
              <label htmlFor="admin-user-email">Email</label>
              <input
                id="admin-user-email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
                disabled={submitting}
              />
            </div>
            <PasswordInput
              className="admin-field"
              label="Password"
              id="admin-user-password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              minLength={6}
              required
              disabled={submitting}
            />
            <button
              type="submit"
              className="pl-btn pl-btn--primary"
              disabled={submitting}
            >
              {submitting ? 'Creating…' : 'Create user'}
            </button>
          </form>
        ) : null}

        {loading ? (
          <p className="pl-status" role="status">
            Loading users...
          </p>
        ) : null}

        {!loading && users.length === 0 ? (
          <p className="pl-empty">No users found.</p>
        ) : null}

        {!loading && users.length > 0 ? (
          <div className="pl-table-wrap">
            <table className="pl-table">
              <thead>
                <tr>
                  <th scope="col">User ID</th>
                  <th scope="col">Name</th>
                  <th scope="col">Email</th>
                  <th scope="col">Credit Limit</th>
                  <th scope="col">Current Due</th>
                </tr>
              </thead>
              <tbody>
                {users.map((user) => (
                  <tr key={user.user_id}>
                    <td>{user.user_id}</td>
                    <td>{user.name}</td>
                    <td>{user.email}</td>
                    <td>{formatCurrency(user.credit_limit)}</td>
                    <td>{formatCurrency(user.current_due)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : null}
      </section>
    </div>
  )
}

export default AdminUsersPage
