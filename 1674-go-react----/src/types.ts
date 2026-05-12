export enum Currency {
  CNY = 'CNY',
  USD = 'USD',
  EUR = 'EUR',
  JPY = 'JPY'
}

export const SUPPORTED_CURRENCIES: Currency[] = [
  Currency.CNY, Currency.USD, Currency.EUR, Currency.JPY
];

export enum TransactionStatus {
  PENDING = 'PENDING',
  COMPLETED = 'COMPLETED',
  PENDING_APPROVAL = 'PENDING_APPROVAL',
  CANCELLED = 'CANCELLED',
  AWAITING_RATE = 'AWAITING_RATE'
}

export interface Account {
  id: string;
  currency: Currency;
  balance: number;
  created_at: number;
}

export interface ExchangeRate {
  currency: Currency;
  rate: number;
  updated_at: number;
}

export interface Transaction {
  id: string;
  source_account_id: string;
  target_account_id: string;
  source_currency: Currency;
  target_currency: Currency;
  source_amount: number;
  target_amount?: number;
  exchange_rate_info?: string;
  status: TransactionStatus;
  created_at: number;
  updated_at: number;
}

export interface ExchangeRateInfo {
  sourceRate: number;
  targetRate: number;
  timestamp: number;
}
