import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { clearAuth, getToken } from '../auth/authStorage'
import { ApiError } from '../services/http'
import { getPayments, type Payment } from '../services/paymentService'
import { formatCurrency, formatDateTime } from '../utils/money'

function PaymentHistoryPage() {
  const navigate = useNavigate()
  const [payments, setPayments] = useState<Payment[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false

    async function loadPayments() {
      const token = getToken()
      if (!token) {
        clearAuth()
        navigate('/login', { replace: true })
        return
      }

      setLoading(true)
      setError(null)

      try {
        const list = await getPayments(token)
        if (!cancelled) {
          setPayments(list)
        }
      } catch (err) {
        if (cancelled) {
          return
        }

        if (err instanceof ApiError && err.status === 401) {
          clearAuth()
          navigate('/login', { replace: true })
          return
        }

        if (err instanceof ApiError) {
          setError(err.message)
        } else if (err instanceof Error) {
          setError(err.message)
        } else {
          setError('Unable to load payment history.')
        }
      } finally {
        if (!cancelled) {
          setLoading(false)
        }
      }
    }

    void loadPayments()

    return () => {
      cancelled = true
    }
  }, [navigate])

  return (
    <div className="pl-page">
      <header className="pl-page__intro">
        <h2>Payment History</h2>
        <p>View your previous repayments.</p>
      </header>

      {loading ? (
        <p className="pl-status" role="status">
          Loading payment history...
        </p>
      ) : null}

      {!loading && error ? (
        <p className="pl-error" role="alert">
          {error}
        </p>
      ) : null}

      {!loading && !error && payments.length === 0 ? (
        <section className="pl-card" aria-live="polite">
          <h3 className="pl-card__title">No payments yet</h3>
          <p className="pl-empty">
            Your payment history will appear here after you make a payment.
          </p>
        </section>
      ) : null}

      {!loading && !error && payments.length > 0 ? (
        <section className="pl-card" aria-labelledby="payment-history-heading">
          <h3 id="payment-history-heading" className="pl-card__title">
            Your payments
          </h3>
          <div className="pl-table-wrap">
            <table className="pl-table">
              <thead>
                <tr>
                  <th scope="col">Payment ID</th>
                  <th scope="col">Amount</th>
                  <th scope="col">Date &amp; Time</th>
                </tr>
              </thead>
              <tbody>
                {payments.map((payment) => (
                  <tr key={payment.payment_id}>
                    <td>#{payment.payment_id}</td>
                    <td>{formatCurrency(payment.amount)}</td>
                    <td>{formatDateTime(payment.paid_at)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </section>
      ) : null}
    </div>
  )
}

export default PaymentHistoryPage
