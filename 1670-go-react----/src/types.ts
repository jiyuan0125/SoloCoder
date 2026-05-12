export type SellerLevel = 'ordinary' | 'silver' | 'gold';

export const LEVEL_FEES: Record<SellerLevel, number> = {
  ordinary: 0.02,
  silver: 0.015,
  gold: 0.01,
};

export const SETTLEMENT_STATUS = {
  PENDING: 'pending',
  CONFIRMED: 'confirmed',
  PAID: 'paid',
} as const;

export type SettlementStatus = typeof SETTLEMENT_STATUS[keyof typeof SETTLEMENT_STATUS];

export interface Seller {
  id: string;
  name: string;
  level: SellerLevel;
  created_at: number;
}

export interface Transaction {
  id: string;
  seller_id: string;
  type: 'receivable' | 'refund';
  amount: number;
  cycle: string;
  created_at: number;
}

export interface Settlement {
  id: string;
  seller_id: string;
  cycle: string;
  total_receivable: number;
  total_refund: number;
  net_amount: number;
  fee: number;
  payable_amount: number;
  status: SettlementStatus;
  created_at: number;
  confirmed_at: number | null;
  paid_at: number | null;
}

export interface FeeRecord {
  id: string;
  settlement_id: string;
  seller_id: string;
  cycle: string;
  amount: number;
  created_at: number;
}
