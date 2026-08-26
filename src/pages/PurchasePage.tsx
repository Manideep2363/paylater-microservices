import { useEffect, useState, type FormEvent, type ChangeEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAccount } from '../auth/AccountContext'
import { clearAuth, getToken } from '../auth/authStorage'
import { getUserIdFromToken } from '../auth/jwt'
import { ApiError } from '../services/http'
import { getMerchants, type MerchantOption } from '../services/merchantService'
import { createPurchase } from '../services/purchaseService'
import { getCurrentUser } from '../services/userService'
import { formatCurrency } from '../utils/money'

function PurchasePage() {
  const navigate = useNavigate()
  const { refresh: refreshAccount } = useAccount()

  const [merchants, setMerchants] = useState<MerchantOption[]>([])
  const [merchantsLoading, setMerchantsLoading] = useState(true)
  const [merchantsError, setMerchantsError] = useState<string | null>(null)

  const [merchantId, setMerchantId] = useState('')
  const [amount, setAmount] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [formError, setFormError] = useState<string | null>(null)
  const [successMessage, setSuccessMessage] = useState<string | null>(null)
  const [updatedCurrentDue, setUpdatedCurrentDue] = useState<string | null>(
    null,
  )

  useEffect(() => {
    let cancelled = false

    async function loadMerchants() {
      const token = getToken()
      if (!token) {
        clearAuth()
        navigate('/login', { replace: true })
        return
      }

      setMerchantsLoading(true)
      setMerchantsError(null)

      try {
        const list = await getMerchants(token)
        if (!cancelled) {
          setMerchants(list)
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
          setMerchantsError(err.message)
        } else if (err instanceof Error) {
          setMerchantsError(err.message)
        } else {
          setMerchantsError('Unable to load merchants.')
        }
      } finally {
        if (!cancelled) {
          setMerchantsLoading(false)
        }
      }
    }

    void loadMerchants()

    return () => {
      cancelled = true
    }
  }, [navigate])

  function validate(): string | null {
    if (!merchantId) {
      return 'Please select a merchant.'
    }

    const parsedAmount = Number(amount)
    if (!amount.trim() || Number.isNaN(parsedAmount)) {
      return 'Enter a valid purchase amount.'
    }

    if (parsedAmount <= 0) {
      return 'Amount must be greater than zero.'
    }

    return null
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()

    if (submitting) {
      return
    }

    const validationError = validate()
    if (validationError) {
      setFormError(validationError)
      setSuccessMessage(null)
      setUpdatedCurrentDue(null)
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
    setUpdatedCurrentDue(null)
    setSubmitting(true)

    try {
      const result = await createPurchase(
        Number(merchantId),
        Number(amount),
        token,
      )

      setSuccessMessage(result.message)
      setAmount('')

      try {
        const userId = getUserIdFromToken(token)
        const profile = await getCurrentUser(userId, token)
        setUpdatedCurrentDue(profile.current_due)
        await refreshAccount()
      } catch (refreshErr) {
        if (refreshErr instanceof ApiError && refreshErr.status === 401) {
          clearAuth()
          navigate('/login', { replace: true })
          return
        }

        setUpdatedCurrentDue(null)
        setFormError(
          refreshErr instanceof ApiError
            ? `Purchase succeeded, but due could not be refreshed: ${refreshErr.message}`
            : 'Purchase succeeded, but due could not be refreshed.',
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
        setFormError('Unable to complete the purchase.')
      }
    } finally {
      setSubmitting(false)
    }
  }

  function handleMerchantChange(event: ChangeEvent<HTMLSelectElement>) {
    setMerchantId(event.target.value)
    if (formError) {
      setFormError(null)
    }
  }

  function handleAmountChange(event: ChangeEvent<HTMLInputElement>) {
    setAmount(event.target.value)
    if (formError) {
      setFormError(null)
    }
  }

  return (
    <div className="pl-page">
      <header className="pl-page__intro">
        <h2>Make a Purchase</h2>
        <p>Buy now from a merchant and pay later within your credit limit.</p>
      </header>

      <section className="pl-card" aria-labelledby="purchase-form-heading">
        <h3 id="purchase-form-heading" className="pl-card__title">
          Purchase details
        </h3>

        {merchantsLoading ? (
          <p className="pl-status" role="status">
            Loading merchants…
          </p>
        ) : null}

        {!merchantsLoading && merchantsError ? (
          <p className="pl-error" role="alert">
            {merchantsError}
          </p>
        ) : null}

        {!merchantsLoading && !merchantsError ? (
          <form className="pl-form" onSubmit={handleSubmit} noValidate>
            <div className="pl-field">
              <label htmlFor="purchase-merchant">Merchant</label>
              <select
                id="purchase-merchant"
                name="merchant_id"
                value={merchantId}
                onChange={handleMerchantChange}
                disabled={submitting || merchants.length === 0}
                aria-invalid={formError !== null}
              >
                <option value="">Select a merchant</option>
                {merchants.map((merchant) => (
                  <option
                    key={merchant.merchant_id}
                    value={merchant.merchant_id}
                  >
                    {merchant.name}
                  </option>
                ))}
              </select>
            </div>

            {merchants.length === 0 ? (
              <p className="pl-status" role="status">
                No merchants are available right now.
              </p>
            ) : null}

            <div className="pl-field">
              <label htmlFor="purchase-amount">Amount</label>
              <input
                id="purchase-amount"
                name="amount"
                type="number"
                inputMode="decimal"
                min="0.01"
                step="0.01"
                value={amount}
                onChange={handleAmountChange}
                placeholder="0.00"
                disabled={submitting}
                aria-invalid={formError !== null}
                aria-describedby={formError ? 'purchase-error' : undefined}
              />
            </div>

            {formError ? (
              <p id="purchase-error" className="pl-error" role="alert">
                {formError}
              </p>
            ) : null}

            {successMessage ? (
              <div className="pl-success" role="status">
                <p>{successMessage}</p>
                {updatedCurrentDue !== null ? (
                  <p>
                    Updated current due:{' '}
                    <strong>{formatCurrency(updatedCurrentDue)}</strong>
                  </p>
                ) : null}
              </div>
            ) : null}

            <button
              type="submit"
              className="pl-btn pl-btn--primary"
              disabled={submitting || merchants.length === 0}
            >
              {submitting ? 'Processing…' : 'Purchase'}
            </button>
          </form>
        ) : null}
      </section>
    </div>
  )
}

export default PurchasePage
