export function calculateDaysBetween(startDate: string, endDate: string): number {
  const start = new Date(startDate);
  const end = new Date(endDate);
  const diffTime = end.getTime() - start.getTime();
  const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
  return Math.max(0, diffDays);
}

export function calculateInterest(
  principal: number,
  annualRate: number,
  startDate: string,
  endDate: string
): number {
  const days = calculateDaysBetween(startDate, endDate);
  const interest = principal * (annualRate / 365) * days;
  return Math.round(interest);
}

export function calculateTotalInterest(
  principal: number,
  annualRate: number,
  penaltyRate: number,
  fundedDate: string,
  dueDate: string,
  paymentDate: string
): number {
  const funded = new Date(fundedDate);
  const due = new Date(dueDate);
  const payment = new Date(paymentDate);

  if (payment <= due) {
    return calculateInterest(principal, annualRate, fundedDate, paymentDate);
  }

  const normalInterest = calculateInterest(principal, annualRate, fundedDate, dueDate);
  const penaltyInterest = calculateInterest(principal, penaltyRate, dueDate, paymentDate);
  return normalInterest + penaltyInterest;
}

export function formatDate(date: Date = new Date()): string {
  return date.toISOString().split('T')[0];
}

export function isDateValid(dateString: string): boolean {
  const date = new Date(dateString);
  return !isNaN(date.getTime());
}

export function isRateValid(rate: number, min: number = 0, max: number = 1): boolean {
  return typeof rate === 'number' && rate > min && rate < max;
}
