import { useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
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

function MerchantTransactionsPage() {
  const navigate = useNavigate()
  const [transactions, setTransactions] = useState<MerchantTransaction[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false

    async function loadTransactions() {
      const token = getToken()
      if (!token) {
        clearAuth()
        navigate('/login', { replace: true })
        return
      }

      setLoading(true)
      setError(null)

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
        setError(
          err instanceof ApiError
            ? err.message
            : 'Unable to load transactions.',
        )
      } finally {
        if (!cancelled) setLoading(false)
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
    return {
      totalCustomerSpend,
      totalCommission,
      netSettlement: totalCustomerSpend - totalCommission,
      count: transactions.length,
      customers: new Set(transactions.map((tx) => tx.user_id)).size,
    }
  }, [transactions])

  return (
    <div className="pl-page">
      <header className="pl-page__intro">
        <h2>Transactions</h2>
        <p>All purchases processed through your merchant account.</p>
      </header>

      {loading ? (
        <p className="pl-status" role="status">
          Loading transactions...
        </p>
      ) : null}

      {!loading && error ? (
        <p className="pl-error" role="alert">
          {error}
        </p>
      ) : null}

      {!loading && !error ? (
        <>
          <section className="pl-grid pl-grid--4" aria-label="Transaction totals">
            <article className="pl-card">
              <p className="pl-metric__label">Transactions</p>
              <p className="pl-metric__value">{summary.count}</p>
            </article>
            <article className="pl-card">
              <p className="pl-metric__label">Customers</p>
              <p className="pl-metric__value">{summary.customers}</p>
            </article>
            <article className="pl-card">
              <p className="pl-metric__label">Customer Spend</p>
              <p className="pl-metric__value">
                {formatCurrency(summary.totalCustomerSpend)}
              </p>
            </article>
            <article className="pl-card">
              <p className="pl-metric__label">Net Settlement</p>
              <p className="pl-metric__value">
                {formatCurrency(summary.netSettlement)}
              </p>
            </article>
          </section>

          <section className="pl-card">
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
                    {transactions.map((tx) => (
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

export default MerchantTransactionsPage
