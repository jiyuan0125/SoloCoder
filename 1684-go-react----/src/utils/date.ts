export function getCurrentDate(): string {
  return new Date().toISOString().split('T')[0];
}

export function getCurrentDateTime(): string {
  return new Date().toISOString();
}

export function parseDate(dateStr: string): Date {
  return new Date(dateStr);
}

export function formatDate(date: Date): string {
  return date.toISOString().split('T')[0];
}

export function daysBetween(startDate: string, endDate: string): number {
  const start = parseDate(startDate);
  const end = parseDate(endDate);
  const diffTime = Math.abs(end.getTime() - start.getTime());
  const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
  return diffDays;
}

export function daysBetweenInclusive(startDate: string, endDate: string): number {
  return daysBetween(startDate, endDate) + 1;
}

export function isDateInRange(date: string, start: string, end: string): boolean {
  const d = parseDate(date);
  const s = parseDate(start);
  const e = parseDate(end);
  return d >= s && d <= e;
}

export function isValidTimeSlot(slot: string): boolean {
  return slot === 'morning' || slot === 'afternoon';
}

export function formatAmount(amount: number): string {
  return `¥${(amount / 100).toFixed(2)}`;
}

export function monthsBetween(startDate: string, endDate: string): number {
  const start = parseDate(startDate);
  const end = parseDate(endDate);
  const yearDiff = end.getFullYear() - start.getFullYear();
  const monthDiff = end.getMonth() - start.getMonth();
  return yearDiff * 12 + monthDiff;
}

export function addMonths(dateStr: string, months: number): string {
  const date = parseDate(dateStr);
  date.setMonth(date.getMonth() + months);
  return formatDate(date);
}

export function isDateBefore(date1: string, date2: string): boolean {
  return parseDate(date1) < parseDate(date2);
}

export function isDateAfter(date1: string, date2: string): boolean {
  return parseDate(date1) > parseDate(date2);
}
