import { LegalBasis } from '../types';
import { createHash } from 'crypto';

export const validLegalBases: LegalBasis[] = ['consent', 'contract', 'legal_obligation', 'legitimate_interest'];

export function isValidLegalBasis(value: string): value is LegalBasis {
  return validLegalBases.includes(value as LegalBasis);
}

export function addDays(date: Date, days: number): Date {
  const result = new Date(date);
  result.setDate(result.getDate() + days);
  return result;
}

export function isOverdue(deadline: Date): boolean {
  return new Date() > deadline;
}

export function anonymizeValue(value: string): string {
  const hash = createHash('sha256');
  hash.update(value + Date.now().toString());
  return `anon_${hash.digest('hex').substring(0, 12)}`;
}

export function daysBetween(date1: Date, date2: Date): number {
  const oneDay = 24 * 60 * 60 * 1000;
  const diff = Math.abs(date1.getTime() - date2.getTime());
  return Math.floor(diff / oneDay);
}
