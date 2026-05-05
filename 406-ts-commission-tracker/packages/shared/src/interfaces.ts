import {
  SalespersonId,
  OrderId,
  SettlementId,
  AmountInCents,
  Year,
  Month,
  Quarter,
  Salesperson,
  Order,
  Settlement,
  Ranking,
  RankingFilter,
  CommissionCalculation,
} from './types';

export interface CreateSalespersonRequest {
  name: string;
  joinDate: string;
}

export interface CreateSalespersonResponse {
  success: boolean;
  data?: Salesperson;
  error?: ApiError;
}

export interface GetSalespersonRequest {
  id: SalespersonId;
}

export interface GetSalespersonResponse {
  success: boolean;
  data?: Salesperson;
  error?: ApiError;
}

export interface ListSalespeopleResponse {
  success: boolean;
  data?: Salesperson[];
  error?: ApiError;
}

export interface ResignSalespersonRequest {
  id: SalespersonId;
}

export interface ResignSalespersonResponse {
  success: boolean;
  data?: {
    salesperson: Salesperson;
    settlement: Settlement;
  };
  error?: ApiError;
}

export interface CreateOrderRequest {
  amount: AmountInCents;
  date: string;
  salesContributions: {
    salespersonId: SalespersonId;
    isPrimary: boolean;
  }[];
}

export interface CreateOrderResponse {
  success: boolean;
  data?: Order;
  error?: ApiError;
}

export interface GetOrderRequest {
  id: OrderId;
}

export interface GetOrderResponse {
  success: boolean;
  data?: Order;
  error?: ApiError;
}

export interface RefundOrderRequest {
  id: OrderId;
}

export interface RefundOrderResponse {
  success: boolean;
  data?: Order;
  error?: ApiError;
}

export interface ListOrdersRequest {
  salespersonId?: SalespersonId;
  fromDate?: string;
  toDate?: string;
}

export interface ListOrdersResponse {
  success: boolean;
  data?: Order[];
  error?: ApiError;
}

export interface GetSalespersonCommissionRequest {
  salespersonId: SalespersonId;
  month: Month;
  year: Year;
}

export interface GetSalespersonCommissionResponse {
  success: boolean;
  data?: CommissionCalculation;
  error?: ApiError;
}

export interface GetSalespersonBalanceRequest {
  salespersonId: SalespersonId;
}

export interface GetSalespersonBalanceResponse {
  success: boolean;
  data?: {
    salespersonId: SalespersonId;
    balance: AmountInCents;
    thisMonthSettled: AmountInCents;
  };
  error?: ApiError;
}

export interface GetRankingsRequest extends RankingFilter {}

export interface GetRankingsResponse {
  success: boolean;
  data?: Ranking[];
  error?: ApiError;
}

export interface GetSettlementsRequest {
  salespersonId?: SalespersonId;
  year?: Year;
  month?: Month;
}

export interface GetSettlementsResponse {
  success: boolean;
  data?: Settlement[];
  error?: ApiError;
}

export interface GetSettlementRequest {
  id: SettlementId;
}

export interface GetSettlementResponse {
  success: boolean;
  data?: Settlement;
  error?: ApiError;
}

export interface TriggerMonthlySettlementRequest {
  month: Month;
  year: Year;
}

export interface TriggerMonthlySettlementResponse {
  success: boolean;
  data?: {
    settlements: Settlement[];
    totalSettled: AmountInCents;
  };
  error?: ApiError;
}

export interface ApiError {
  code: string;
  message: string;
  details?: Record<string, unknown>;
}

export interface DataStore {
  getSalesperson(id: SalespersonId): Salesperson | undefined;
  getSalespeople(): Salesperson[];
  saveSalesperson(salesperson: Salesperson): void;

  getOrder(id: OrderId): Order | undefined;
  getOrders(): Order[];
  saveOrder(order: Order): void;

  getSettlement(id: SettlementId): Settlement | undefined;
  getSettlements(): Settlement[];
  saveSettlement(settlement: Settlement): void;

  getSettlementsForSalesperson(salespersonId: SalespersonId): Settlement[];
  getSettlementsForMonth(month: Month, year: Year): Settlement[];

  getOrdersForSalesperson(salespersonId: SalespersonId): Order[];
  getOrdersForMonth(month: Month, year: Year): Order[];

  lockOrder(id: OrderId): void;
  isOrderLocked(id: OrderId): boolean;
}
