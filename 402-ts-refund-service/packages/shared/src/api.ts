import { Refund, RefundStatus, Config, Order } from './types';
import { ServiceError } from './error-codes';

export interface ApiResponse<T = unknown> {
  success: boolean;
  data?: T;
  error?: ServiceError;
}

export interface CreateRefundRequest {
  orderId: string;
  userId: string;
  refundReason: string;
  productIds: string[];
}

export interface SubmitLogisticsRequest {
  refundId: string;
  logisticsNumber: string;
}

export interface WarehouseConfirmRequest {
  refundId: string;
  received: boolean;
}

export interface QueryRefundsRequest {
  page?: number;
  pageSize?: number;
  status?: RefundStatus;
  startTime?: string;
  endTime?: string;
  orderId?: string;
}

export interface QueryRefundsResponse {
  refunds: Refund[];
  total: number;
  page: number;
  pageSize: number;
}

export interface UpdateConfigRequest {
  refundPeriodDays?: number;
  virtualProductRefundThreshold?: number;
}

export interface MockOrderRequest {
  order: Omit<Order, 'orderTime'> & { orderTime?: string };
}
