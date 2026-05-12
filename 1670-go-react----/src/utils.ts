import { LEVEL_FEES, SellerLevel } from './types';

export function validateAmount(amount: number): boolean {
  return Number.isInteger(amount) && amount >= 0;
}

export function validateSellerLevel(level: string): level is SellerLevel {
  return level in LEVEL_FEES;
}

export function calculateFee(netAmount: number, level: SellerLevel): number {
  const rate = LEVEL_FEES[level];
  return Math.floor(netAmount * rate);
}

export function generateId(): string {
  return Math.random().toString(36).substring(2, 15);
}
