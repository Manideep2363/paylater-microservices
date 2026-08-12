import { useState, type FormEvent, type ChangeEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  getPostLoginPath,
  saveAuth,
  type AuthRole,
} from '../auth/authStorage'
import {
  AuthApiError,
  loginAdmin,
  loginMerchant,
  loginUser,
} from '../services/authService'
import PasswordInput from '../components/PasswordInput'
import './LoginPage.css'

type AccountType = AuthRole

const ACCOUNT_TYPES: { value: AccountType; label: string }[] = [
  { value: 'user', label: 'User' },
  { value: 'merchant', label: 'Merchant' },
  { value: 'admin', label: 'Admin' },
]

const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

function LoginPage() {
  const navigate = useNavigate()
  const [accountType, setAccountType] = useState<AccountType>('user')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  function validate(): string | null {
    const trimmedEmail = email.trim()

    if (!trimmedEmail) {
      return 'Email is required.'
    }

    if (!EMAIL_PATTERN.test(trimmedEmail)) {
      return 'Enter a valid email address.'
    }

    if (!password) {
      return 'Password is required.'
    }

    if (password.length < 6) {
      return 'Password must be at least 6 characters.'
    }

    return null
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()

    const validationError = validate()
    if (validationError) {
      setError(validationError)
      return
    }

    setError(null)
    setLoading(true)

    try {
      const trimmedEmail = email.trim()
      let result

      if (accountType === 'user') {
        result = await loginUser(trimmedEmail, password)
      } else if (accountType === 'merchant') {
        result = await loginMerchant(trimmedEmail, password)
      } else {
        result = await loginAdmin(trimmedEmail, password)
      }

      saveAuth(result.token, accountType)
      navigate(getPostLoginPath(accountType), { replace: true })
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

  function handleEmailChange(event: ChangeEvent<HTMLInputElement>) {
    setEmail(event.target.value)
    if (error) {
      setError(null)
    }
  }

  function handlePasswordChange(event: ChangeEvent<HTMLInputElement>) {
    setPassword(event.target.value)
    if (error) {
      setError(null)
    }
  }

  return (
    <div className="login-page">
      <main className="login-page__card">
        <header className="login-page__header">
          <p className="login-page__brand">PayLater</p>
          <h1 className="login-page__title">Welcome back</h1>
          <p className="login-page__subtitle">
            Sign in to manage purchases, payments, and credit.
          </p>
        </header>

        <form className="login-page__form" onSubmit={handleSubmit} noValidate>
          <fieldset className="login-page__fieldset">
            <legend className="login-page__legend">Account type</legend>
            <div
              className="login-page__account-types"
              role="radiogroup"
              aria-label="Account type"
            >
              {ACCOUNT_TYPES.map((type) => (
                <label
                  key={type.value}
                  className={
                    accountType === type.value
                      ? 'login-page__account-option login-page__account-option--active'
                      : 'login-page__account-option'
                  }
                >
                  <input
                    type="radio"
                    name="accountType"
                    value={type.value}
                    checked={accountType === type.value}
                    onChange={() => setAccountType(type.value)}
                  />
                  {type.label}
                </label>
              ))}
            </div>
          </fieldset>

          <div className="login-page__field">
            <label htmlFor="login-email">Email</label>
            <input
              id="login-email"
              name="email"
              type="email"
              autoComplete="email"
              value={email}
              onChange={handleEmailChange}
              placeholder="you@example.com"
              disabled={loading}
              aria-invalid={error !== null}
              aria-describedby={error ? 'login-error' : undefined}
            />
          </div>

          <PasswordInput
            className="login-page__field"
            label="Password"
            id="login-password"
            name="password"
            autoComplete="current-password"
            value={password}
            onChange={handlePasswordChange}
            placeholder="Enter your password"
            disabled={loading}
            aria-invalid={error !== null}
            aria-describedby={error ? 'login-error' : undefined}
          />

          {error ? (
            <p id="login-error" className="login-page__error" role="alert">
              {error}
            </p>
          ) : null}

          <button
            type="submit"
            className="login-page__submit"
            disabled={loading}
          >
            {loading ? 'Signing in…' : 'Sign in'}
          </button>
        </form>

        <footer className="login-page__footer">
          <p>
            New here?{' '}
            <button
              type="button"
              className="login-page__link"
              onClick={() => navigate('/register')}
            >
              Create a user account
            </button>
          </p>
          <p>
            Selling with PayLater?{' '}
            <button
              type="button"
              className="login-page__link"
              onClick={() => navigate('/merchant/register')}
            >
              Register as a merchant
            </button>
          </p>
        </footer>
      </main>
    </div>
  )
}

export default LoginPage
