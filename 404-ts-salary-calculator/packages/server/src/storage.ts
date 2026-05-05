import { SalaryRecord, SalaryComparisonResult, MonthComparison, YoYComparison } from '@salary/shared';
import { randomUUID } from 'crypto';
import { promises as fs } from 'fs';
import path from 'path';

const DATA_DIR = path.resolve(process.cwd(), 'data');
const DATA_FILE = path.join(DATA_DIR, 'salary-records.json');

let salaryRecords: SalaryRecord[] = [];

function deepClone<T>(obj: T): T {
  return JSON.parse(JSON.stringify(obj)) as T;
}

export async function loadData(): Promise<void> {
  try {
    await fs.access(DATA_DIR);
  } catch {
    await fs.mkdir(DATA_DIR, { recursive: true });
  }
  
  try {
    const data = await fs.readFile(DATA_FILE, 'utf-8');
    salaryRecords = JSON.parse(data) as SalaryRecord[];
  } catch {
    salaryRecords = [];
  }
}

export async function saveData(): Promise<void> {
  try {
    await fs.access(DATA_DIR);
  } catch {
    await fs.mkdir(DATA_DIR, { recursive: true });
  }
  
  await fs.writeFile(DATA_FILE, JSON.stringify(salaryRecords, null, 2), 'utf-8');
}

export function saveRecord(result: Omit<SalaryRecord, 'id' | 'createdAt'>): SalaryRecord {
  const record: SalaryRecord = {
    ...deepClone(result),
    id: randomUUID(),
    createdAt: Date.now(),
  };
  
  salaryRecords = [...salaryRecords, record];
  
  saveData().catch((err) => {
    console.error('Failed to save data:', err);
  });
  
  return deepClone(record);
}

export function queryRecords(employeeId?: string, month?: string): SalaryRecord[] {
  let results = [...salaryRecords];
  
  if (employeeId) {
    results = results.filter((r) => r.employeeId === employeeId);
  }
  
  if (month) {
    results = results.filter((r) => r.month === month);
  }
  
  return results.map((r) => deepClone(r));
}

function parseMonth(month: string): { year: number; month: number } | null {
  const match = month.match(/^(\d{4})-(\d{2})$/);
  if (!match) return null;
  return {
    year: parseInt(match[1], 10),
    month: parseInt(match[2], 10),
  };
}

function getPreviousMonth(month: string): string | null {
  const parsed = parseMonth(month);
  if (!parsed) return null;
  
  let { year, month: m } = parsed;
  m -= 1;
  
  if (m < 1) {
    m = 12;
    year -= 1;
  }
  
  return `${year.toString().padStart(4, '0')}-${m.toString().padStart(2, '0')}`;
}

function getLastYearSameMonth(month: string): string | null {
  const parsed = parseMonth(month);
  if (!parsed) return null;
  
  return `${(parsed.year - 1).toString().padStart(4, '0')}-${parsed.month.toString().padStart(2, '0')}`;
}

function findRecordForMonth(records: SalaryRecord[], employeeId: string, month: string): SalaryRecord | null {
  return records.find((r) => r.employeeId === employeeId && r.month === month) || null;
}

function calculatePercentage(current: number, previous: number | null): number | null {
  if (previous === null || previous === 0) return null;
  return roundToTwoDecimals(((current - previous) / previous) * 100);
}

function roundToTwoDecimals(value: number): number {
  return Math.round(value * 100) / 100;
}

export function getComparison(employeeId: string, month: string): SalaryComparisonResult | null {
  const currentRecord = findRecordForMonth(salaryRecords, employeeId, month);
  if (!currentRecord) return null;
  
  const previousMonth = getPreviousMonth(month);
  const lastYearSameMonth = getLastYearSameMonth(month);
  
  const previousRecord = previousMonth
    ? findRecordForMonth(salaryRecords, employeeId, previousMonth)
    : null;
  const lastYearRecord = lastYearSameMonth
    ? findRecordForMonth(salaryRecords, employeeId, lastYearSameMonth)
    : null;
  
  const currentNet = currentRecord.netSalary;
  const previousNet = previousRecord?.netSalary ?? null;
  const lastYearNet = lastYearRecord?.netSalary ?? null;
  
  const monthOverMonth: MonthComparison = {
    current: currentNet,
    previous: previousNet,
    difference: previousNet !== null ? currentNet - previousNet : null,
    percentage: calculatePercentage(currentNet, previousNet),
  };
  
  const yearOverYear: YoYComparison = {
    current: currentNet,
    lastYearSameMonth: lastYearNet,
    difference: lastYearNet !== null ? currentNet - lastYearNet : null,
    percentage: calculatePercentage(currentNet, lastYearNet),
  };
  
  return {
    employeeId: currentRecord.employeeId,
    employeeName: currentRecord.employeeName,
    month,
    monthOverMonth,
    yearOverYear,
  };
}
