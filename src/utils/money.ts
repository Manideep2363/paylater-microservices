export function parseAmount(value: string | number): number {
  const parsed = typeof value === 'number' ? value : Number(value)
  return Number.isFinite(parsed) ? parsed : 0
}

export function formatCurrency(amount: string | number): string {
  const value = parseAmount(amount)
  return new Intl.NumberFormat('en-IN', {
    style: 'currency',
    currency: 'INR',
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(value)
}

export function formatDateTime(value: string): string {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }

  return new Intl.DateTimeFormat('en-IN', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(date)
}

export function calcAvailableCredit(
  creditLimit: string,
  currentDue: string,
): number {
  return Math.max(0, parseAmount(creditLimit) - parseAmount(currentDue))
}

export function calcCreditUsagePercent(
  creditLimit: string,
  currentDue: string,
): number {
  const limit = parseAmount(creditLimit)
  if (limit <= 0) {
    return 0
  }

  const usage = (parseAmount(currentDue) / limit) * 100
  return Math.min(100, Math.max(0, usage))
}

export function sumAmounts(values: Array<string | number>): number {
  return values.reduce<number>((total, value) => total + parseAmount(value), 0)
}
