export type PaymentStatus = 'PENDING' | 'PAID' | 'CLOSED' | 'REFUNDED';

export type RoutingStrategy = 'RATE' | 'SUCCESS_RATE' | 'SPECIFIED';

export interface Channel {
  id: string;
  name: string;
  feeTiers: FeeTier[];
  isDegraded: boolean;
  degradedAt: number | null;
  createdAt: number;
}

export interface FeeTier {
  maxAmount: number;
  rate: number;
}

export interface PaymentRecord {
  id: string;
  merchantOrderNo: string;
  amount: number;
  channelId: string;
  channelTransactionNo: string | null;
  status: PaymentStatus;
  refundedAmount: number;
  createdAt: number;
  paidAt: number | null;
  closedAt: number | null;
  refundedAt: number | null;
}

export interface HealthMetric {
  id: string;
  channelId: string;
  hour: number;
  totalRequests: number;
  successCount: number;
  totalResponseTime: number;
  createdAt: number;
}

export interface CreatePaymentRequest {
  merchantOrderNo: string;
  amount: number;
  strategy: RoutingStrategy;
  channelId?: string;
}

export interface RefundRequest {
  paymentId: string;
  refundAmount: number;
}
