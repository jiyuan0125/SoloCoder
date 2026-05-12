import { format, addYears, isWithinInterval, parse, startOfYear, endOfYear, subDays, differenceInDays } from 'date-fns';
import type { OrganizationType, EvaluationGrade, InspectionResult } from './types';

export const VALID_TYPES: OrganizationType[] = ['社会团体', '民办非企业', '基金会'];

export function isValidCreditCode(code: string): boolean {
  const regex = /^[0-9A-HJ-NPQRTUWXY]{2}\d{6}[0-9A-HJ-NPQRTUWXY]{10}$/;
  return regex.test(code);
}

export function normalizeName(name: string): string {
  return name.replace(/[\s\p{P}]/gu, '');
}

export function nowISO(): string {
  return new Date().toISOString();
}

export function generateCertificateNo(): string {
  const timestamp = Date.now().toString();
  const random = Math.floor(Math.random() * 10000).toString().padStart(4, '0');
  return `REG-${timestamp.slice(-8)}-${random}`;
}

export function getCertificateDates(): { issueDate: string; expiryDate: string } {
  const now = new Date();
  const issueDate = format(now, 'yyyy-MM-dd');
  const expiryDate = format(addYears(now, 5), 'yyyy-MM-dd');
  return { issueDate, expiryDate };
}

export function isInInspectionPeriod(date: Date, year: number): boolean {
  const start = parse(`${year}-03-01`, 'yyyy-MM-dd', startOfYear(date));
  const end = parse(`${year}-05-31`, 'yyyy-MM-dd', endOfYear(date));
  return isWithinInterval(date, { start, end });
}

export function isInspectionReminderDue(date: Date, year: number): boolean {
  const start = parse(`${year}-03-01`, 'yyyy-MM-dd', startOfYear(date));
  const reminderStart = subDays(start, 30);
  return isWithinInterval(date, { start: reminderStart, end: start });
}

export function isWithinReSubmitWindow(rejectTime: string): boolean {
  const rejectDate = new Date(rejectTime);
  return differenceInDays(new Date(), rejectDate) <= 30;
}

export function calculateEvaluationGrade(totalScore: number): EvaluationGrade {
  if (totalScore >= 90) return '5A';
  if (totalScore >= 80) return '4A';
  if (totalScore >= 70) return '3A';
  if (totalScore >= 60) return '2A';
  return '1A';
}

export function canAcceptGovernmentServices(grade: EvaluationGrade): boolean {
  return ['5A', '4A', '3A'].includes(grade);
}

export function shouldAutoCancel(
  results: InspectionResult[]
): boolean {
  const recent = results.slice(-2);
  return recent.length === 2 && recent.every(r => r === '不合格');
}
