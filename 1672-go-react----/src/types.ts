export enum AccountType {
  INCOME = 'income',
  EXPENSE = 'expense',
  MAIN_POOL = 'main_pool'
}

export enum TransferStatus {
  PENDING = 'pending',
  SUCCESS = 'success',
  FAILED = 'failed'
}

export interface Account {
  id: string;
  name: string;
  bank: string;
  balance: number;
  type: AccountType;
  currency: string;
  allowOverdraft: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface Transfer {
  id: string;
  fromAccountId: string;
  toAccountId: string;
  amount: number;
  reason: string;
  status: TransferStatus;
  operator: string;
  createdAt: string;
  completedAt?: string;
  failureReason?: string;
}

export interface InterestRecord {
  id: string;
  accountId: string;
  balanceSnapshot: number;
  dailyInterest: number;
  date: string;
  createdAt: string;
}

export interface InterestSettlement {
  id: string;
  accountId: string;
  year: number;
  month: number;
  totalInterest: number;
  settlementDate: string;
  createdAt: string;
}

export interface SystemConfig {
  dailyInterestRate: number;
  updatedAt: string;
}
