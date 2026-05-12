import { randomUUID } from 'crypto';

export function generateId(): string {
  return randomUUID();
}

export function getCurrentTime(): number {
  return Date.now();
}

export function isWithinHours(timestamp: number, hours: number): boolean {
  const diff = timestamp - getCurrentTime();
  return diff > 0 && diff <= hours * 60 * 60 * 1000;
}

export function isPastHours(timestamp: number, hours: number): boolean {
  const diff = getCurrentTime() - timestamp;
  return diff >= hours * 60 * 60 * 1000;
}

export function getYearMonth(ts: number): { year: number; month: number } {
  const date = new Date(ts);
  return {
    year: date.getFullYear(),
    month: date.getMonth() + 1,
  };
}
