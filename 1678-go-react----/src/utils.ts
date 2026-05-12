import { v4 as uuidv4 } from 'uuid';
import { ProjectStatus } from './types';

const CREDIT_CODE_REGEX = /^[0-9A-HJ-NPQRTUWXY]{2}\d{6}[0-9A-HJ-NPQRTUWXY]{10}$/;

export function validateCreditCode(code: string): boolean {
  return CREDIT_CODE_REGEX.test(code);
}

export function validateBudget(amount: number): boolean {
  return amount > 0;
}

export function validateScore(score: number): boolean {
  return score >= 1 && score <= 100;
}

export function generateId(): string {
  return uuidv4();
}

export function getCurrentTime(): string {
  return new Date().toISOString();
}

export function addDays(date: string, days: number): string {
  const d = new Date(date);
  d.setDate(d.getDate() + days);
  return d.toISOString();
}

const STATUS_TRANSITIONS: Record<ProjectStatus, ProjectStatus[]> = {
  '申请中': ['初审'],
  '初审': ['评审', '不通过'],
  '评审': ['批准', '整改', '不通过'],
  '整改': ['评审'],
  '批准': [],
  '不通过': []
};

export function isValidStatusTransition(from: ProjectStatus, to: ProjectStatus): boolean {
  return STATUS_TRANSITIONS[from]?.includes(to) ?? false;
}

export function calculateAverageScore(scores: number[]): number {
  if (scores.length === 0) return 0;
  if (scores.length < 3) {
    return scores.reduce((sum, s) => sum + s, 0) / scores.length;
  }
  
  const sorted = [...scores].sort((a, b) => a - b);
  const trimmed = sorted.slice(1, -1);
  return trimmed.reduce((sum, s) => sum + s, 0) / trimmed.length;
}

export function getReviewDecision(averageScore: number): '批准' | '整改' | '不通过' {
  if (averageScore < 60) return '不通过';
  if (averageScore >= 60 && averageScore < 80) return '整改';
  return '批准';
}

export function getRandomItems<T>(arr: T[], min: number, max: number): T[] {
  const count = Math.floor(Math.random() * (max - min + 1)) + min;
  const shuffled = [...arr].sort(() => Math.random() - 0.5);
  return shuffled.slice(0, Math.min(count, arr.length));
}
