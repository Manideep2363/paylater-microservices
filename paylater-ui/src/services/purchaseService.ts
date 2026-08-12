import { apiRequest } from './http'

export type CreatePurchaseRequest = {
  merchant_id: number
  amount: number
}

export type CreatePurchaseResponse = {
  message: string
}

export function createPurchase(
  merchantId: number,
  amount: number,
  token: string,
): Promise<CreatePurchaseResponse> {
  const body: CreatePurchaseRequest = {
    merchant_id: merchantId,
    amount,
  }

  return apiRequest<CreatePurchaseResponse>('/purchases', {
    method: 'POST',
    token,
    body,
  })
}
