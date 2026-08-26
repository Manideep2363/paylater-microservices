import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react'
import { useNavigate } from 'react-router-dom'
import { clearAuth, getRole, getToken, type AuthRole } from '../auth/authStorage'
import { getUserIdFromToken } from '../auth/jwt'
import { ApiError } from '../services/http'
import {
  getMerchantProfile,
  type MerchantProfile,
} from '../services/merchantService'
import { getCurrentUser, type User } from '../services/userService'

type AccountContextValue = {
  role: AuthRole | null
  user: User | null
  merchant: MerchantProfile | null
  displayName: string
  email: string | null
  loading: boolean
  error: string | null
  refresh: () => Promise<void>
}

const AccountContext = createContext<AccountContextValue | null>(null)

export function AccountProvider({ children }: { children: ReactNode }) {
  const navigate = useNavigate()
  const role = getRole()
  const [user, setUser] = useState<User | null>(null)
  const [merchant, setMerchant] = useState<MerchantProfile | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const refresh = useCallback(async () => {
    const token = getToken()
    if (!token) {
      clearAuth()
      navigate('/login', { replace: true })
      return
    }

    setLoading(true)
    setError(null)

    try {
      if (role === 'user') {
        const userId = getUserIdFromToken(token)
        const profile = await getCurrentUser(userId, token)
        setUser(profile)
        setMerchant(null)
      } else if (role === 'merchant') {
        const profile = await getMerchantProfile(token)
        setMerchant(profile)
        setUser(null)
      } else {
        setUser(null)
        setMerchant(null)
      }
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        clearAuth()
        navigate('/login', { replace: true })
        return
      }
      setError(
        err instanceof ApiError ? err.message : 'Unable to load account profile.',
      )
    } finally {
      setLoading(false)
    }
  }, [navigate, role])

  useEffect(() => {
    void refresh()
  }, [refresh])

  const value = useMemo<AccountContextValue>(() => {
    const displayName =
      role === 'admin'
        ? 'Admin'
        : user?.name || merchant?.name || 'Account'

    const email = user?.email || merchant?.email || null

    return {
      role,
      user,
      merchant,
      displayName,
      email,
      loading,
      error,
      refresh,
    }
  }, [role, user, merchant, loading, error, refresh])

  return (
    <AccountContext.Provider value={value}>{children}</AccountContext.Provider>
  )
}

export function useAccount() {
  const context = useContext(AccountContext)
  if (!context) {
    throw new Error('useAccount must be used within AccountProvider')
  }
  return context
}
