import {
  Order,
  OrderStatus,
  Amount,
  PaymentChannel,
  CreateOrderRequest,
  CreateOrderResponse,
  OrderQueryRequest,
  OrderQueryResponse,
  CloseOrderRequest,
  CloseOrderResponse,
  selectChannelByAmount,
  validateAmount,
  addAmount,
} from "@payment-gateway/shared";
import { ErrorCode, PaymentGatewayError } from "@payment-gateway/shared";
import { store } from "../store";
import { generatePaymentUrl } from "../channel-mock";

export function createOrder(request: CreateOrderRequest): CreateOrderResponse {
  if (!request.orderNo || typeof request.orderNo !== "string") {
    throw new PaymentGatewayError(ErrorCode.INVALID_PARAMS, "订单号不能为空");
  }

  validateAmount(request.amount);

  const existingOrder = store.getOrder(request.orderNo);
  if (existingOrder) {
    throw new PaymentGatewayError(ErrorCode.INVALID_PARAMS, `订单号 ${request.orderNo} 已存在`);
  }

  const channel = selectChannelByAmount(request.amount);
  const now = Date.now();

  const order: Order = {
    orderNo: request.orderNo,
    amount: request.amount,
    channel: channel,
    status: OrderStatus.PENDING,
    transactionNo: null,
    paidAt: null,
    createdAt: now,
    updatedAt: now,
    closedAt: null,
    lastSyncAt: null,
    totalRefundedAmount: 0,
    refundableAmount: request.amount,
  };

  store.saveOrder(order);

  const paymentUrl = generatePaymentUrl(channel, order.orderNo, order.amount);

  return {
    orderNo: order.orderNo,
    amount: order.amount,
    channel: order.channel,
    status: order.status,
    paymentUrl: paymentUrl,
    createdAt: order.createdAt,
  };
}

export function queryOrder(request: OrderQueryRequest): OrderQueryResponse {
  if (!request.orderNo) {
    throw new PaymentGatewayError(ErrorCode.INVALID_PARAMS, "订单号不能为空");
  }

  const order = store.getOrder(request.orderNo);
  if (!order) {
    throw new PaymentGatewayError(ErrorCode.ORDER_NOT_FOUND);
  }

  const refunds = store.getRefundsByOrder(order.orderNo);

  return {
    orderNo: order.orderNo,
    amount: order.amount,
    channel: order.channel,
    status: order.status,
    transactionNo: order.transactionNo,
    paidAt: order.paidAt,
    createdAt: order.createdAt,
    totalRefundedAmount: order.totalRefundedAmount,
    refundableAmount: order.refundableAmount,
    refunds: refunds.map((r) => ({
      refundNo: r.refundNo,
      amount: r.amount,
      status: r.status,
      createdAt: r.createdAt,
      completedAt: r.completedAt,
    })),
  };
}

export function closeOrder(request: CloseOrderRequest): CloseOrderResponse {
  if (!request.orderNo) {
    throw new PaymentGatewayError(ErrorCode.INVALID_PARAMS, "订单号不能为空");
  }

  const order = store.getOrder(request.orderNo);
  if (!order) {
    throw new PaymentGatewayError(ErrorCode.ORDER_NOT_FOUND);
  }

  if (order.status === OrderStatus.CLOSED) {
    throw new PaymentGatewayError(ErrorCode.ORDER_STATUS_ERROR, "订单已关闭");
  }

  if (order.status === OrderStatus.SUCCESS) {
    throw new PaymentGatewayError(ErrorCode.ORDER_STATUS_ERROR, "订单已支付成功，无法关闭");
  }

  if (order.status === OrderStatus.REFUNDED || order.status === OrderStatus.PARTIAL_REFUNDED) {
    throw new PaymentGatewayError(ErrorCode.ORDER_STATUS_ERROR, "订单已退款，无法关闭");
  }

  const now = Date.now();
  order.status = OrderStatus.CLOSED;
  order.closedAt = now;

  store.saveOrder(order);

  return {
    orderNo: order.orderNo,
    status: order.status,
    closedAt: order.closedAt,
  };
}

export function canPayOrder(order: Order): boolean {
  const allowedStatuses = [OrderStatus.PENDING];
  return allowedStatuses.includes(order.status);
}

export function canRefundOrder(order: Order): boolean {
  const allowedStatuses = [OrderStatus.SUCCESS, OrderStatus.PARTIAL_REFUNDED];
  return allowedStatuses.includes(order.status);
}

export function getOrder(orderNo: string): Order | null {
  return store.getOrder(orderNo);
}

export function saveOrder(order: Order): void {
  store.saveOrder(order);
}

export function getPendingConfirmOrders(): Order[] {
  return store.getPendingConfirmOrders();
}

export function getProcessingOrders(): Order[] {
  return store.getProcessingOrders();
}
