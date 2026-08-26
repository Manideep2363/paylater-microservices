import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { clearAuth, getToken } from '../auth/authStorage'
import { ApiError } from '../services/http'
import {
  getAdminMerchants,
  getAdminUsers,
  getMerchantCommissionReport,
  getOutstandingBalanceReport,
  getUsersAtCreditLimitReport,
  getUsersDueReport,
  type MerchantCommissionReportRow,
  type UserAtCreditLimitRow,
  type UserDueReportRow,
} from '../services/adminService'
import { formatCurrency, sumAmounts } from '../utils/money'
import '../components/AdminShared.css'

function AdminDashboard() {
  const navigate = useNavigate()
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [totalUsers, setTotalUsers] = useState(0)
  const [totalMerchants, setTotalMerchants] = useState(0)
  const [outstandingBalance, setOutstandingBalance] = useState('0.00')
  const [usersDue, setUsersDue] = useState<UserDueReportRow[]>([])
  const [usersAtLimit, setUsersAtLimit] = useState<UserAtCreditLimitRow[]>([])
  const [commissions, setCommissions] = useState<MerchantCommissionReportRow[]>(
    [],
  )

  useEffect(() => {
    let cancelled = false

    async function loadReports() {
      const token = getToken()
      if (!token) {
        clearAuth()
        navigate('/login', { replace: true })
        return
      }

      setLoading(true)
      setError(null)

      try {
        const [
          users,
          merchants,
          balance,
          due,
          atLimit,
          merchantCommissions,
        ] = await Promise.all([
          getAdminUsers(token),
          getAdminMerchants(token),
          getOutstandingBalanceReport(token),
          getUsersDueReport(token),
          getUsersAtCreditLimitReport(token),
          getMerchantCommissionReport(token),
        ])

        if (!cancelled) {
          setTotalUsers(users.length)
          setTotalMerchants(merchants.length)
          setOutstandingBalance(balance.total_outstanding_balance)
          setUsersDue(due)
          setUsersAtLimit(atLimit)
          setCommissions(merchantCommissions)
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
            : 'Unable to load admin dashboard.',
        )
      } finally {
        if (!cancelled) setLoading(false)
      }
    }

    void loadReports()
    return () => {
      cancelled = true
    }
  }, [navigate])

  const totalCommission = sumAmounts(
    commissions.map((row) => row.total_commission),
  )

  return (
    <div className="pl-page">
      <header className="pl-page__intro">
        <h2>Admin Dashboard</h2>
        <p>Platform overview from live admin APIs and reports.</p>
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

      {!loading && !error ? (
        <>
          <section aria-labelledby="mgmt-heading">
            <h3 id="mgmt-heading" className="pl-card__title">
              Management Overview
            </h3>
            <div className="pl-grid pl-grid--2">
              <article className="pl-card">
                <p className="pl-metric__label">Users</p>
                <p className="pl-metric__value">{totalUsers}</p>
              </article>
              <article className="pl-card">
                <p className="pl-metric__label">Merchants</p>
                <p className="pl-metric__value">{totalMerchants}</p>
              </article>
            </div>
          </section>

          <section aria-labelledby="finance-heading">
            <h3 id="finance-heading" className="pl-card__title">
              Financial Overview
            </h3>
            <div className="pl-grid pl-grid--4">
              <article className="pl-card">
                <p className="pl-metric__label">Outstanding Balance</p>
                <p className="pl-metric__value">
                  {formatCurrency(outstandingBalance)}
                </p>
              </article>
              <article className="pl-card">
                <p className="pl-metric__label">Users With Due</p>
                <p className="pl-metric__value">{usersDue.length}</p>
              </article>
              <article className="pl-card">
                <p className="pl-metric__label">At Credit Limit</p>
                <p className="pl-metric__value">{usersAtLimit.length}</p>
              </article>
              <article className="pl-card">
                <p className="pl-metric__label">Merchant Commissions</p>
                <p className="pl-metric__value">
                  {formatCurrency(totalCommission)}
                </p>
              </article>
            </div>
          </section>

          <section className="pl-card" aria-labelledby="due-heading">
            <h3 id="due-heading" className="pl-card__title">
              Users with outstanding due
            </h3>
            {usersDue.length === 0 ? (
              <p className="pl-empty">No users with outstanding due.</p>
            ) : (
              <div className="pl-table-wrap">
                <table className="pl-table">
                  <thead>
                    <tr>
                      <th scope="col">User ID</th>
                      <th scope="col">Name</th>
                      <th scope="col">Current Due</th>
                    </tr>
                  </thead>
                  <tbody>
                    {usersDue.map((row) => (
                      <tr key={row.user_id}>
                        <td>{row.user_id}</td>
                        <td>{row.name}</td>
                        <td>{formatCurrency(row.current_due)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </section>

          <section className="pl-card" aria-labelledby="limit-heading">
            <h3 id="limit-heading" className="pl-card__title">
              Users at credit limit
            </h3>
            {usersAtLimit.length === 0 ? (
              <p className="pl-empty">No users at credit limit.</p>
            ) : (
              <div className="pl-table-wrap">
                <table className="pl-table">
                  <thead>
                    <tr>
                      <th scope="col">User ID</th>
                      <th scope="col">Name</th>
                      <th scope="col">Credit Limit</th>
                      <th scope="col">Current Due</th>
                    </tr>
                  </thead>
                  <tbody>
                    {usersAtLimit.map((row) => (
                      <tr key={row.user_id}>
                        <td>{row.user_id}</td>
                        <td>{row.name}</td>
                        <td>{formatCurrency(row.credit_limit)}</td>
                        <td>{formatCurrency(row.current_due)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </section>

          <section className="pl-card" aria-labelledby="commission-heading">
            <h3 id="commission-heading" className="pl-card__title">
              Merchant commissions
            </h3>
            {commissions.length === 0 ? (
              <p className="pl-empty">No merchant commission rows.</p>
            ) : (
              <div className="pl-table-wrap">
                <table className="pl-table">
                  <thead>
                    <tr>
                      <th scope="col">Merchant ID</th>
                      <th scope="col">Total Commission</th>
                    </tr>
                  </thead>
                  <tbody>
                    {commissions.map((row) => (
                      <tr key={row.merchant_id}>
                        <td>{row.merchant_id}</td>
                        <td>{formatCurrency(row.total_commission)}</td>
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

export default AdminDashboard
