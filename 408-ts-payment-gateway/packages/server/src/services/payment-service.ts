import {
  Order,
  OrderStatus,
  Payment,
  PaymentChannel,
  Amount,
  PayOrderRequest,
  PayOrderResponse,
  generateId,
  generateTransactionNo,
} from "@payment-gateway/shared";
import { ErrorCode, PaymentGatewayError } from "@payment-gateway/shared";
import { store } from "../store";
import { getOrder, saveOrder, canPayOrder } from "./order-service";
import { generatePaymentUrl } from "../channel-mock";

export function payOrder(request: PayOrderRequest): PayOrderResponse {
  if (!request.orderNo) {
    throw new PaymentGatewayError(ErrorCode.INVALID_PARAMS, "订单号不能为空");
  }

  const order = getOrder(request.orderNo);
  if (!order) {
    throw new PaymentGatewayError(ErrorCode.ORDER_NOT_FOUND);
  }

  if (order.status === OrderStatus.CLOSED) {
    throw new PaymentGatewayError(ErrorCode.ORDER_CLOSED);
  }

  if (!canPayOrder(order)) {
    if (order.status === OrderStatus.PROCESSING || order.status === OrderStatus.PENDING_CONFIRM) {
      const payments = store.getPaymentsByOrder(order.orderNo);
      if (payments.length > 0) {
        const payment = payments[payments.length - 1];
        const paymentUrl = generatePaymentUrl(order.channel, order.orderNo, order.amount);
        return {
          orderNo: order.orderNo,
          channel: order.channel,
          amount: order.amount,
          status: order.status,
          paymentUrl: paymentUrl,
          paymentId: payment.paymentId,
        };
      }
    }
    throw new PaymentGatewayError(ErrorCode.ORDER_STATUS_ERROR, `订单状态为 ${order.status}，无法发起支付`);
  }

  const paymentId = generateId();
  const now = Date.now();

  const payment: Payment = {
    paymentId: paymentId,
    orderNo: order.orderNo,
    channel: order.channel,
    amount: order.amount,
    status: OrderStatus.PROCESSING,
    channelPaymentId: null,
    createdAt: now,
    paidAt: null,
  };

  store.savePayment(payment);

  order.status = OrderStatus.PROCESSING;
  saveOrder(order);

  const paymentUrl = generatePaymentUrl(order.channel, order.orderNo, order.amount);

  return {
    orderNo: order.orderNo,
    channel: order.channel,
    amount: order.amount,
    status: order.status,
    paymentUrl: paymentUrl,
    paymentId: payment.paymentId,
  };
}

export function confirmPaymentSuccess(
  order: Order,
  transactionNo: string,
  paidAt: number
): void {
  order.status = OrderStatus.SUCCESS;
  order.transactionNo = transactionNo;
  order.paidAt = paidAt;
  saveOrder(order);

  const payments = store.getPaymentsByOrder(order.orderNo);
  for (const payment of payments) {
    if (payment.status !== OrderStatus.SUCCESS) {
      payment.status = OrderStatus.SUCCESS;
      payment.paidAt = paidAt;
      payment.channelPaymentId = transactionNo;
      store.savePayment(payment);
    }
  }
}

export function confirmPaymentFailed(order: Order): void {
  order.status = OrderStatus.FAILED;
  saveOrder(order);

  const payments = store.getPaymentsByOrder(order.orderNo);
  for (const payment of payments) {
    if (payment.status === OrderStatus.PROCESSING) {
      payment.status = OrderStatus.FAILED;
      store.savePayment(payment);
    }
  }
}

export function setOrderPendingConfirm(order: Order): void {
  order.status = OrderStatus.PENDING_CONFIRM;
  saveOrder(order);
}

export function getPayment(paymentId: string): Payment | null {
  return store.getPayment(paymentId);
}

export function getPaymentsByOrder(orderNo: string): Payment[] {
  return store.getPaymentsByOrder(orderNo);
}
