import { useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { clearAuth, getToken } from '../auth/authStorage'
import { ApiError } from '../services/http'
import {
  getAdminPurchaseById,
  getAdminPurchases,
  type AdminPurchase,
} from '../services/adminService'
import {
  formatCurrency,
  formatDateTime,
  sumAmounts,
} from '../utils/money'

function AdminPurchasesPage() {
  const navigate = useNavigate()
  const [purchases, setPurchases] = useState<AdminPurchase[]>([])
  const [selected, setSelected] = useState<AdminPurchase | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false

    async function loadPurchases() {
      const token = getToken()
      if (!token) {
        clearAuth()
        navigate('/login', { replace: true })
        return
      }

      setLoading(true)
      setError(null)

      try {
        const list = await getAdminPurchases(token)
        if (!cancelled) setPurchases(list)
      } catch (err) {
        if (cancelled) return
        if (err instanceof ApiError && err.status === 401) {
          clearAuth()
          navigate('/login', { replace: true })
          return
        }
        setError(
          err instanceof ApiError ? err.message : 'Unable to load purchases.',
        )
      } finally {
        if (!cancelled) setLoading(false)
      }
    }

    void loadPurchases()
    return () => {
      cancelled = true
    }
  }, [navigate])

  async function handleViewDetails(id: number) {
    const token = getToken()
    if (!token) {
      clearAuth()
      navigate('/login', { replace: true })
      return
    }

    setError(null)

    try {
      setSelected(await getAdminPurchaseById(id, token))
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        clearAuth()
        navigate('/login', { replace: true })
        return
      }
      setError(
        err instanceof ApiError ? err.message : 'Unable to load purchase.',
      )
    }
  }

  const totals = useMemo(() => {
    return {
      count: purchases.length,
      amount: sumAmounts(purchases.map((p) => p.amount)),
      commission: sumAmounts(purchases.map((p) => p.commission_amount)),
    }
  }, [purchases])

  return (
    <div className="pl-page">
      <header className="pl-page__intro">
        <h2>Purchases</h2>
        <p>All ledger purchase transactions.</p>
      </header>

      {error ? (
        <p className="pl-error" role="alert">
          {error}
        </p>
      ) : null}

      {loading ? (
        <p className="pl-status" role="status">
          Loading purchases...
        </p>
      ) : null}

      {!loading ? (
        <section className="pl-grid pl-grid--3" aria-label="Purchase summary">
          <article className="pl-card">
            <p className="pl-metric__label">Purchases</p>
            <p className="pl-metric__value">{totals.count}</p>
          </article>
          <article className="pl-card">
            <p className="pl-metric__label">Total Amount</p>
            <p className="pl-metric__value">{formatCurrency(totals.amount)}</p>
          </article>
          <article className="pl-card">
            <p className="pl-metric__label">Total Commission</p>
            <p className="pl-metric__value">
              {formatCurrency(totals.commission)}
            </p>
          </article>
        </section>
      ) : null}

      <section className="pl-card">
        <h3 className="pl-card__title">Purchase list</h3>

        {!loading && purchases.length === 0 ? (
          <p className="pl-empty">No purchases found.</p>
        ) : null}

        {!loading && purchases.length > 0 ? (
          <div className="pl-table-wrap">
            <table className="pl-table">
              <thead>
                <tr>
                  <th scope="col">ID</th>
                  <th scope="col">User ID</th>
                  <th scope="col">Merchant ID</th>
                  <th scope="col">Amount</th>
                  <th scope="col">Commission</th>
                  <th scope="col">Date</th>
                  <th scope="col">Actions</th>
                </tr>
              </thead>
              <tbody>
                {purchases.map((purchase) => (
                  <tr key={purchase.transaction_id}>
                    <td>#{purchase.transaction_id}</td>
                    <td>{purchase.user_id}</td>
                    <td>{purchase.merchant_id}</td>
                    <td>{formatCurrency(purchase.amount)}</td>
                    <td>{formatCurrency(purchase.commission_amount)}</td>
                    <td>{formatDateTime(purchase.created_at)}</td>
                    <td>
                      <button
                        type="button"
                        className="pl-btn"
                        onClick={() =>
                          void handleViewDetails(purchase.transaction_id)
                        }
                      >
                        View Details
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : null}
      </section>

      {selected ? (
        <section className="pl-card" aria-labelledby="purchase-detail-heading">
          <h3 id="purchase-detail-heading" className="pl-card__title">
            Purchase details
          </h3>
          <dl className="pl-details">
            <div>
              <dt>Transaction ID</dt>
              <dd>{selected.transaction_id}</dd>
            </div>
            <div>
              <dt>User ID</dt>
              <dd>{selected.user_id}</dd>
            </div>
            <div>
              <dt>Merchant ID</dt>
              <dd>{selected.merchant_id}</dd>
            </div>
            <div>
              <dt>Amount</dt>
              <dd>{formatCurrency(selected.amount)}</dd>
            </div>
            <div>
              <dt>Commission %</dt>
              <dd>{selected.commission_percentage}</dd>
            </div>
            <div>
              <dt>Commission Amount</dt>
              <dd>{formatCurrency(selected.commission_amount)}</dd>
            </div>
            <div>
              <dt>Created At</dt>
              <dd>{formatDateTime(selected.created_at)}</dd>
            </div>
          </dl>
        </section>
      ) : null}
    </div>
  )
}

export default AdminPurchasesPage
