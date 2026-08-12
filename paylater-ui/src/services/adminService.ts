import { apiRequest } from './http'
import type { User } from './userService'
import type { MerchantProfile } from './merchantService'

export type AdminUser = User

export type CreateAdminUserRequest = {
  name: string
  email: string
  password: string
}

export type AdminMerchant = MerchantProfile

export type CreateAdminMerchantRequest = {
  name: string
  email: string
  phone: string
  password: string
  commission: number
}

export type UpdateMerchantCommissionRequest = {
  commission: number
}

export type MessageResponse = {
  message: string
}

export type AdminPurchase = {
  transaction_id: number
  user_id: number
  merchant_id: number
  amount: string
  commission_percentage: string
  commission_amount: string
  created_at: string
}

export type AdminPayment = {
  payment_id: number
  user_id: number
  amount: string
  paid_at: string
}

export type OutstandingBalanceReport = {
  total_outstanding_balance: string
}

export type UserDueReportRow = {
  user_id: number
  name: string
  current_due: string
}

export type UserAtCreditLimitRow = {
  user_id: number
  name: string
  credit_limit: string
  current_due: string
}

export type MerchantCommissionReportRow = {
  merchant_id: number
  total_commission: string
}

export function getAdminUsers(token: string): Promise<AdminUser[]> {
  return apiRequest<AdminUser[]>('/admin/users', { token })
}

export function createAdminUser(
  payload: CreateAdminUserRequest,
  token: string,
): Promise<AdminUser> {
  return apiRequest<AdminUser>('/admin/users', {
    method: 'POST',
    token,
    body: payload,
  })
}

export function getAdminMerchants(token: string): Promise<AdminMerchant[]> {
  return apiRequest<AdminMerchant[]>('/admin/merchants', { token })
}

export function getAdminMerchantById(
  id: number,
  token: string,
): Promise<AdminMerchant> {
  return apiRequest<AdminMerchant>(`/admin/merchants/${id}`, { token })
}

export function createAdminMerchant(
  payload: CreateAdminMerchantRequest,
  token: string,
): Promise<AdminMerchant> {
  return apiRequest<AdminMerchant>('/admin/merchants', {
    method: 'POST',
    token,
    body: payload,
  })
}

export function updateMerchantCommission(
  id: number,
  commissionPercentage: number,
  token: string,
): Promise<MessageResponse> {
  const body: UpdateMerchantCommissionRequest = {
    commission: commissionPercentage,
  }

  return apiRequest<MessageResponse>(`/admin/merchants/${id}/commission`, {
    method: 'PUT',
    token,
    body,
  })
}

export function getAdminPurchases(token: string): Promise<AdminPurchase[]> {
  return apiRequest<AdminPurchase[]>('/admin/purchases', { token })
}

export function getAdminPurchaseById(
  id: number,
  token: string,
): Promise<AdminPurchase> {
  return apiRequest<AdminPurchase>(`/admin/purchases/${id}`, { token })
}

export function getAdminPaymentById(
  id: number,
  token: string,
): Promise<AdminPayment> {
  return apiRequest<AdminPayment>(`/admin/payments/${id}`, { token })
}

export function getAdminUserPayments(
  userId: number,
  token: string,
): Promise<AdminPayment[]> {
  return apiRequest<AdminPayment[]>(`/admin/users/${userId}/payments`, {
    token,
  })
}

export function getOutstandingBalanceReport(
  token: string,
): Promise<OutstandingBalanceReport> {
  return apiRequest<OutstandingBalanceReport>(
    '/admin/reports/outstanding-balance',
    { token },
  )
}

export function getUsersDueReport(token: string): Promise<UserDueReportRow[]> {
  return apiRequest<UserDueReportRow[]>('/admin/reports/users-due', { token })
}

export function getUsersAtCreditLimitReport(
  token: string,
): Promise<UserAtCreditLimitRow[]> {
  return apiRequest<UserAtCreditLimitRow[]>(
    '/admin/reports/users-at-credit-limit',
    { token },
  )
}

export function getMerchantCommissionReport(
  token: string,
): Promise<MerchantCommissionReportRow[]> {
  return apiRequest<MerchantCommissionReportRow[]>(
    '/admin/reports/merchant-commissions',
    { token },
  )
}
