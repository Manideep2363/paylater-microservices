import { useEffect, useState, type FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import PasswordInput from '../components/PasswordInput'
import { clearAuth, getToken } from '../auth/authStorage'
import { ApiError } from '../services/http'
import {
  createAdminMerchant,
  getAdminMerchantById,
  getAdminMerchants,
  updateMerchantCommission,
  type AdminMerchant,
} from '../services/adminService'
import '../components/AdminShared.css'

function AdminMerchantsPage() {
  const navigate = useNavigate()
  const [merchants, setMerchants] = useState<AdminMerchant[]>([])
  const [selected, setSelected] = useState<AdminMerchant | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState<string | null>(null)
  const [showCreate, setShowCreate] = useState(false)
  const [submitting, setSubmitting] = useState(false)

  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [phone, setPhone] = useState('')
  const [password, setPassword] = useState('')
  const [commission, setCommission] = useState('')
  const [commissionUpdate, setCommissionUpdate] = useState('')

  async function loadMerchants() {
    const token = getToken()
    if (!token) {
      clearAuth()
      navigate('/login', { replace: true })
      return
    }

    setLoading(true)
    setError(null)

    try {
      setMerchants(await getAdminMerchants(token))
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        clearAuth()
        navigate('/login', { replace: true })
        return
      }
      setError(
        err instanceof ApiError ? err.message : 'Unable to load merchants.',
      )
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void loadMerchants()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [navigate])

  async function handleViewDetails(id: number) {
    const token = getToken()
    if (!token) {
      clearAuth()
      navigate('/login', { replace: true })
      return
    }

    setError(null)
    setSuccess(null)

    try {
      const merchant = await getAdminMerchantById(id, token)
      setSelected(merchant)
      setCommissionUpdate(merchant.commission_percentage)
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        clearAuth()
        navigate('/login', { replace: true })
        return
      }
      setError(
        err instanceof ApiError ? err.message : 'Unable to load merchant.',
      )
    }
  }

  async function handleCreate(event: FormEvent) {
    event.preventDefault()
    const token = getToken()
    if (!token) {
      clearAuth()
      navigate('/login', { replace: true })
      return
    }

    setSubmitting(true)
    setError(null)
    setSuccess(null)

    try {
      await createAdminMerchant(
        {
          name: name.trim(),
          email: email.trim(),
          phone: phone.trim(),
          password,
          commission: Number(commission),
        },
        token,
      )
      setSuccess('Merchant created successfully.')
      setName('')
      setEmail('')
      setPhone('')
      setPassword('')
      setCommission('')
      setShowCreate(false)
      await loadMerchants()
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        clearAuth()
        navigate('/login', { replace: true })
        return
      }
      setError(
        err instanceof ApiError ? err.message : 'Unable to create merchant.',
      )
    } finally {
      setSubmitting(false)
    }
  }

  async function handleCommissionUpdate(event: FormEvent) {
    event.preventDefault()
    if (!selected) return

    const token = getToken()
    if (!token) {
      clearAuth()
      navigate('/login', { replace: true })
      return
    }

    setSubmitting(true)
    setError(null)
    setSuccess(null)

    try {
      const result = await updateMerchantCommission(
        selected.merchant_id,
        Number(commissionUpdate),
        token,
      )
      setSuccess(result.message)
      await handleViewDetails(selected.merchant_id)
      await loadMerchants()
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        clearAuth()
        navigate('/login', { replace: true })
        return
      }
      setError(
        err instanceof ApiError
          ? err.message
          : 'Unable to update commission.',
      )
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="pl-page">
      <header className="pl-page__intro">
        <h2>Merchants</h2>
        <p>Manage merchant accounts.</p>
      </header>

      <section className="pl-card">
        <div className="pl-card__header">
          <h3 className="pl-card__title">All merchants</h3>
          <button
            type="button"
            className="pl-btn pl-btn--primary"
            onClick={() => setShowCreate((value) => !value)}
          >
            {showCreate ? 'Hide form' : 'Create Merchant'}
          </button>
        </div>

        {success ? <p className="pl-success">{success}</p> : null}
        {error ? (
          <p className="pl-error" role="alert">
            {error}
          </p>
        ) : null}

        {showCreate ? (
          <form className="admin-form" onSubmit={handleCreate}>
            <div className="admin-field">
              <label htmlFor="admin-merchant-name">Name</label>
              <input
                id="admin-merchant-name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                required
                disabled={submitting}
              />
            </div>
            <div className="admin-field">
              <label htmlFor="admin-merchant-email">Email</label>
              <input
                id="admin-merchant-email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
                disabled={submitting}
              />
            </div>
            <div className="admin-field">
              <label htmlFor="admin-merchant-phone">Phone</label>
              <input
                id="admin-merchant-phone"
                value={phone}
                onChange={(e) => setPhone(e.target.value)}
                required
                disabled={submitting}
              />
            </div>
            <PasswordInput
              className="admin-field"
              label="Password"
              id="admin-merchant-password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              minLength={6}
              required
              disabled={submitting}
            />
            <div className="admin-field">
              <label htmlFor="admin-merchant-commission">Commission</label>
              <input
                id="admin-merchant-commission"
                type="number"
                step="0.01"
                value={commission}
                onChange={(e) => setCommission(e.target.value)}
                required
                disabled={submitting}
              />
            </div>
            <button
              type="submit"
              className="pl-btn pl-btn--primary"
              disabled={submitting}
            >
              {submitting ? 'Creating…' : 'Create merchant'}
            </button>
          </form>
        ) : null}

        {loading ? (
          <p className="pl-status" role="status">
            Loading merchants...
          </p>
        ) : null}

        {!loading && merchants.length === 0 ? (
          <p className="pl-empty">No merchants found.</p>
        ) : null}

        {!loading && merchants.length > 0 ? (
          <div className="pl-table-wrap">
            <table className="pl-table">
              <thead>
                <tr>
                  <th scope="col">Merchant</th>
                  <th scope="col">Email</th>
                  <th scope="col">Phone</th>
                  <th scope="col">Commission</th>
                  <th scope="col">Actions</th>
                </tr>
              </thead>
              <tbody>
                {merchants.map((merchant) => (
                  <tr key={merchant.merchant_id}>
                    <td>{merchant.name}</td>
                    <td>{merchant.email}</td>
                    <td>{merchant.phone}</td>
                    <td>{merchant.commission_percentage}%</td>
                    <td>
                      <div className="pl-actions">
                        <button
                          type="button"
                          className="pl-btn"
                          onClick={() =>
                            void handleViewDetails(merchant.merchant_id)
                          }
                        >
                          View
                        </button>
                        <button
                          type="button"
                          className="pl-btn"
                          onClick={() =>
                            void handleViewDetails(merchant.merchant_id)
                          }
                        >
                          Update Commission
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        ) : null}
      </section>

      {selected ? (
        <section className="pl-card" aria-labelledby="merchant-detail-heading">
          <h3 id="merchant-detail-heading" className="pl-card__title">
            Merchant details
          </h3>
          <dl className="pl-details">
            <div>
              <dt>Merchant ID</dt>
              <dd>{selected.merchant_id}</dd>
            </div>
            <div>
              <dt>Name</dt>
              <dd>{selected.name}</dd>
            </div>
            <div>
              <dt>Email</dt>
              <dd>{selected.email}</dd>
            </div>
            <div>
              <dt>Phone</dt>
              <dd>{selected.phone}</dd>
            </div>
            <div>
              <dt>Commission %</dt>
              <dd>{selected.commission_percentage}</dd>
            </div>
          </dl>

          <form className="admin-form" onSubmit={handleCommissionUpdate}>
            <div className="admin-field">
              <label htmlFor="admin-update-commission">Update Commission</label>
              <input
                id="admin-update-commission"
                type="number"
                step="0.01"
                value={commissionUpdate}
                onChange={(e) => setCommissionUpdate(e.target.value)}
                required
                disabled={submitting}
              />
            </div>
            <button
              type="submit"
              className="pl-btn pl-btn--primary"
              disabled={submitting}
            >
              {submitting ? 'Updating…' : 'Update Commission'}
            </button>
          </form>
        </section>
      ) : null}
    </div>
  )
}

export default AdminMerchantsPage
