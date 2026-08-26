import {
  useEffect,
  useState,
  type FormEvent,
  type ChangeEvent,
} from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { useAccount } from '../auth/AccountContext'
import { clearAuth, getToken } from '../auth/authStorage'
import { getUserIdFromToken } from '../auth/jwt'
import { ApiError } from '../services/http'
import { createPayment } from '../services/paymentService'
import { getCurrentUser, type User } from '../services/userService'
import type { PaymentNavigationState } from '../types/navigation'
import {
  formatCurrency,
  parseAmount,
} from '../utils/money'

function hasOutstandingDue(currentDue: string): boolean {
  return parseAmount(currentDue) > 0
}

function PaymentPage() {
  const navigate = useNavigate()
  const location = useLocation()
  const { refresh: refreshAccount } = useAccount()

  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)
  const [loadError, setLoadError] = useState<string | null>(null)

  const [amount, setAmount] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)
  const [successMessage, setSuccessMessage] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false

    async function loadUser() {
      const token = getToken()
      if (!token) {
        clearAuth()
        navigate('/login', { replace: true })
        return
      }

      setLoading(true)
      setLoadError(null)

      try {
        const userId = getUserIdFromToken(token)
        const profile = await getCurrentUser(userId, token)
        if (!cancelled) {
          setUser(profile)

          const navState = location.state as PaymentNavigationState | null
          if (navState?.payFullDue || navState?.amount != null) {
            const due = parseAmount(profile.current_due)
            const requested = parseAmount(navState.amount ?? profile.current_due)
            if (due > 0) {
              const prefill = Math.min(requested > 0 ? requested : due, due)
              setAmount(prefill.toFixed(2))
            }
          }
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

        if (err instanceof Error && err.message.includes('user_id')) {
          clearAuth()
          navigate('/login', { replace: true })
          return
        }

        if (err instanceof ApiError) {
          setLoadError(err.message)
        } else if (err instanceof Error) {
          setLoadError(err.message)
        } else {
          setLoadError('Unable to load your account.')
        }
      } finally {
        if (!cancelled) {
          setLoading(false)
        }
      }
    }

    void loadUser()

    return () => {
      cancelled = true
    }
  }, [navigate, location.state])

  function validate(currentDue: string): string | null {
    if (!hasOutstandingDue(currentDue)) {
      return 'You have no outstanding due to repay.'
    }

    const parsedAmount = Number(amount)
    if (!amount.trim() || Number.isNaN(parsedAmount)) {
      return 'Enter a valid payment amount.'
    }

    if (parsedAmount <= 0) {
      return 'Amount must be greater than zero.'
    }

    const due = parseAmount(currentDue)
    if (parsedAmount > due) {
      return `Amount cannot exceed outstanding due (${formatCurrency(due)}).`
    }

    return null
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()

    if (submitting || !user) {
      return
    }

    const validationError = validate(user.current_due)
    if (validationError) {
      setFormError(validationError)
      setSuccessMessage(null)
      return
    }

    const token = getToken()
    if (!token) {
      clearAuth()
      navigate('/login', { replace: true })
      return
    }

    setFormError(null)
    setSuccessMessage(null)
    setSubmitting(true)

    try {
      const result = await createPayment(Number(amount), token)
      setSuccessMessage(result.message)
      setAmount('')

      try {
        const userId = getUserIdFromToken(token)
        const profile = await getCurrentUser(userId, token)
        setUser(profile)
        await refreshAccount()
      } catch (refreshErr) {
        if (refreshErr instanceof ApiError && refreshErr.status === 401) {
          clearAuth()
          navigate('/login', { replace: true })
          return
        }

        setFormError(
          refreshErr instanceof ApiError
            ? `Payment succeeded, but due could not be refreshed: ${refreshErr.message}`
            : 'Payment succeeded, but due could not be refreshed.',
        )
      }
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        clearAuth()
        navigate('/login', { replace: true })
        return
      }

      if (err instanceof ApiError) {
        setFormError(err.message)
      } else if (err instanceof Error) {
        setFormError(err.message)
      } else {
        setFormError('Unable to complete the payment.')
      }
    } finally {
      setSubmitting(false)
    }
  }

  function handleAmountChange(event: ChangeEvent<HTMLInputElement>) {
    setAmount(event.target.value)
    if (formError) {
      setFormError(null)
    }
  }

  function fillFullDue() {
    if (!user) return
    const due = parseAmount(user.current_due)
    if (due > 0) {
      setAmount(due.toFixed(2))
      setFormError(null)
    }
  }

  const canPay = user ? hasOutstandingDue(user.current_due) : false

  return (
    <div className="pl-page">
      <header className="pl-page__intro">
        <h2>Make a Payment</h2>
        <p>Repay part or all of your outstanding PayLater balance.</p>
      </header>

      <section className="pl-card" aria-labelledby="payment-form-heading">
        <h3 id="payment-form-heading" className="pl-card__title">
          Payment details
        </h3>

        {loading ? (
          <p className="pl-status" role="status">
            Loading your account…
          </p>
        ) : null}

        {!loading && loadError ? (
          <p className="pl-error" role="alert">
            {loadError}
          </p>
        ) : null}

        {!loading && !loadError && user ? (
          <>
            <div className="pl-grid pl-grid--2" style={{ marginBottom: 16 }}>
              <article>
                <p className="pl-metric__label">Outstanding Due</p>
                <p className="pl-metric__value pl-metric__value--due">
                  {formatCurrency(user.current_due)}
                </p>
              </article>
              <article>
                <p className="pl-metric__label">Amount to Pay</p>
                <p className="pl-metric__value">
                  {amount ? formatCurrency(amount) : formatCurrency(0)}
                </p>
              </article>
            </div>

            {successMessage ? (
              <div className="pl-success" role="status">
                <p>{successMessage}</p>
                <p>
                  Remaining due:{' '}
                  <strong>{formatCurrency(user.current_due)}</strong>
                </p>
              </div>
            ) : null}

            <form className="pl-form" onSubmit={handleSubmit} noValidate>
              <div className="pl-field">
                <label htmlFor="payment-amount">Payment amount</label>
                <input
                  id="payment-amount"
                  name="amount"
                  type="number"
                  inputMode="decimal"
                  min="0.01"
                  step="0.01"
                  max={parseAmount(user.current_due) || undefined}
                  value={amount}
                  onChange={handleAmountChange}
                  placeholder="0.00"
                  disabled={submitting || !canPay}
                  aria-invalid={formError !== null}
                  aria-describedby={formError ? 'payment-error' : undefined}
                />
              </div>

              {!canPay ? (
                <p className="pl-status" role="status">
                  You have no outstanding due to repay.
                </p>
              ) : null}

              {formError ? (
                <p id="payment-error" className="pl-error" role="alert">
                  {formError}
                </p>
              ) : null}

              <div className="pl-actions">
                {canPay ? (
                  <button
                    type="button"
                    className="pl-btn"
                    onClick={fillFullDue}
                    disabled={submitting}
                  >
                    Pay Full Due
                  </button>
                ) : null}
                <button
                  type="submit"
                  className="pl-btn pl-btn--primary"
                  disabled={submitting || !canPay}
                >
                  {submitting ? 'Processing…' : 'Make Payment'}
                </button>
              </div>
            </form>
          </>
        ) : null}
      </section>
    </div>
  )
}

export default PaymentPage
