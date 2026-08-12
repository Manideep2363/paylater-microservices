import { useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAccount } from '../auth/AccountContext'
import { clearAuth, getToken } from '../auth/authStorage'
import { ApiError } from '../services/http'
import {
  getMerchantTransactions,
  type MerchantTransaction,
} from '../services/merchantService'
import {
  formatCurrency,
  formatDateTime,
  parseAmount,
  sumAmounts,
} from '../utils/money'

const RECENT_TRANSACTION_LIMIT = 5

function MerchantDashboard() {
  const navigate = useNavigate()
  const { merchant, loading: profileLoading, error: profileError, refresh } =
    useAccount()
  const [transactions, setTransactions] = useState<MerchantTransaction[]>([])
  const [loadingTx, setLoadingTx] = useState(true)
  const [txError, setTxError] = useState<string | null>(null)

  useEffect(() => {
    void refresh()
  }, [refresh])

  useEffect(() => {
    let cancelled = false

    async function loadTransactions() {
      const token = getToken()
      if (!token) {
        clearAuth()
        navigate('/login', { replace: true })
        return
      }

      setLoadingTx(true)
      setTxError(null)

      try {
        const list = await getMerchantTransactions(token)
        if (!cancelled) {
          setTransactions(list)
        }
      } catch (err) {
        if (cancelled) return
        if (err instanceof ApiError && err.status === 401) {
          clearAuth()
          navigate('/login', { replace: true })
          return
        }
        setTxError(
          err instanceof ApiError
            ? err.message
            : 'Unable to load merchant transactions.',
        )
      } finally {
        if (!cancelled) setLoadingTx(false)
      }
    }

    void loadTransactions()
    return () => {
      cancelled = true
    }
  }, [navigate])

  const summary = useMemo(() => {
    const totalCustomerSpend = sumAmounts(transactions.map((tx) => tx.amount))
    const totalCommission = sumAmounts(
      transactions.map((tx) => tx.commission_amount),
    )
    const netSettlement = totalCustomerSpend - totalCommission
    const customerCount = new Set(transactions.map((tx) => tx.user_id)).size

    return {
      totalCustomerSpend,
      totalCommission,
      netSettlement,
      customerCount,
      totalTransactions: transactions.length,
    }
  }, [transactions])

  const loading = (profileLoading && !merchant) || loadingTx
  const error = profileError || txError
  const recentTransactions = transactions.slice(0, RECENT_TRANSACTION_LIMIT)

  return (
    <div className="pl-page">
      <header className="pl-page__intro">
        <h2>
          {merchant ? `Welcome back, ${merchant.name}` : 'Merchant Dashboard'}
        </h2>
        <p>Track customer spend, commission, and settlement.</p>
      </header>

      {loading ? (
        <p className="pl-status" role="status">
          Loading dashboard...
        </p>
      ) : null}

      {!loading && error ? (
        <p className="pl-error" role="alert">
          {error}
        </p>
      ) : null}

      {!loading && !error && merchant ? (
        <>
          <section aria-labelledby="metrics-heading">
            <h3 id="metrics-heading" className="pl-card__title">
              Business Metrics
            </h3>
            <div className="pl-grid pl-grid--4">
              <article className="pl-card">
                <p className="pl-metric__label">Total Transactions</p>
                <p className="pl-metric__value">{summary.totalTransactions}</p>
              </article>
              <article className="pl-card">
                <p className="pl-metric__label">Customers</p>
                <p className="pl-metric__value">{summary.customerCount}</p>
              </article>
              <article className="pl-card">
                <p className="pl-metric__label">Total Customer Spend</p>
                <p className="pl-metric__value">
                  {formatCurrency(summary.totalCustomerSpend)}
                </p>
              </article>
              <article className="pl-card">
                <p className="pl-metric__label">Total Commission</p>
                <p className="pl-metric__value">
                  {formatCurrency(summary.totalCommission)}
                </p>
              </article>
            </div>
          </section>

          <section className="pl-card" aria-labelledby="settlement-heading">
            <h3 id="settlement-heading" className="pl-card__title">
              Net Settlement
            </h3>
            <dl className="pl-details">
              <div>
                <dt>Customer Purchases</dt>
                <dd>{formatCurrency(summary.totalCustomerSpend)}</dd>
              </div>
              <div>
                <dt>Commission</dt>
                <dd>{formatCurrency(summary.totalCommission)}</dd>
              </div>
              <div>
                <dt>Net Settlement</dt>
                <dd>
                  <strong>{formatCurrency(summary.netSettlement)}</strong>
                </dd>
              </div>
            </dl>
          </section>

          <section className="pl-card" aria-labelledby="profile-heading">
            <h3 id="profile-heading" className="pl-card__title">
              Merchant Profile
            </h3>
            <dl className="pl-details">
              <div>
                <dt>Merchant Name</dt>
                <dd>{merchant.name}</dd>
              </div>
              <div>
                <dt>Email</dt>
                <dd>{merchant.email}</dd>
              </div>
              <div>
                <dt>Phone</dt>
                <dd>{merchant.phone}</dd>
              </div>
              <div>
                <dt>Commission Percentage</dt>
                <dd>{merchant.commission_percentage}%</dd>
              </div>
            </dl>
          </section>

          <section className="pl-card" aria-labelledby="recent-heading">
            <div className="pl-card__header">
              <h3 id="recent-heading" className="pl-card__title">
                Recent transactions
              </h3>
              <button
                type="button"
                className="pl-btn"
                onClick={() => navigate('/merchant/transactions')}
              >
                View Details
              </button>
            </div>

            {transactions.length === 0 ? (
              <p className="pl-empty" role="status">
                No transactions yet.
              </p>
            ) : (
              <div className="pl-table-wrap">
                <table className="pl-table">
                  <thead>
                    <tr>
                      <th scope="col">Transaction ID</th>
                      <th scope="col">Customer</th>
                      <th scope="col">Amount</th>
                      <th scope="col">Commission</th>
                      <th scope="col">Net Settlement</th>
                      <th scope="col">Date</th>
                    </tr>
                  </thead>
                  <tbody>
                    {recentTransactions.map((tx) => (
                      <tr key={tx.transaction_id}>
                        <td>#{tx.transaction_id}</td>
                        <td>Customer #{tx.user_id}</td>
                        <td>{formatCurrency(tx.amount)}</td>
                        <td>{formatCurrency(tx.commission_amount)}</td>
                        <td>
                          {formatCurrency(
                            parseAmount(tx.amount) -
                              parseAmount(tx.commission_amount),
                          )}
                        </td>
                        <td>{formatDateTime(tx.created_at)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </section>
        </>
      ) : null}
    </div>
  )
}

export default MerchantDashboard
