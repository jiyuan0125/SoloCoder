export function generateId(): string {
  const timestamp = Date.now().toString(36);
  const random = Math.random().toString(36).substring(2, 8);
  return `${timestamp}-${random}`;
}

export function isRefundPeriodExpired(orderTime: Date, refundPeriodDays: number): boolean {
  const now = new Date();
  const orderDate = new Date(orderTime);
  const diffMs = now.getTime() - orderDate.getTime();
  const diffDays = diffMs / (1000 * 60 * 60 * 24);
  return diffDays > refundPeriodDays;
}

export function calculateVirtualRefundAmount(
  paymentAmount: number,
  consumptionProgress: number,
  threshold: number
): number {
  if (consumptionProgress >= threshold) {
    const unconsumedProgress = 100 - consumptionProgress;
    return Math.floor((paymentAmount * unconsumedProgress) / 100);
  }
  return paymentAmount;
}
