import { Amount, MAX_AMOUNT, MIN_AMOUNT, PaymentChannel, WECHAT_THRESHOLD } from "./types";
import { ErrorCode, PaymentGatewayError } from "./errors";
import * as crypto from "crypto";

export function generateId(): string {
  const timestamp = Date.now().toString(36);
  const random = Math.random().toString(36).substring(2, 10);
  return `${timestamp}-${random}`;
}

export function generateTransactionNo(channel: PaymentChannel): string {
  const prefix = channel === PaymentChannel.WECHAT ? "WX" : "ALI";
  const timestamp = Date.now().toString();
  const random = Math.random().toString().substring(2, 8);
  return `${prefix}${timestamp}${random}`;
}

export function generateChannelPaymentId(channel: PaymentChannel): string {
  const prefix = channel === PaymentChannel.WECHAT ? "wxpay_" : "alipay_";
  return `${prefix}${Date.now()}_${Math.random().toString(36).substring(2, 10)}`;
}

export function generateChannelRefundId(channel: PaymentChannel): string {
  const prefix = channel === PaymentChannel.WECHAT ? "wxrefund_" : "alirefund_";
  return `${prefix}${Date.now()}_${Math.random().toString(36).substring(2, 10)}`;
}

export function selectChannelByAmount(amount: Amount): PaymentChannel {
  if (amount < WECHAT_THRESHOLD) {
    return PaymentChannel.WECHAT;
  }
  return PaymentChannel.ALIPAY;
}

export function validateAmount(amount: Amount): void {
  if (!Number.isInteger(amount)) {
    throw new PaymentGatewayError(ErrorCode.INVALID_AMOUNT, "金额必须为整数(单位:分)");
  }
  if (amount < MIN_AMOUNT) {
    throw new PaymentGatewayError(ErrorCode.INVALID_AMOUNT, `金额必须大于0分，当前: ${amount}分`);
  }
  if (amount > MAX_AMOUNT) {
    throw new PaymentGatewayError(ErrorCode.INVALID_AMOUNT, `金额不能超过${MAX_AMOUNT}分(50万)，当前: ${amount}分`);
  }
}

export function addAmount(a: Amount, b: Amount): Amount {
  return Math.floor(a) + Math.floor(b);
}

export function subtractAmount(a: Amount, b: Amount): Amount {
  return Math.floor(a) - Math.floor(b);
}

export function isAmountGreaterThan(a: Amount, b: Amount): boolean {
  return Math.floor(a) > Math.floor(b);
}

export function isAmountGreaterOrEqual(a: Amount, b: Amount): boolean {
  return Math.floor(a) >= Math.floor(b);
}

export function generateCallbackKey(orderNo: string, channel: PaymentChannel): string {
  return `${channel}:${orderNo}`;
}

export function generateWechatSignature(
  data: Record<string, string | number>,
  secretKey: string
): string {
  const sortedKeys = Object.keys(data).filter((k) => data[k] !== null && data[k] !== undefined).sort();
  const signStr = sortedKeys
    .map((k) => `${k}=${data[k]}`)
    .join("&")
    .concat(`&key=${secretKey}`);
  return crypto.createHash("md5").update(signStr, "utf8").digest("hex").toUpperCase();
}

export function generateAlipaySignature(
  data: Record<string, string | number>,
  privateKey: string
): string {
  const sortedKeys = Object.keys(data).filter((k) => data[k] !== null && data[k] !== undefined).sort();
  const signStr = sortedKeys.map((k) => `${k}=${data[k]}`).join("&");
  const sign = crypto
    .createSign("RSA-SHA256")
    .update(signStr, "utf8")
    .sign(privateKey, "base64");
  return sign;
}

export function verifyWechatSignature(
  data: Record<string, string | number>,
  signature: string,
  secretKey: string
): boolean {
  const { signature: _, ...dataWithoutSign } = data;
  const expectedSign = generateWechatSignature(dataWithoutSign, secretKey);
  return expectedSign === signature;
}

export function verifyAlipaySignature(
  data: Record<string, string | number>,
  signature: string,
  publicKey: string
): boolean {
  const { signature: _, ...dataWithoutSign } = data;
  const sortedKeys = Object.keys(dataWithoutSign)
    .filter((k) => dataWithoutSign[k] !== null && dataWithoutSign[k] !== undefined)
    .sort();
  const signStr = sortedKeys.map((k) => `${k}=${dataWithoutSign[k]}`).join("&");
  try {
    const verify = crypto.createVerify("RSA-SHA256");
    verify.update(signStr, "utf8");
    return verify.verify(publicKey, signature, "base64");
  } catch {
    return false;
  }
}
