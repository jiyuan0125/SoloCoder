import {
  PaymentChannel,
  Amount,
  WechatCallbackData,
  AlipayCallbackData,
  ChannelRefundCallbackData,
  generateWechatSignature,
  generateAlipaySignature,
  generateTransactionNo,
  generateChannelPaymentId,
  generateChannelRefundId,
} from "@payment-gateway/shared";
import { getWechatSecretKey, getAlipayPrivateKey, WECHAT_APP_ID, WECHAT_MCH_ID, ALIPAY_APP_ID } from "./config";

export function generateWechatPayUrl(orderNo: string, amount: Amount): string {
  return `weixin://wxpay/bizpayurl?pr=${orderNo}&amt=${amount}`;
}

export function generateAlipayPayUrl(orderNo: string, amount: Amount): string {
  return `https://qr.alipay.com/bax${orderNo}xxxxx`;
}

export function generatePaymentUrl(channel: PaymentChannel, orderNo: string, amount: Amount): string {
  if (channel === PaymentChannel.WECHAT) {
    return generateWechatPayUrl(orderNo, amount);
  }
  return generateAlipayPayUrl(orderNo, amount);
}

export function simulateWechatPaymentCallback(
  orderNo: string,
  amount: Amount,
  success: boolean = true
): WechatCallbackData {
  const transactionId = generateTransactionNo(PaymentChannel.WECHAT);
  const timestamp = Math.floor(Date.now() / 1000);
  const nonceStr = Math.random().toString(36).substring(2, 10);

  const data: WechatCallbackData = {
    channel: PaymentChannel.WECHAT,
    orderNo: orderNo,
    status: success ? "success" : "failed",
    transactionId: transactionId,
    amount: amount,
    timestamp: timestamp,
    appId: WECHAT_APP_ID,
    mchId: WECHAT_MCH_ID,
    nonceStr: nonceStr,
    signature: "",
  };

  const signData: Record<string, string | number> = {
    appId: data.appId,
    mchId: data.mchId,
    orderNo: data.orderNo,
    transactionId: data.transactionId,
    amount: data.amount,
    status: data.status,
    timestamp: data.timestamp,
    nonceStr: data.nonceStr,
  };

  data.signature = generateWechatSignature(signData, getWechatSecretKey());

  return data;
}

export function simulateAlipayPaymentCallback(
  orderNo: string,
  amount: Amount,
  success: boolean = true
): AlipayCallbackData {
  const tradeNo = generateTransactionNo(PaymentChannel.ALIPAY);
  const timestamp = Math.floor(Date.now() / 1000);
  const notifyId = `notify_${Date.now()}`;
  const sellerId = "2088123456789012";

  const data: AlipayCallbackData = {
    channel: PaymentChannel.ALIPAY,
    orderNo: orderNo,
    status: success ? "success" : "failed",
    transactionId: tradeNo,
    tradeNo: tradeNo,
    amount: amount,
    timestamp: timestamp,
    notifyId: notifyId,
    sellerId: sellerId,
    signature: "",
  };

  const signData: Record<string, string | number> = {
    appId: ALIPAY_APP_ID,
    tradeNo: data.tradeNo,
    orderNo: data.orderNo,
    transactionId: data.transactionId,
    amount: data.amount,
    status: data.status,
    timestamp: data.timestamp,
    notifyId: data.notifyId,
    sellerId: data.sellerId,
  };

  data.signature = generateAlipaySignature(signData, getAlipayPrivateKey());

  return data;
}

export function simulateWechatRefundCallback(
  orderNo: string,
  refundNo: string,
  amount: Amount,
  success: boolean = true
): ChannelRefundCallbackData {
  const channelRefundId = generateChannelRefundId(PaymentChannel.WECHAT);
  const timestamp = Math.floor(Date.now() / 1000);

  const data: ChannelRefundCallbackData = {
    orderNo: orderNo,
    refundNo: refundNo,
    channel: PaymentChannel.WECHAT,
    status: success ? "success" : "failed",
    channelRefundId: channelRefundId,
    amount: amount,
    timestamp: timestamp,
    signature: "",
  };

  const signData: Record<string, string | number> = {
    orderNo: data.orderNo,
    refundNo: data.refundNo,
    channelRefundId: data.channelRefundId,
    amount: data.amount,
    status: data.status,
    timestamp: data.timestamp,
  };

  data.signature = generateWechatSignature(signData, getWechatSecretKey());

  return data;
}

export function simulateAlipayRefundCallback(
  orderNo: string,
  refundNo: string,
  amount: Amount,
  success: boolean = true
): ChannelRefundCallbackData {
  const channelRefundId = generateChannelRefundId(PaymentChannel.ALIPAY);
  const timestamp = Math.floor(Date.now() / 1000);

  const data: ChannelRefundCallbackData = {
    orderNo: orderNo,
    refundNo: refundNo,
    channel: PaymentChannel.ALIPAY,
    status: success ? "success" : "failed",
    channelRefundId: channelRefundId,
    amount: amount,
    timestamp: timestamp,
    signature: "",
  };

  const signData: Record<string, string | number> = {
    orderNo: data.orderNo,
    refundNo: data.refundNo,
    channelRefundId: data.channelRefundId,
    amount: data.amount,
    status: data.status,
    timestamp: data.timestamp,
  };

  data.signature = generateAlipaySignature(signData, getAlipayPrivateKey());

  return data;
}

export function simulateChannelQueryOrderStatus(
  channel: PaymentChannel,
  orderNo: string
): { status: "success" | "pending" | "failed"; transactionId?: string; paidAt?: number } {
  const random = Math.random();
  if (random < 0.7) {
    return {
      status: "success",
      transactionId: generateTransactionNo(channel),
      paidAt: Date.now() - 60000,
    };
  } else if (random < 0.9) {
    return {
      status: "pending",
    };
  } else {
    return {
      status: "failed",
    };
  }
}

export function simulateChannelQueryRefundStatus(
  channel: PaymentChannel,
  refundNo: string
): { status: "success" | "pending" | "failed"; channelRefundId?: string; completedAt?: number } {
  const random = Math.random();
  if (random < 0.7) {
    return {
      status: "success",
      channelRefundId: generateChannelRefundId(channel),
      completedAt: Date.now() - 30000,
    };
  } else if (random < 0.9) {
    return {
      status: "pending",
    };
  } else {
    return {
      status: "failed",
    };
  }
}
