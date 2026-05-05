import {
  Order,
  OrderStatus,
  CallbackRecord,
  PaymentChannel,
  Amount,
  WechatCallbackData,
  AlipayCallbackData,
  ChannelRefundCallbackData,
  generateCallbackKey,
  verifyWechatSignature,
  verifyAlipaySignature,
  generateId,
} from "@payment-gateway/shared";
import { ErrorCode, PaymentGatewayError } from "@payment-gateway/shared";
import { store } from "../store";
import { getOrder, saveOrder } from "./order-service";
import { confirmPaymentSuccess, confirmPaymentFailed } from "./payment-service";
import { getRefund, confirmRefundSuccess, confirmRefundFailed } from "./refund-service";
import { getWechatSecretKey, getAlipayPublicKey } from "../config";

export function processPaymentCallback(
  channel: PaymentChannel,
  callbackData: WechatCallbackData | AlipayCallbackData
): void {
  const orderNo = callbackData.orderNo;
  const order = getOrder(orderNo);

  if (!order) {
    throw new PaymentGatewayError(ErrorCode.ORDER_NOT_FOUND);
  }

  const existingRecord = store.getCallbackRecord(orderNo, channel);
  if (existingRecord) {
    throw new PaymentGatewayError(ErrorCode.CALLBACK_DUPLICATE);
  }

  const signatureValid = verifyPaymentSignature(channel, callbackData);
  if (!signatureValid) {
    throw new PaymentGatewayError(ErrorCode.SIGNATURE_INVALID);
  }

  const record: CallbackRecord = {
    id: generateId(),
    orderNo: orderNo,
    channel: channel,
    rawData: JSON.stringify(callbackData),
    processedAt: Date.now(),
  };

  store.saveCallbackRecord(record);

  if (callbackData.status === "success") {
    confirmPaymentSuccess(order, callbackData.transactionId, callbackData.timestamp * 1000);
  } else {
    confirmPaymentFailed(order);
  }
}

export function processRefundCallback(
  channel: PaymentChannel,
  callbackData: ChannelRefundCallbackData
): void {
  const orderNo = callbackData.orderNo;
  const refundNo = callbackData.refundNo;

  const order = getOrder(orderNo);
  if (!order) {
    throw new PaymentGatewayError(ErrorCode.ORDER_NOT_FOUND);
  }

  const refund = getRefund(refundNo);
  if (!refund) {
    throw new PaymentGatewayError(ErrorCode.REFUND_NOT_FOUND);
  }

  const existingRecord = store.getCallbackRecord(`${orderNo}:${refundNo}`, channel);
  if (existingRecord) {
    throw new PaymentGatewayError(ErrorCode.CALLBACK_DUPLICATE);
  }

  const signatureValid = verifyRefundSignature(channel, callbackData);
  if (!signatureValid) {
    throw new PaymentGatewayError(ErrorCode.SIGNATURE_INVALID);
  }

  const record: CallbackRecord = {
    id: generateId(),
    orderNo: `${orderNo}:${refundNo}`,
    channel: channel,
    rawData: JSON.stringify(callbackData),
    processedAt: Date.now(),
  };

  store.saveCallbackRecord(record);

  if (callbackData.status === "success") {
    confirmRefundSuccess(order, refund, callbackData.channelRefundId, callbackData.timestamp * 1000);
  } else {
    confirmRefundFailed(refund);
  }
}

function verifyPaymentSignature(
  channel: PaymentChannel,
  data: WechatCallbackData | AlipayCallbackData
): boolean {
  const { signature, ...dataWithoutSign } = data;

  if (channel === PaymentChannel.WECHAT) {
    const wechatData = dataWithoutSign as Omit<WechatCallbackData, "signature">;
    const signData: Record<string, string | number> = {
      appId: wechatData.appId,
      mchId: wechatData.mchId,
      orderNo: wechatData.orderNo,
      transactionId: wechatData.transactionId,
      amount: wechatData.amount,
      status: wechatData.status,
      timestamp: wechatData.timestamp,
      nonceStr: wechatData.nonceStr,
    };
    return verifyWechatSignature(signData, signature, getWechatSecretKey());
  } else {
    const alipayData = dataWithoutSign as Omit<AlipayCallbackData, "signature">;
    const signData: Record<string, string | number> = {
      tradeNo: alipayData.tradeNo,
      orderNo: alipayData.orderNo,
      transactionId: alipayData.transactionId,
      amount: alipayData.amount,
      status: alipayData.status,
      timestamp: alipayData.timestamp,
      notifyId: alipayData.notifyId,
      sellerId: alipayData.sellerId,
    };
    return verifyAlipaySignature(signData, signature, getAlipayPublicKey());
  }
}

function verifyRefundSignature(
  channel: PaymentChannel,
  data: ChannelRefundCallbackData
): boolean {
  const { signature, ...dataWithoutSign } = data;
  const signData: Record<string, string | number> = {
    orderNo: dataWithoutSign.orderNo,
    refundNo: dataWithoutSign.refundNo,
    channelRefundId: dataWithoutSign.channelRefundId,
    amount: dataWithoutSign.amount,
    status: dataWithoutSign.status,
    timestamp: dataWithoutSign.timestamp,
  };

  if (channel === PaymentChannel.WECHAT) {
    return verifyWechatSignature(signData, signature, getWechatSecretKey());
  } else {
    return verifyAlipaySignature(signData, signature, getAlipayPublicKey());
  }
}

export function isCallbackProcessed(orderNo: string, channel: PaymentChannel): boolean {
  return store.getCallbackRecord(orderNo, channel) !== null;
}
