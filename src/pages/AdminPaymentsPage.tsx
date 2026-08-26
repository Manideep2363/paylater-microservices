import { useState, type FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { clearAuth, getToken } from '../auth/authStorage'
import { ApiError } from '../services/http'
import {
  getAdminPaymentById,
  getAdminUserPayments,
  type AdminPayment,
} from '../services/adminService'
import { formatCurrency, formatDateTime } from '../utils/money'
import '../components/AdminShared.css'

function AdminPaymentsPage() {
  const navigate = useNavigate()
  const [paymentId, setPaymentId] = useState('')
  const [userId, setUserId] = useState('')
  const [payment, setPayment] = useState<AdminPayment | null>(null)
  const [userPayments, setUserPayments] = useState<AdminPayment[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function handleLookupPayment(event: FormEvent) {
    event.preventDefault()
    const token = getToken()
    if (!token) {
      clearAuth()
      navigate('/login', { replace: true })
      return
    }

    setLoading(true)
    setError(null)
    setPayment(null)

    try {
      setPayment(await getAdminPaymentById(Number(paymentId), token))
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        clearAuth()
        navigate('/login', { replace: true })
        return
      }
      setError(
        err instanceof ApiError ? err.message : 'Unable to load payment.',
      )
    } finally {
      setLoading(false)
    }
  }

  async function handleLookupUserPayments(event: FormEvent) {
    event.preventDefault()
    const token = getToken()
    if (!token) {
      clearAuth()
      navigate('/login', { replace: true })
      return
    }

    setLoading(true)
    setError(null)
    setUserPayments([])

    try {
      setUserPayments(await getAdminUserPayments(Number(userId), token))
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        clearAuth()
        navigate('/login', { replace: true })
        return
      }
      setError(
        err instanceof ApiError
          ? err.message
          : 'Unable to load user payments.',
      )
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="pl-page">
      <header className="pl-page__intro">
        <h2>Payments</h2>
        <p>Look up payments using existing admin payment APIs.</p>
      </header>

      {error ? (
        <p className="pl-error" role="alert">
          {error}
        </p>
      ) : null}

      {loading ? (
        <p className="pl-status" role="status">
          Loading...
        </p>
      ) : null}

      <section className="pl-card">
        <h3 className="pl-card__title">Search payment by Payment ID</h3>
        <p className="pl-note">
          There is no list-all payments endpoint. Use payment ID lookup or view
          payments for a specific user.
        </p>

        <form className="admin-form" onSubmit={handleLookupPayment}>
          <div className="admin-field">
            <label htmlFor="admin-payment-id">Payment ID</label>
            <input
              id="admin-payment-id"
              type="number"
              min="1"
              value={paymentId}
              onChange={(e) => setPaymentId(e.target.value)}
              required
              disabled={loading}
            />
          </div>
          <button
            type="submit"
            className="pl-btn pl-btn--primary"
            disabled={loading}
          >
            Get payment details
          </button>
        </form>

        {payment ? (
          <dl className="pl-details" style={{ marginTop: 16 }}>
            <div>
              <dt>Payment ID</dt>
              <dd>{payment.payment_id}</dd>
            </div>
            <div>
              <dt>User ID</dt>
              <dd>{payment.user_id}</dd>
            </div>
            <div>
              <dt>Amount</dt>
              <dd>{formatCurrency(payment.amount)}</dd>
            </div>
            <div>
              <dt>Paid At</dt>
              <dd>{formatDateTime(payment.paid_at)}</dd>
            </div>
          </dl>
        ) : null}
      </section>

      <section className="pl-card">
        <h3 className="pl-card__title">View payments for a user</h3>
        <form className="admin-form" onSubmit={handleLookupUserPayments}>
          <div className="admin-field">
            <label htmlFor="admin-payment-user-id">User ID</label>
            <input
              id="admin-payment-user-id"
              type="number"
              min="1"
              value={userId}
              onChange={(e) => setUserId(e.target.value)}
              required
              disabled={loading}
            />
          </div>
          <button
            type="submit"
            className="pl-btn pl-btn--primary"
            disabled={loading}
          >
            List user payments
          </button>
        </form>

        {userPayments.length === 0 ? (
          <p className="pl-empty">
            No user payments loaded. Enter a user ID to fetch payments.
          </p>
        ) : (
          <div className="pl-table-wrap">
            <table className="pl-table">
              <thead>
                <tr>
                  <th scope="col">Payment ID</th>
                  <th scope="col">User ID</th>
                  <th scope="col">Amount</th>
                  <th scope="col">Paid At</th>
                  <th scope="col">Actions</th>
                </tr>
              </thead>
              <tbody>
                {userPayments.map((row) => (
                  <tr key={row.payment_id}>
                    <td>#{row.payment_id}</td>
                    <td>{row.user_id}</td>
                    <td>{formatCurrency(row.amount)}</td>
                    <td>{formatDateTime(row.paid_at)}</td>
                    <td>
                      <button
                        type="button"
                        className="pl-btn"
                        onClick={() => {
                          setPaymentId(String(row.payment_id))
                          setPayment(row)
                        }}
                      >
                        View Details
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>
    </div>
  )
}

export default AdminPaymentsPage
