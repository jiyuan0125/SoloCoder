import { randomUUID } from 'crypto';

export function generateId(): string {
  return randomUUID();
}

export function parseNumeric(value: string | number): number | string {
  if (typeof value === 'number') return value;
  if (value === '待核实') return value;
  const num = parseFloat(value);
  return isNaN(num) ? value : num;
}

export function getMaxValue(a: string | number, b: string | number): string | number {
  const aNum = parseNumeric(a);
  const bNum = parseNumeric(b);
  
  if (aNum === '待核实' && bNum === '待核实') return '待核实';
  if (aNum === '待核实') return bNum;
  if (bNum === '待核实') return aNum;
  
  return Math.max(aNum as number, bNum as number);
}

export function isSameTimeRange(time1: string, time2: string, hoursRange: number = 24): boolean {
  const t1 = new Date(time1).getTime();
  const t2 = new Date(time2).getTime();
  const diffHours = Math.abs(t1 - t2) / (1000 * 60 * 60);
  return diffHours <= hoursRange;
}

export function formatDate(date: Date): string {
  return date.toISOString();
}

export function addHours(date: Date, hours: number): Date {
  const result = new Date(date);
  result.setHours(result.getHours() + hours);
  return result;
}

export function parseEconomicLoss(value: string | number): number {
  if (typeof value === 'number') return Math.round(value * 100);
  if (value === '待核实') return 0;
  const num = parseFloat(value);
  return isNaN(num) ? 0 : Math.round(num * 100);
}

export function toYuan(fen: number): number {
  return fen / 100;
}

export function isMajorDisaster(deathMissing: string | number, economicLoss: string | number): boolean {
  const deathNum = typeof deathMissing === 'number' ? deathMissing : (deathMissing === '待核实' ? 0 : parseFloat(deathMissing));
  const lossNum = typeof economicLoss === 'number' ? economicLoss : (economicLoss === '待核实' ? 0 : parseFloat(economicLoss));
  
  return deathNum > 0 || lossNum > 1000;
}

export function getVerificationDeadline(isMajor: boolean): string {
  const now = new Date();
  const hours = isMajor ? 24 : 72;
  return addHours(now, hours).toISOString();
}
