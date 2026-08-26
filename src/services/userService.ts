import { apiRequest } from './http'

/**
 * Matches user-service publicUserResponse (GET /users/:id).
 * credit_limit and current_due are DECIMAL values serialized as strings.
 */
export type User = {
  user_id: number
  name: string
  email: string
  credit_limit: string
  current_due: string
}

export function getCurrentUser(
  userId: number,
  token: string,
): Promise<User> {
  return apiRequest<User>(`/users/${userId}`, { token })
}
