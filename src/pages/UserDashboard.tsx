import { useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAccount } from '../auth/AccountContext'
import type { PaymentNavigationState } from '../types/navigation'
import {
  calcAvailableCredit,
  calcCreditUsagePercent,
  formatCurrency,
  parseAmount,
} from '../utils/money'

function UserDashboard() {
  const navigate = useNavigate()
  const { user, loading, error, refresh } = useAccount()

  useEffect(() => {
    void refresh()
  }, [refresh])

  if (loading && !user) {
    return (
      <div className="pl-page">
        <p className="pl-status" role="status">
          Loading your account…
        </p>
      </div>
    )
  }

  if (error && !user) {
    return (
      <div className="pl-page">
        <p className="pl-error" role="alert">
          {error}
        </p>
      </div>
    )
  }

  if (!user) {
    return (
      <div className="pl-page">
        <p className="pl-status">Unable to load your account.</p>
      </div>
    )
  }

  const currentDue = parseAmount(user.current_due)
  const availableCredit = calcAvailableCredit(
    user.credit_limit,
    user.current_due,
  )
  const creditUsage = calcCreditUsagePercent(
    user.credit_limit,
    user.current_due,
  )
  const hasDue = currentDue > 0

  function goPayFullDue() {
    const state: PaymentNavigationState = {
      payFullDue: true,
      amount: user!.current_due,
    }
    navigate('/payments', { state })
  }

  return (
    <div className="pl-page">
      <header className="pl-page__intro">
        <h2>Welcome back, {user.name}</h2>
        <p>Review your credit position and manage repayments.</p>
      </header>

      <section aria-labelledby="finance-heading">
        <h3 id="finance-heading" className="pl-card__title">
          Financial Overview
        </h3>
        <div className="pl-grid pl-grid--4">
          <article className="pl-card">
            <p className="pl-metric__label">Credit Limit</p>
            <p className="pl-metric__value">
              {formatCurrency(user.credit_limit)}
            </p>
          </article>
          <article className="pl-card">
            <p className="pl-metric__label">Current Due</p>
            <p className="pl-metric__value pl-metric__value--due">
              {formatCurrency(user.current_due)}
            </p>
          </article>
          <article className="pl-card">
            <p className="pl-metric__label">Available Credit</p>
            <p className="pl-metric__value">{formatCurrency(availableCredit)}</p>
          </article>
          <article className="pl-card">
            <p className="pl-metric__label">Credit Used</p>
            <p className="pl-metric__value">{creditUsage.toFixed(1)}%</p>
            <div
              className="pl-progress"
              role="progressbar"
              aria-valuenow={Number(creditUsage.toFixed(1))}
              aria-valuemin={0}
              aria-valuemax={100}
              aria-label="Credit usage"
            >
              <div
                className="pl-progress__bar"
                style={{ width: `${creditUsage}%` }}
              />
            </div>
          </article>
        </div>
      </section>

      <section className="pl-card" aria-labelledby="actions-heading">
        <h3 id="actions-heading" className="pl-card__title">
          Quick Actions
        </h3>
        <div className="pl-actions">
          <button
            type="button"
            className="pl-btn pl-btn--primary"
            onClick={() => navigate('/purchases')}
          >
            Make a Purchase
          </button>
          <button
            type="button"
            className="pl-btn"
            onClick={() => navigate('/payments')}
          >
            Make a Payment
          </button>
          {hasDue ? (
            <button
              type="button"
              className="pl-btn"
              onClick={goPayFullDue}
            >
              Pay Full Due
            </button>
          ) : (
            <p className="pl-status">No outstanding balance</p>
          )}
        </div>
      </section>

      <section className="pl-card" aria-labelledby="account-heading">
        <h3 id="account-heading" className="pl-card__title">
          Account Overview
        </h3>
        <dl className="pl-details">
          <div>
            <dt>Name</dt>
            <dd>{user.name}</dd>
          </div>
          <div>
            <dt>Email</dt>
            <dd>{user.email}</dd>
          </div>
        </dl>
      </section>

      <section className="pl-card" aria-labelledby="purchase-history-heading">
        <h3 id="purchase-history-heading" className="pl-card__title">
          Purchase History
        </h3>
        <p className="pl-note">
          Purchase history will be available here. A user-scoped purchases API
          is not available yet—only admin purchase endpoints exist today.
        </p>
      </section>
    </div>
  )
}

export default UserDashboard
