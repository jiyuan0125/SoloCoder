import { InvestigationResult, IsolationType, TestResult } from './types';
import { v4 as uuidv4 } from 'uuid';

export const VALID_INVESTIGATION_RESULTS: InvestigationResult[] = ['normal', 'need_isolation', 'need_hospital'];
export const VALID_ISOLATION_TYPES: IsolationType[] = ['home', 'centralized'];
export const VALID_TEST_RESULTS: TestResult[] = ['positive', 'negative', 'pending'];
export const ISOLATION_TEST_DAYS = [1, 7, 14];

export function generateId(): string {
  return uuidv4();
}

export function isValidInvestigationResult(result: string): result is InvestigationResult {
  return VALID_INVESTIGATION_RESULTS.includes(result as InvestigationResult);
}

export function isValidIsolationType(type: string): type is IsolationType {
  return VALID_ISOLATION_TYPES.includes(type as IsolationType);
}

export function isValidTestResult(result: string): result is TestResult {
  return VALID_TEST_RESULTS.includes(result as TestResult);
}

export function isValidIsolationTestDay(day: number, isKeyPerson: boolean): boolean {
  if (isKeyPerson) {
    return day >= 1;
  }
  return ISOLATION_TEST_DAYS.includes(day);
}

export function hasHighRiskTravel(travelHistory: string): boolean {
  if (!travelHistory || travelHistory.trim() === '') {
    return false;
  }
  const keywords = ['高风险', '中风险', '中高风险', 'high-risk', 'medium-risk'];
  return keywords.some(keyword => travelHistory.includes(keyword));
}

export function calculateEndDate(startDate: string, days: number = 14): string {
  const start = new Date(startDate);
  start.setDate(start.getDate() + days - 1);
  return start.toISOString().split('T')[0];
}

export function calculateIsolationDay(startDate: string, checkDate: string): number {
  const start = new Date(startDate);
  const check = new Date(checkDate);
  const diffTime = check.getTime() - start.getTime();
  return Math.floor(diffTime / (1000 * 60 * 60 * 24)) + 1;
}

export function getDaysDiff(date1: string, date2: string): number {
  const d1 = new Date(date1);
  const d2 = new Date(date2);
  const diffTime = d2.getTime() - d1.getTime();
  return Math.floor(diffTime / (1000 * 60 * 60 * 24));
}

export function hasRespiratorySymptoms(symptoms: string): boolean {
  if (!symptoms) return false;
  const keywords = ['咳嗽', '乏力', '发热', '呼吸困难', '胸闷', 'cough', 'fever', 'shortness of breath'];
  return keywords.some(keyword => symptoms.toLowerCase().includes(keyword.toLowerCase()));
}

export function isHighTemperature(temp: number): boolean {
  return temp > 37.3;
}

export function today(): string {
  return new Date().toISOString().split('T')[0];
}

export function yesterday(): string {
  const d = new Date();
  d.setDate(d.getDate() - 1);
  return d.toISOString().split('T')[0];
}
