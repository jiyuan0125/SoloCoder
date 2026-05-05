export type SalespersonId = string;
export type OrderId = string;
export type SettlementId = string;

export type AmountInCents = number;
export type Percentage = number;

export type Year = number;
export type Month = number;
export type Quarter = 1 | 2 | 3 | 4;

export interface Salesperson {
  id: SalespersonId;
  name: string;
  joinDate: string;
  status: SalespersonStatus;
  balance: AmountInCents;
}

export type SalespersonStatus = 'active' | 'resigned';

export interface Order {
  id: OrderId;
  amount: AmountInCents;
  date: string;
  status: OrderStatus;
  locked: boolean;
  salesContributions: SalesContribution[];
}

export type OrderStatus = 'active' | 'refunded';

export interface SalesContribution {
  salespersonId: SalespersonId;
  isPrimary: boolean;
}

export interface CommissionTier {
  from: AmountInCents;
  to: AmountInCents | null;
  rate: Percentage;
}

export interface CommissionCalculation {
  salespersonId: SalespersonId;
  month: Month;
  year: Year;
  totalSales: AmountInCents;
  commissionBeforeDiscount: AmountInCents;
  discountRate: number;
  finalCommission: AmountInCents;
  tierBreakdown: TierBreakdown[];
}

export interface TierBreakdown {
  tier: string;
  salesAmount: AmountInCents;
  commission: AmountInCents;
}

export interface Settlement {
  id: SettlementId;
  salespersonId: SalespersonId;
  month: Month;
  year: Year;
  status: SettlementStatus;
  settledAt: string;
  details: SettlementDetails;
  summary: SettlementSummary;
}

export type SettlementStatus = 'pending' | 'completed' | 'reversed';

export interface SettlementDetails {
  orders: SettledOrder[];
  refunds: SettledRefund[];
  balanceAdjustment: AmountInCents;
}

export interface SettledOrder {
  orderId: OrderId;
  amount: AmountInCents;
  commission: AmountInCents;
}

export interface SettledRefund {
  orderId: OrderId;
  amount: AmountInCents;
  commissionDeducted: AmountInCents;
}

export interface SettlementSummary {
  totalSales: AmountInCents;
  totalCommissionBeforeDiscount: AmountInCents;
  totalDiscount: AmountInCents;
  totalRefundDeduction: AmountInCents;
  balanceAdjustment: AmountInCents;
  finalSettlementAmount: AmountInCents;
}

export interface Ranking {
  salespersonId: SalespersonId;
  salespersonName: string;
  rank: number;
  totalSales: AmountInCents;
  totalCommission: AmountInCents;
}

export interface RankingFilter {
  dimension: 'monthly' | 'quarterly' | 'yearly';
  year: Year;
  month?: Month;
  quarter?: Quarter;
}

export interface BalanceInfo {
  salespersonId: SalespersonId;
  balance: AmountInCents;
  thisMonthSettled: AmountInCents;
}

export interface ResignResult {
  salesperson: Salesperson;
  settlement: Settlement;
}

export interface TriggerSettlementResult {
  settlements: Settlement[];
  totalSettled: AmountInCents;
}
