export type AuthRole = 'user' | 'merchant' | 'admin'

export type AuthSession = {
  token: string
  role: AuthRole
}

const AUTH_STORAGE_KEY = 'paylater.auth'

function isAuthRole(value: unknown): value is AuthRole {
  return value === 'user' || value === 'merchant' || value === 'admin'
}

export function saveAuth(token: string, role: AuthRole): void {
  const session: AuthSession = { token, role }
  sessionStorage.setItem(AUTH_STORAGE_KEY, JSON.stringify(session))
}

export function getAuth(): AuthSession | null {
  const raw = sessionStorage.getItem(AUTH_STORAGE_KEY)
  if (!raw) {
    return null
  }

  try {
    const parsed = JSON.parse(raw) as Partial<AuthSession>
    if (
      typeof parsed.token === 'string' &&
      parsed.token.length > 0 &&
      isAuthRole(parsed.role)
    ) {
      return { token: parsed.token, role: parsed.role }
    }
  } catch {
    // Corrupt session data — treat as logged out.
  }

  clearAuth()
  return null
}

export function getToken(): string | null {
  return getAuth()?.token ?? null
}

export function getRole(): AuthRole | null {
  return getAuth()?.role ?? null
}

export function isAuthenticated(): boolean {
  return getAuth() !== null
}

export function clearAuth(): void {
  sessionStorage.removeItem(AUTH_STORAGE_KEY)
}

export function getPostLoginPath(role: AuthRole): string {
  switch (role) {
    case 'user':
      return '/dashboard'
    case 'merchant':
      return '/merchant/dashboard'
    case 'admin':
      return '/admin/dashboard'
  }
}
