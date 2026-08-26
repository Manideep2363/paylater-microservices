/**
 * Client-side JWT payload decode for reading claims used in routing/API calls.
 * This is NOT authorization — the backend still validates every request.
 */

export type JwtPayload = {
  user_id?: number
  email?: string
  role?: string
  exp?: number
  iat?: number
}

function decodeBase64Url(value: string): string {
  const normalized = value.replace(/-/g, '+').replace(/_/g, '/')
  const padding =
    normalized.length % 4 === 0 ? '' : '='.repeat(4 - (normalized.length % 4))
  return atob(normalized + padding)
}

export function decodeJwtPayload(token: string): JwtPayload {
  const parts = token.split('.')
  if (parts.length < 2 || !parts[1]) {
    throw new Error('Invalid token format.')
  }

  try {
    const json = decodeBase64Url(parts[1])
    return JSON.parse(json) as JwtPayload
  } catch {
    throw new Error('Unable to decode token payload.')
  }
}

export function getUserIdFromToken(token: string): number {
  const payload = decodeJwtPayload(token)
  const userId = payload.user_id

  if (typeof userId !== 'number' || !Number.isInteger(userId) || userId <= 0) {
    throw new Error('Token does not contain a valid user_id.')
  }

  return userId
}
