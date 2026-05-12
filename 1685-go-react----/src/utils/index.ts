import { ServiceLevel } from '../types';
import { v4 as uuidv4 } from 'uuid';

export function generateId(): string {
  return uuidv4();
}

export function now(): string {
  return new Date().toISOString();
}

export function addDays(date: Date, days: number): Date {
  const result = new Date(date);
  result.setDate(result.getDate() + days);
  return result;
}

export function addMonths(date: Date, months: number): Date {
  const result = new Date(date);
  result.setMonth(result.getMonth() + months);
  return result;
}

export function formatDate(date: Date): string {
  return date.toISOString().split('T')[0];
}

export function calculateServiceLevel(totalScore: number): ServiceLevel {
  if (totalScore >= 160) {
    return ServiceLevel.LEVEL_1;
  } else if (totalScore >= 120) {
    return ServiceLevel.LEVEL_2;
  } else {
    return ServiceLevel.LEVEL_3;
  }
}

export function getHigherLevel(level1: ServiceLevel, level2: ServiceLevel): ServiceLevel {
  const levelOrder: Record<ServiceLevel, number> = {
    [ServiceLevel.LEVEL_1]: 1,
    [ServiceLevel.LEVEL_2]: 2,
    [ServiceLevel.LEVEL_3]: 3,
  };
  return levelOrder[level1] > levelOrder[level2] ? level1 : level2;
}

export function getVisitFrequencyDays(level: ServiceLevel): number {
  switch (level) {
    case ServiceLevel.LEVEL_1:
      return 30;
    case ServiceLevel.LEVEL_2:
      return 15;
    case ServiceLevel.LEVEL_3:
      return 7;
    default:
      return 30;
  }
}
