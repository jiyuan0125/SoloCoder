export function generateId(): string {
  const timestamp = Date.now().toString(36);
  const random = Math.random().toString(36).substring(2, 10);
  return `${timestamp}-${random}`;
}

export function getCurrentDate(): string {
  return new Date().toISOString();
}

export function addDays(date: string | Date, days: number): string {
  const d = new Date(date);
  d.setDate(d.getDate() + days);
  return d.toISOString();
}

export function addMonths(date: string | Date, months: number): string {
  const d = new Date(date);
  d.setMonth(d.getMonth() + months);
  return d.toISOString();
}

export function getDaysInMonth(date: string | Date): number {
  const d = new Date(date);
  return new Date(d.getFullYear(), d.getMonth() + 1, 0).getDate();
}

export function getDaysBetween(start: string | Date, end: string | Date): number {
  const startDate = new Date(start);
  const endDate = new Date(end);
  const diffTime = endDate.getTime() - startDate.getTime();
  return Math.ceil(diffTime / (1000 * 60 * 60 * 24));
}

export function getFirstDayOfMonth(date: string | Date): string {
  const d = new Date(date);
  return new Date(d.getFullYear(), d.getMonth(), 1).toISOString();
}

export function getLastDayOfMonth(date: string | Date): string {
  const d = new Date(date);
  return new Date(d.getFullYear(), d.getMonth() + 1, 0).toISOString();
}

export function isDateAfter(date1: string | Date, date2: string | Date): boolean {
  return new Date(date1).getTime() > new Date(date2).getTime();
}

export function isDateBefore(date1: string | Date, date2: string | Date): boolean {
  return new Date(date1).getTime() < new Date(date2).getTime();
}

export function isDateBetween(
  date: string | Date,
  start: string | Date,
  end: string | Date
): boolean {
  const d = new Date(date).getTime();
  return d >= new Date(start).getTime() && d <= new Date(end).getTime();
}

export function getMonthsBetween(start: string | Date, end: string | Date): number {
  const startDate = new Date(start);
  const endDate = new Date(end);
  let months = (endDate.getFullYear() - startDate.getFullYear()) * 12;
  months += endDate.getMonth() - startDate.getMonth();
  return months;
}
