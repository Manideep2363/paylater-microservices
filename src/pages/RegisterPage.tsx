import { useState, type FormEvent, type ChangeEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { AuthApiError, registerUser } from '../services/authService'
import PasswordInput from '../components/PasswordInput'
import './RegisterPage.css'

const EMAIL_PATTERN = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

function RegisterPage() {
  const navigate = useNavigate()
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [successMessage, setSuccessMessage] = useState<string | null>(null)

  function validate(): string | null {
    const trimmedName = name.trim()
    const trimmedEmail = email.trim()

    if (!trimmedName) {
      return 'Full name is required.'
    }

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
      setSuccessMessage(null)
      return
    }

    setError(null)
    setSuccessMessage(null)
    setLoading(true)

    try {
      const result = await registerUser(name.trim(), email.trim(), password)
      setSuccessMessage(result.message)
      setName('')
      setEmail('')
      setPassword('')
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

  function handlePasswordChange(event: ChangeEvent<HTMLInputElement>) {
    setPassword(event.target.value)
    clearMessages()
  }

  return (
    <div className="register-page">
      <main className="register-page__card">
        <header className="register-page__header">
          <p className="register-page__brand">PayLater</p>
          <h1 className="register-page__title">Create your account</h1>
          <p className="register-page__subtitle">
            Register to buy now and pay later with PayLater.
          </p>
        </header>

        {successMessage ? (
          <div className="register-page__success" role="status">
            <p>{successMessage}</p>
            <button
              type="button"
              className="register-page__submit"
              onClick={() => navigate('/login')}
            >
              Go to Login
            </button>
          </div>
        ) : (
          <form
            className="register-page__form"
            onSubmit={handleSubmit}
            noValidate
          >
            <div className="register-page__field">
              <label htmlFor="register-name">Full Name</label>
              <input
                id="register-name"
                name="name"
                type="text"
                autoComplete="name"
                value={name}
                onChange={handleNameChange}
                placeholder="Jane Doe"
                disabled={loading}
                aria-invalid={error !== null}
                aria-describedby={error ? 'register-error' : undefined}
              />
            </div>

            <div className="register-page__field">
              <label htmlFor="register-email">Email</label>
              <input
                id="register-email"
                name="email"
                type="email"
                autoComplete="email"
                value={email}
                onChange={handleEmailChange}
                placeholder="you@example.com"
                disabled={loading}
                aria-invalid={error !== null}
                aria-describedby={error ? 'register-error' : undefined}
              />
            </div>

            <PasswordInput
              className="register-page__field"
              label="Password"
              id="register-password"
              name="password"
              autoComplete="new-password"
              value={password}
              onChange={handlePasswordChange}
              placeholder="At least 6 characters"
              disabled={loading}
              aria-invalid={error !== null}
              aria-describedby={error ? 'register-error' : undefined}
            />

            {error ? (
              <p id="register-error" className="register-page__error" role="alert">
                {error}
              </p>
            ) : null}

            <button
              type="submit"
              className="register-page__submit"
              disabled={loading}
            >
              {loading ? 'Creating account…' : 'Create account'}
            </button>
          </form>
        )}

        <footer className="register-page__footer">
          <p>
            Already have an account?{' '}
            <button
              type="button"
              className="register-page__link"
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

export default RegisterPage
