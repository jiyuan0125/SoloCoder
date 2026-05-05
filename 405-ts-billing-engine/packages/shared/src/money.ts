export function multiplyByPercent(amount: number, percent: number): number {
  return Math.floor((amount * percent) / 100);
}

export function applyDiscount(amount: number, discountPercent: number): number {
  return amount - multiplyByPercent(amount, discountPercent);
}

export function calculateDailyRate(monthlyFee: number, daysInMonth: number): number {
  return Math.floor(monthlyFee / daysInMonth);
}

export function calculateProratedFee(
  dailyRate: number,
  days: number
): number {
  return dailyRate * days;
}

export function calculateOverageFee(
  overageUnits: number,
  feePerUnit: number
): number {
  return overageUnits * feePerUnit;
}

export function calculateYearlyDiscount(yearlyFee: number, discountPercent: number): number {
  return multiplyByPercent(yearlyFee, discountPercent);
}

export function calculateRefundAmount(
  remainingMonths: number,
  monthlyFee: number,
  refundFeePercent: number
): { totalAmount: number; refundAmount: number; feeAmount: number } {
  const totalAmount = remainingMonths * monthlyFee;
  const feeAmount = multiplyByPercent(totalAmount, refundFeePercent);
  const refundAmount = totalAmount - feeAmount;
  return { totalAmount, refundAmount, feeAmount };
}

export function formatMoney(amount: number): string {
  const yuan = Math.floor(amount / 100);
  const fen = amount % 100;
  return `${yuan}.${fen.toString().padStart(2, '0')}`;
}

export function parseMoney(str: string): number {
  const trimmed = str.trim();
  if (!trimmed.includes('.')) {
    return parseInt(trimmed, 10) * 100;
  }
  const [yuanStr, fenStr] = trimmed.split('.', 2);
  const yuan = yuanStr ? parseInt(yuanStr, 10) : 0;
  const fen = fenStr ? parseInt(fenStr.padEnd(2, '0').slice(0, 2), 10) : 0;
  return yuan * 100 + fen;
}
