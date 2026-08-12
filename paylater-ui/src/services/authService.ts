const API_BASE_URL = 'http://localhost:8080'

export type LoginRequest = {
  email: string
  password: string
}

export type LoginResponse = {
  token: string
}

type BackendErrorBody = {
  error?: string
  message?: string
}

export class AuthApiError extends Error {
  status: number

  constructor(message: string, status: number) {
    super(message)
    this.name = 'AuthApiError'
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
    // Response body was empty or not JSON — use the fallback below.
  }

  return fallbackMessage
}

async function postLogin(
  path: string,
  email: string,
  password: string,
): Promise<LoginResponse> {
  const body: LoginRequest = { email, password }

  let response: Response

  try {
    response = await fetch(`${API_BASE_URL}${path}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(body),
    })
  } catch {
    throw new AuthApiError(
      'Unable to reach the server. Check that the API is running.',
      0,
    )
  }

  if (!response.ok) {
    const message = await readErrorMessage(
      response,
      `Login failed with status ${response.status}.`,
    )
    throw new AuthApiError(message, response.status)
  }

  const data = (await response.json()) as LoginResponse

  if (!data.token) {
    throw new AuthApiError('Login succeeded but no token was returned.', response.status)
  }

  return data
}

export function loginUser(
  email: string,
  password: string,
): Promise<LoginResponse> {
  return postLogin('/login', email, password)
}

export function loginMerchant(
  email: string,
  password: string,
): Promise<LoginResponse> {
  return postLogin('/merchant/login', email, password)
}

export function loginAdmin(
  email: string,
  password: string,
): Promise<LoginResponse> {
  return postLogin('/admin/login', email, password)
}

export type RegisterUserRequest = {
  name: string
  email: string
  password: string
}

export type RegisterUserResponse = {
  message: string
}

export async function registerUser(
  name: string,
  email: string,
  password: string,
): Promise<RegisterUserResponse> {
  const body: RegisterUserRequest = { name, email, password }

  let response: Response

  try {
    response = await fetch(`${API_BASE_URL}/register`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(body),
    })
  } catch {
    throw new AuthApiError(
      'Unable to reach the server. Check that the API is running.',
      0,
    )
  }

  if (!response.ok) {
    const message = await readErrorMessage(
      response,
      `Registration failed with status ${response.status}.`,
    )
    throw new AuthApiError(message, response.status)
  }

  // Backend returns HTTP 201 with {"message": "..."}.
  try {
    const data = (await response.json()) as RegisterUserResponse
    return {
      message: data.message || 'User registered successfully',
    }
  } catch {
    return { message: 'User registered successfully' }
  }
}

export type RegisterMerchantRequest = {
  name: string
  email: string
  phone: string
  password: string
  commission_percentage: number
}

export type RegisterMerchantResponse = {
  message: string
}

export async function registerMerchant(
  name: string,
  email: string,
  phone: string,
  password: string,
  commissionPercentage: number,
): Promise<RegisterMerchantResponse> {
  const body: RegisterMerchantRequest = {
    name,
    email,
    phone,
    password,
    commission_percentage: commissionPercentage,
  }

  let response: Response

  try {
    response = await fetch(`${API_BASE_URL}/merchant/register`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(body),
    })
  } catch {
    throw new AuthApiError(
      'Unable to reach the server. Check that the API is running.',
      0,
    )
  }

  if (!response.ok) {
    const message = await readErrorMessage(
      response,
      `Merchant registration failed with status ${response.status}.`,
    )
    throw new AuthApiError(message, response.status)
  }

  try {
    const data = (await response.json()) as RegisterMerchantResponse
    return {
      message: data.message || 'merchant registered successfully',
    }
  } catch {
    return { message: 'merchant registered successfully' }
  }
}
