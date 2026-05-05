import {
  Order,
  OrderStatus,
  Refund,
  RefundStatus,
  PaymentChannel,
  Amount,
  CreateRefundRequest,
  CreateRefundResponse,
  RefundQueryRequest,
  RefundQueryResponse,
  validateAmount,
  addAmount,
  subtractAmount,
  isAmountGreaterThan,
  isAmountGreaterOrEqual,
  generateId,
} from "@payment-gateway/shared";
import { ErrorCode, PaymentGatewayError } from "@payment-gateway/shared";
import { store } from "../store";
import { getOrder, saveOrder, canRefundOrder } from "./order-service";

function getPendingRefundsTotal(orderNo: string): Amount {
  const refunds = store.getRefundsByOrder(orderNo);
  let total: Amount = 0;
  for (const refund of refunds) {
    if (refund.status === RefundStatus.PENDING || refund.status === RefundStatus.PROCESSING) {
      total = addAmount(total, refund.amount);
    }
  }
  return total;
}

function getActualRefundableAmount(order: Order): Amount {
  const pendingTotal = getPendingRefundsTotal(order.orderNo);
  return subtractAmount(order.refundableAmount, pendingTotal);
}

export function createRefund(request: CreateRefundRequest): CreateRefundResponse {
  if (!request.orderNo) {
    throw new PaymentGatewayError(ErrorCode.INVALID_PARAMS, "订单号不能为空");
  }
  if (!request.refundNo) {
    throw new PaymentGatewayError(ErrorCode.INVALID_PARAMS, "退款单号不能为空");
  }

  validateAmount(request.amount);

  const order = getOrder(request.orderNo);
  if (!order) {
    throw new PaymentGatewayError(ErrorCode.ORDER_NOT_FOUND);
  }

  if (!canRefundOrder(order)) {
    throw new PaymentGatewayError(ErrorCode.ORDER_STATUS_ERROR, `订单状态为 ${order.status}，无法退款`);
  }

  const existingRefund = store.getRefund(request.refundNo);
  if (existingRefund) {
    throw new PaymentGatewayError(ErrorCode.INVALID_PARAMS, `退款单号 ${request.refundNo} 已存在`);
  }

  const refundAmount = request.amount;
  const actualRefundableAmount = getActualRefundableAmount(order);

  let actualRefundAmount = refundAmount;
  const isFinalRefund = request.isFinalRefund ?? false;

  if (isFinalRefund) {
    if (actualRefundableAmount <= 0) {
      throw new PaymentGatewayError(
        ErrorCode.REFUND_AMOUNT_EXCEEDED,
        `无可退款金额，当前待处理退款已占用全部可退款额度`
      );
    }
    actualRefundAmount = actualRefundableAmount;
  } else {
    if (isAmountGreaterThan(refundAmount, actualRefundableAmount)) {
      throw new PaymentGatewayError(
        ErrorCode.REFUND_AMOUNT_EXCEEDED,
        `退款金额 ${refundAmount} 分超过实际可退款金额 ${actualRefundableAmount} 分（含待处理退款）`
      );
    }
  }

  const now = Date.now();

  const refund: Refund = {
    refundNo: request.refundNo,
    orderNo: order.orderNo,
    channel: order.channel,
    amount: actualRefundAmount,
    status: RefundStatus.PENDING,
    channelRefundId: null,
    createdAt: now,
    completedAt: null,
    isFinalRefund: isFinalRefund,
  };

  store.saveRefund(refund);

  refund.status = RefundStatus.PROCESSING;
  store.saveRefund(refund);

  return {
    refundNo: refund.refundNo,
    orderNo: refund.orderNo,
    channel: refund.channel,
    amount: refund.amount,
    status: refund.status,
    createdAt: refund.createdAt,
  };
}

export function queryRefund(request: RefundQueryRequest): RefundQueryResponse {
  if (!request.orderNo) {
    throw new PaymentGatewayError(ErrorCode.INVALID_PARAMS, "订单号不能为空");
  }

  const order = getOrder(request.orderNo);
  if (!order) {
    throw new PaymentGatewayError(ErrorCode.ORDER_NOT_FOUND);
  }

  let refunds = store.getRefundsByOrder(order.orderNo);

  if (request.refundNo) {
    refunds = refunds.filter((r) => r.refundNo === request.refundNo);
  }

  return {
    orderNo: order.orderNo,
    totalAmount: order.amount,
    totalRefundedAmount: order.totalRefundedAmount,
    refundableAmount: order.refundableAmount,
    refunds: refunds.map((r) => ({
      refundNo: r.refundNo,
      amount: r.amount,
      status: r.status,
      channelRefundId: r.channelRefundId,
      createdAt: r.createdAt,
      completedAt: r.completedAt,
      isFinalRefund: r.isFinalRefund,
    })),
  };
}

export function confirmRefundSuccess(
  order: Order,
  refund: Refund,
  channelRefundId: string,
  completedAt: number
): void {
  const refundAmount = refund.amount;

  refund.status = RefundStatus.SUCCESS;
  refund.channelRefundId = channelRefundId;
  refund.completedAt = completedAt;
  store.saveRefund(refund);

  order.totalRefundedAmount = addAmount(order.totalRefundedAmount, refundAmount);
  order.refundableAmount = subtractAmount(order.refundableAmount, refundAmount);

  if (order.refundableAmount === 0) {
    order.status = OrderStatus.REFUNDED;
  } else {
    order.status = OrderStatus.PARTIAL_REFUNDED;
  }

  saveOrder(order);
}

export function confirmRefundFailed(refund: Refund): void {
  refund.status = RefundStatus.FAILED;
  store.saveRefund(refund);
}

export function getRefund(refundNo: string): Refund | null {
  return store.getRefund(refundNo);
}

export function getRefundsByOrder(orderNo: string): Refund[] {
  return store.getRefundsByOrder(orderNo);
}

export function getPendingRefunds(): Refund[] {
  return store
    .getAllRefunds()
    .filter((r) => r.status === RefundStatus.PENDING || r.status === RefundStatus.PROCESSING);
}
