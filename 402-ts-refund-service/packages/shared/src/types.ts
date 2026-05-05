export type RefundStatus = 'pending' | 'processing' | 'waiting_for_return' | 'waiting_for_warehouse_confirm' | 'refunded' | 'rejected';

export type ProductType = 'physical' | 'virtual';

export interface OrderProduct {
  productId: string;
  productName: string;
  productType: ProductType;
  unitPrice: number;
  quantity: number;
  paymentAmount: number;
  consumptionProgress?: number;
}

export interface Order {
  orderId: string;
  userId: string;
  orderTime: Date;
  totalAmount: number;
  shippingFee: number;
  products: OrderProduct[];
}

export interface RefundProduct {
  productId: string;
  productName: string;
  productType: ProductType;
  unitPrice: number;
  quantity: number;
  paymentAmount: number;
  refundAmount: number;
  consumptionProgress?: number;
}

export interface Refund {
  refundId: string;
  orderId: string;
  userId: string;
  refundReason: string;
  refundAmount: number;
  status: RefundStatus;
  products: RefundProduct[];
  isFullRefund: boolean;
  shippingFeeRefund: number;
  logisticsNumber?: string;
  createdAt: Date;
  updatedAt: Date;
  statusHistory: StatusHistoryItem[];
}

export interface StatusHistoryItem {
  status: RefundStatus;
  time: Date;
  note?: string;
}

export interface Config {
  refundPeriodDays: number;
  virtualProductRefundThreshold: number;
  dataStoragePath: string;
}
