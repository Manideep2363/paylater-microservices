import { apiRequest } from './http'

export type CreatePaymentRequest = {
  amount: number
}

export type PaymentResponse = {
  message: string
}

export type Payment = {
  payment_id: number
  user_id: number
  amount: string
  paid_at: string
}

export function createPayment(
  amount: number,
  token: string,
): Promise<PaymentResponse> {
  const body: CreatePaymentRequest = { amount }

  return apiRequest<PaymentResponse>('/payments', {
    method: 'POST',
    token,
    body,
  })
}

export function getPayments(token: string): Promise<Payment[]> {
  return apiRequest<Payment[]>('/payments', { token })
}
