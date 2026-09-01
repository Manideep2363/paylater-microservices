const API_BASE_URL = '/api'

type BackendErrorBody = {
  error?: string
  message?: string
}

export class ApiError extends Error {
  status: number

  constructor(message: string, status: number) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

async function readErrorMessage(
  response: Response,
  fallbackMessage: string,
): Promise<string> {
  try {
    const data = (await response.json()) as BackendErrorBody

    if (typeof data.error === 'string' && data.error.trim()) {
      return data.error
    }

    if (typeof data.message === 'string' && data.message.trim()) {
      return data.message
    }
  } catch {
    // Response body was empty or not JSON.
  }

  return fallbackMessage
}

type RequestOptions = {
  method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'
  token?: string
  body?: unknown
}

export async function apiRequest<T>(
  path: string,
  options: RequestOptions = {},
): Promise<T> {
  const headers: Record<string, string> = {}

  if (options.body !== undefined) {
    headers['Content-Type'] = 'application/json'
  }

  if (options.token) {
    headers.Authorization = `Bearer ${options.token}`
  }

  let response: Response

  try {
    response = await fetch(`${API_BASE_URL}${path}`, {
      method: options.method ?? 'GET',
      headers,
      body:
        options.body !== undefined ? JSON.stringify(options.body) : undefined,
    })
  } catch {
    throw new ApiError(
      'Unable to reach the server. Check that the API is running.',
      0,
    )
  }

  if (!response.ok) {
    const message = await readErrorMessage(
      response,
      `Request failed with status ${response.status}.`,
    )
    throw new ApiError(message, response.status)
  }

  if (response.status === 204) {
    return undefined as T
  }

  return (await response.json()) as T
}
