import { useState, type FormEvent, type ChangeEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { AuthApiError, registerMerchant } from '../services/authService'
import PasswordInput from '../components/PasswordInput'
import './MerchantRegisterPage.css'

const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
const PHONE_PATTERN = /^[+\d][\d\s()-]{6,}$/

function MerchantRegisterPage() {
  const navigate = useNavigate()
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [phone, setPhone] = useState('')
  const [password, setPassword] = useState('')
  const [commissionPercentage, setCommissionPercentage] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [successMessage, setSuccessMessage] = useState<string | null>(null)

  function validate(): string | null {
    const trimmedName = name.trim()
    const trimmedEmail = email.trim()
    const trimmedPhone = phone.trim()
    const parsedCommission = Number(commissionPercentage)

    if (!trimmedName) {
      return 'Business name is required.'
    }

    if (!trimmedEmail) {
      return 'Email is required.'
    }

    if (!EMAIL_PATTERN.test(trimmedEmail)) {
      return 'Enter a valid email address.'
    }

    if (!trimmedPhone) {
      return 'Phone is required.'
    }

    if (!PHONE_PATTERN.test(trimmedPhone)) {
      return 'Enter a valid phone number.'
    }

    if (!password) {
      return 'Password is required.'
    }

    if (password.length < 6) {
      return 'Password must be at least 6 characters.'
    }

    if (!commissionPercentage.trim() || Number.isNaN(parsedCommission)) {
      return 'Enter a valid commission percentage.'
    }

    if (parsedCommission < 0) {
      return 'Commission percentage cannot be negative.'
    }

    return null
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()

    const validationError = validate()
    if (validationError) {
      setError(validationError)
      setSuccessMessage(null)
      return
    }

    setError(null)
    setSuccessMessage(null)
    setLoading(true)

    try {
      const result = await registerMerchant(
        name.trim(),
        email.trim(),
        phone.trim(),
        password,
        Number(commissionPercentage),
      )
      setSuccessMessage(result.message)
      setName('')
      setEmail('')
      setPhone('')
      setPassword('')
      setCommissionPercentage('')
    } catch (err) {
      if (err instanceof AuthApiError) {
        setError(err.message)
      } else {
        setError('Something went wrong. Please try again.')
      }
    } finally {
      setLoading(false)
    }
  }

  function clearMessages() {
    if (error) {
      setError(null)
    }
    if (successMessage) {
      setSuccessMessage(null)
    }
  }

  function handleNameChange(event: ChangeEvent<HTMLInputElement>) {
    setName(event.target.value)
    clearMessages()
  }

  function handleEmailChange(event: ChangeEvent<HTMLInputElement>) {
    setEmail(event.target.value)
    clearMessages()
  }

  function handlePhoneChange(event: ChangeEvent<HTMLInputElement>) {
    setPhone(event.target.value)
    clearMessages()
  }

  function handlePasswordChange(event: ChangeEvent<HTMLInputElement>) {
    setPassword(event.target.value)
    clearMessages()
  }

  function handleCommissionChange(event: ChangeEvent<HTMLInputElement>) {
    setCommissionPercentage(event.target.value)
    clearMessages()
  }

  return (
    <div className="merchant-register-page">
      <main className="merchant-register-page__card">
        <header className="merchant-register-page__header">
          <p className="merchant-register-page__brand">PayLater</p>
          <h1 className="merchant-register-page__title">
            Register as a merchant
          </h1>
          <p className="merchant-register-page__subtitle">
            Create a merchant account to accept PayLater purchases.
          </p>
        </header>

        {successMessage ? (
          <div className="merchant-register-page__success" role="status">
            <p>{successMessage}</p>
            <button
              type="button"
              className="merchant-register-page__submit"
              onClick={() => navigate('/login')}
            >
              Go to Merchant Login
            </button>
          </div>
        ) : (
          <form
            className="merchant-register-page__form"
            onSubmit={handleSubmit}
            noValidate
          >
            <div className="merchant-register-page__field">
              <label htmlFor="merchant-register-name">
                Merchant / Business Name
              </label>
              <input
                id="merchant-register-name"
                name="name"
                type="text"
                autoComplete="organization"
                value={name}
                onChange={handleNameChange}
                placeholder="ABC Store"
                disabled={loading}
                aria-invalid={error !== null}
              />
            </div>

            <div className="merchant-register-page__field">
              <label htmlFor="merchant-register-email">Email</label>
              <input
                id="merchant-register-email"
                name="email"
                type="email"
                autoComplete="email"
                value={email}
                onChange={handleEmailChange}
                placeholder="merchant@example.com"
                disabled={loading}
                aria-invalid={error !== null}
              />
            </div>

            <div className="merchant-register-page__field">
              <label htmlFor="merchant-register-phone">Phone</label>
              <input
                id="merchant-register-phone"
                name="phone"
                type="tel"
                autoComplete="tel"
                value={phone}
                onChange={handlePhoneChange}
                placeholder="+91 98765 43210"
                disabled={loading}
                aria-invalid={error !== null}
              />
            </div>

            <PasswordInput
              className="merchant-register-page__field"
              label="Password"
              id="merchant-register-password"
              name="password"
              autoComplete="new-password"
              value={password}
              onChange={handlePasswordChange}
              placeholder="At least 6 characters"
              disabled={loading}
              aria-invalid={error !== null}
            />

            <div className="merchant-register-page__field">
              <label htmlFor="merchant-register-commission">
                Commission Percentage
              </label>
              <input
                id="merchant-register-commission"
                name="commission_percentage"
                type="number"
                inputMode="decimal"
                min="0"
                step="0.01"
                value={commissionPercentage}
                onChange={handleCommissionChange}
                placeholder="5.00"
                disabled={loading}
                aria-invalid={error !== null}
                aria-describedby={error ? 'merchant-register-error' : undefined}
              />
            </div>

            {error ? (
              <p
                id="merchant-register-error"
                className="merchant-register-page__error"
                role="alert"
              >
                {error}
              </p>
            ) : null}

            <button
              type="submit"
              className="merchant-register-page__submit"
              disabled={loading}
            >
              {loading ? 'Creating account…' : 'Create merchant account'}
            </button>
          </form>
        )}

        <footer className="merchant-register-page__footer">
          <p>
            Already registered?{' '}
            <button
              type="button"
              className="merchant-register-page__link"
              onClick={() => navigate('/login')}
            >
              Sign in
            </button>
          </p>
        </footer>
      </main>
    </div>
  )
}

export default MerchantRegisterPage
