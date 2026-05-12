import { v4 as uuidv4 } from 'uuid';

export function generateId(): string {
  return uuidv4();
}

export function getNow(): string {
  return new Date().toISOString();
}

export function calculateAge(idCard: string): number {
  if (idCard.length !== 18) {
    throw new Error('身份证号格式不正确');
  }
  const birthStr = idCard.substring(6, 14);
  const year = parseInt(birthStr.substring(0, 4));
  const month = parseInt(birthStr.substring(4, 6)) - 1;
  const day = parseInt(birthStr.substring(6, 8));
  const birthDate = new Date(year, month, day);
  const today = new Date();
  let age = today.getFullYear() - birthDate.getFullYear();
  const monthDiff = today.getMonth() - birthDate.getMonth();
  if (monthDiff < 0 || (monthDiff === 0 && today.getDate() < birthDate.getDate())) {
    age--;
  }
  return age;
}

export function addDays(dateStr: string, days: number): string {
  const date = new Date(dateStr);
  date.setDate(date.getDate() + days);
  return date.toISOString();
}

export function isWithinDays(dateStr: string, days: number): boolean {
  const date = new Date(dateStr);
  const now = new Date();
  const diffMs = date.getTime() - now.getTime();
  const diffDays = Math.ceil(diffMs / (1000 * 60 * 60 * 24));
  return diffDays <= days && diffDays >= 0;
}

export function getStarLevel(totalHours: number): number {
  if (totalHours >= 1000) return 5;
  if (totalHours >= 500) return 4;
  if (totalHours >= 200) return 3;
  if (totalHours >= 100) return 2;
  if (totalHours >= 50) return 1;
  return 0;
}
