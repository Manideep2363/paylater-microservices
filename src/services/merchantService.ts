import { apiRequest } from './http'

export type MerchantOption = {
  merchant_id: number
  name: string
}

/** Matches merchant-service publicMerchantResponse (GET /merchant/profile). */
export type MerchantProfile = {
  merchant_id: number
  name: string
  email: string
  phone: string
  commission_percentage: string
}

/**
 * Matches ledger-service transactionResponse (GET /merchant/transactions).
 * amount and commission fields are DECIMAL values serialized as strings.
 */
export type MerchantTransaction = {
  transaction_id: number
  user_id: number
  merchant_id: number
  amount: string
  commission_percentage: string
  commission_amount: string
  created_at: string
}

export function getMerchants(token: string): Promise<MerchantOption[]> {
  return apiRequest<MerchantOption[]>('/merchants', { token })
}

export function getMerchantProfile(token: string): Promise<MerchantProfile> {
  return apiRequest<MerchantProfile>('/merchant/profile', { token })
}

export function getMerchantTransactions(
  token: string,
): Promise<MerchantTransaction[]> {
  return apiRequest<MerchantTransaction[]>('/merchant/transactions', {
    token,
  })
}
