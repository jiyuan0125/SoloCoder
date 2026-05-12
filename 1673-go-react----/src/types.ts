export enum AccountType {
  MAIN = 'main',
  RECHARGE = 'recharge',
  ESCROW = 'escrow',
  WITHDRAW = 'withdraw',
  PLATFORM_INCOME = 'platform_income'
}

export enum AccountStatus {
  ACTIVE = 'active',
  FROZEN = 'frozen',
  CLOSED = 'closed'
}

export enum TransactionType {
  RECHARGE = 'recharge',
  CONSUME = 'consume',
  REFUND = 'refund',
  WITHDRAW = 'withdraw',
  TRANSFER = 'transfer',
  FEE = 'fee'
}

export enum EscrowTransactionStatus {
  PENDING = 'pending',
  PAID = 'paid',
  SHIPPED = 'shipped',
  COMPLETED = 'completed',
  REFUNDED = 'refunded'
}

export interface User {
  id: string;
  name: string;
  created_at: number;
}

export interface Account {
  id: string;
  user_id: string;
  type: AccountType;
  balance: number;
  overdraft_limit: number;
  status: AccountStatus;
  parent_id: string | null;
  created_at: number;
  updated_at: number;
}

export interface Transaction {
  id: string;
  account_id: string;
  type: TransactionType;
  amount: number;
  balance_before: number;
  balance_after: number;
  description: string;
  reference_id: string | null;
  created_at: number;
}

export interface EscrowTransaction {
  id: string;
  order_id: string;
  buyer_user_id: string;
  seller_user_id: string;
  buyer_escrow_account_id: string;
  seller_recharge_account_id: string;
  amount: number;
  platform_fee: number;
  status: EscrowTransactionStatus;
  created_at: number;
  updated_at: number;
}

export interface WithdrawRecord {
  id: string;
  user_id: string;
  account_id: string;
  bank_account: string;
  amount: number;
  status: string;
  created_at: number;
}
