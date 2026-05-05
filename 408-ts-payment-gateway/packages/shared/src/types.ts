export type Amount = number;

export enum PaymentChannel {
  WECHAT = "wechat",
  ALIPAY = "alipay",
}

export enum OrderStatus {
  PENDING = "pending",
  PROCESSING = "processing",
  SUCCESS = "success",
  FAILED = "failed",
  PENDING_CONFIRM = "pending_confirm",
  CLOSED = "closed",
  REFUNDED = "refunded",
  PARTIAL_REFUNDED = "partial_refunded",
}

export enum RefundStatus {
  PENDING = "pending",
  PROCESSING = "processing",
  SUCCESS = "success",
  FAILED = "failed",
}

export interface Order {
  orderNo: string;
  amount: Amount;
  channel: PaymentChannel;
  status: OrderStatus;
  transactionNo: string | null;
  paidAt: number | null;
  createdAt: number;
  updatedAt: number;
  closedAt: number | null;
  lastSyncAt: number | null;
  totalRefundedAmount: Amount;
  refundableAmount: Amount;
}

export interface Payment {
  paymentId: string;
  orderNo: string;
  channel: PaymentChannel;
  amount: Amount;
  status: OrderStatus;
  channelPaymentId: string | null;
  createdAt: number;
  paidAt: number | null;
}

export interface Refund {
  refundNo: string;
  orderNo: string;
  channel: PaymentChannel;
  amount: Amount;
  status: RefundStatus;
  channelRefundId: string | null;
  createdAt: number;
  completedAt: number | null;
  isFinalRefund: boolean;
}

export interface CallbackRecord {
  id: string;
  orderNo: string;
  channel: PaymentChannel;
  rawData: string;
  processedAt: number;
}

export interface ChannelCallbackData {
  orderNo: string;
  channel: PaymentChannel;
  status: "success" | "failed";
  transactionId: string;
  amount: Amount;
  timestamp: number;
  signature: string;
}

export interface WechatCallbackData extends ChannelCallbackData {
  channel: PaymentChannel.WECHAT;
  appId: string;
  mchId: string;
  nonceStr: string;
}

export interface AlipayCallbackData extends ChannelCallbackData {
  channel: PaymentChannel.ALIPAY;
  tradeNo: string;
  sellerId: string;
  notifyId: string;
}

export interface ChannelRefundCallbackData {
  orderNo: string;
  refundNo: string;
  channel: PaymentChannel;
  status: "success" | "failed";
  channelRefundId: string;
  amount: Amount;
  timestamp: number;
  signature: string;
}

export interface CreateOrderRequest {
  orderNo: string;
  amount: Amount;
}

export interface CreateOrderResponse {
  orderNo: string;
  amount: Amount;
  channel: PaymentChannel;
  status: OrderStatus;
  paymentUrl: string;
  createdAt: number;
}

export interface PayOrderRequest {
  orderNo: string;
}

export interface PayOrderResponse {
  orderNo: string;
  channel: PaymentChannel;
  amount: Amount;
  status: OrderStatus;
  paymentUrl: string;
  paymentId: string;
}

export interface OrderQueryRequest {
  orderNo: string;
}

export interface OrderQueryResponse {
  orderNo: string;
  amount: Amount;
  channel: PaymentChannel;
  status: OrderStatus;
  transactionNo: string | null;
  paidAt: number | null;
  createdAt: number;
  totalRefundedAmount: Amount;
  refundableAmount: Amount;
  refunds: Array<{
    refundNo: string;
    amount: Amount;
    status: RefundStatus;
    createdAt: number;
    completedAt: number | null;
  }>;
}

export interface CreateRefundRequest {
  orderNo: string;
  refundNo: string;
  amount: Amount;
  isFinalRefund?: boolean;
}

export interface CreateRefundResponse {
  refundNo: string;
  orderNo: string;
  channel: PaymentChannel;
  amount: Amount;
  status: RefundStatus;
  createdAt: number;
}

export interface RefundQueryRequest {
  orderNo: string;
  refundNo?: string;
}

export interface RefundQueryResponse {
  orderNo: string;
  totalAmount: Amount;
  totalRefundedAmount: Amount;
  refundableAmount: Amount;
  refunds: Array<{
    refundNo: string;
    amount: Amount;
    status: RefundStatus;
    channelRefundId: string | null;
    createdAt: number;
    completedAt: number | null;
    isFinalRefund: boolean;
  }>;
}

export interface CloseOrderRequest {
  orderNo: string;
}

export interface CloseOrderResponse {
  orderNo: string;
  status: OrderStatus;
  closedAt: number;
}

export interface CallbackRequest {
  channel: PaymentChannel;
  data: string;
}

export interface CallbackResponse {
  success: boolean;
  message: string;
}

export interface ApiResponse<T = unknown> {
  code: number;
  message: string;
  data: T | null;
}

export const MIN_AMOUNT: Amount = 1;
export const MAX_AMOUNT: Amount = 50000000;
export const WECHAT_THRESHOLD: Amount = 10000;
export const CALLBACK_TIMEOUT_SECONDS: number = 30;
export const SYNC_INTERVAL_HOURS: number = 24;
