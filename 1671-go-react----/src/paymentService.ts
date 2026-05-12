import { v4 as uuidv4 } from 'uuid';
import { PaymentRecord, PaymentStatus, CreatePaymentRequest, RefundRequest } from './types';
import {
  getPaymentByMerchantOrderNo,
  insertPayment,
  updatePaymentStatus,
  getPaymentById,
  incrementRefundedAmount,
} from './db';
import { routePayment, AllChannelsDegradedError, ChannelNotFoundError } from './routing';
import { chargeViaChannel, refundViaChannel } from './channelAdapter';
import { recordChannelRequest } from './healthMonitor';

export class MerchantOrderExistsError extends Error {}
export class InvalidAmountError extends Error {}
export class PaymentNotFoundError extends Error {}
export class InvalidStatusError extends Error {}
export class RefundExceedsAmountError extends Error {}
export class AlreadyRefundedError extends Error {}

const ALLOWED_TRANSITIONS: Record<PaymentStatus, PaymentStatus[]> = {
  PENDING: ['PAID', 'CLOSED'],
  PAID: ['CLOSED', 'REFUNDED'],
  CLOSED: [],
  REFUNDED: [],
};

function canTransition(from: PaymentStatus, to: PaymentStatus): boolean {
  return ALLOWED_TRANSITIONS[from].includes(to);
}

async function attemptReverseRefund(
  channelId: string,
  channelTransactionNo: string,
  amount: number
): Promise<void> {
  try {
    await refundViaChannel(channelId, channelTransactionNo, amount);
  } catch (e) {
    console.error('反向退款失败，请人工处理', {
      channelId,
      channelTransactionNo,
      amount,
      error: e,
    });
  }
}

export async function createPayment(req: CreatePaymentRequest): Promise<PaymentRecord> {
  if (req.amount <= 0) {
    throw new InvalidAmountError('金额必须大于0');
  }

  const existing = getPaymentByMerchantOrderNo(req.merchantOrderNo);
  if (existing) {
    throw new MerchantOrderExistsError('商户订单号已存在');
  }

  const routingResult = routePayment(req.amount, req.strategy, req.channelId);
  const channel = routingResult.channel;
  const paymentId = uuidv4();
  const now = Date.now();

  const pendingPayment: PaymentRecord = {
    id: paymentId,
    merchantOrderNo: req.merchantOrderNo,
    amount: req.amount,
    channelId: channel.id,
    channelTransactionNo: null,
    status: 'PENDING',
    refundedAmount: 0,
    createdAt: now,
    paidAt: null,
    closedAt: null,
    refundedAt: null,
  };

  insertPayment(pendingPayment);

  let chargeResult;
  try {
    chargeResult = await chargeViaChannel(channel.id, req.amount);
  } catch (err) {
    updatePaymentStatus(pendingPayment.id, 'CLOSED');
    throw err;
  }

  recordChannelRequest(channel.id, chargeResult.success, chargeResult.responseTimeMs);

  if (chargeResult.success) {
    try {
      updatePaymentStatus(pendingPayment.id, 'PAID', {
        channelTransactionNo: chargeResult.transactionNo,
      });
      return {
        ...pendingPayment,
        status: 'PAID',
        channelTransactionNo: chargeResult.transactionNo,
        paidAt: Date.now(),
      };
    } catch (dbError) {
      await attemptReverseRefund(channel.id, chargeResult.transactionNo, req.amount);
      updatePaymentStatus(pendingPayment.id, 'CLOSED');
      throw new Error('支付记录保存失败，已触发反向退款');
    }
  } else {
    updatePaymentStatus(pendingPayment.id, 'CLOSED');
    return {
      ...pendingPayment,
      status: 'CLOSED',
      closedAt: Date.now(),
    };
  }
}

export async function processRefund(req: RefundRequest): Promise<PaymentRecord> {
  if (req.refundAmount <= 0) {
    throw new InvalidAmountError('退款金额必须大于0');
  }

  const payment = getPaymentById(req.paymentId);
  if (!payment) {
    throw new PaymentNotFoundError('支付记录不存在');
  }

  if (payment.status === 'REFUNDED') {
    throw new AlreadyRefundedError('该支付已全额退款');
  }

  if (payment.status !== 'PAID') {
    throw new InvalidStatusError('只有已支付的订单可以退款');
  }

  const remainingAmount = payment.amount - payment.refundedAmount;
  if (req.refundAmount > remainingAmount) {
    throw new RefundExceedsAmountError('退款金额超过可退款金额');
  }

  if (!payment.channelTransactionNo) {
    throw new InvalidStatusError('缺少渠道流水号，无法退款');
  }

  const refundResult = await refundViaChannel(
    payment.channelId,
    payment.channelTransactionNo,
    req.refundAmount
  );

  recordChannelRequest(payment.channelId, refundResult.success, refundResult.responseTimeMs);

  if (!refundResult.success) {
    throw new Error('渠道退款失败');
  }

  incrementRefundedAmount(payment.id, req.refundAmount);

  const updatedPayment = getPaymentById(payment.id)!;
  if (updatedPayment.refundedAmount >= updatedPayment.amount) {
    updatePaymentStatus(updatedPayment.id, 'REFUNDED');
    return { ...updatedPayment, status: 'REFUNDED', refundedAt: Date.now() };
  }

  return updatedPayment;
}

export function getPayment(id: string): PaymentRecord | null {
  return getPaymentById(id);
}

export { AllChannelsDegradedError, ChannelNotFoundError };
