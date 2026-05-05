import * as crypto from "crypto";

const WECHAT_SECRET_KEY = "wechat_test_secret_key_123456";

const { privateKey, publicKey } = crypto.generateKeyPairSync("rsa", {
  modulusLength: 2048,
  publicKeyEncoding: {
    type: "pkcs1",
    format: "pem",
  },
  privateKeyEncoding: {
    type: "pkcs1",
    format: "pem",
  },
});

const ALIPAY_PRIVATE_KEY = privateKey;
const ALIPAY_PUBLIC_KEY = publicKey;

export const WECHAT_APP_ID = "wx1234567890abcdef";
export const WECHAT_MCH_ID = "1234567890";
export const ALIPAY_APP_ID = "2021001123654321";

export function getWechatSecretKey(): string {
  return WECHAT_SECRET_KEY;
}

export function getAlipayPrivateKey(): string {
  return ALIPAY_PRIVATE_KEY;
}

export function getAlipayPublicKey(): string {
  return ALIPAY_PUBLIC_KEY;
}

export const SERVER_PORT = 8080;
export const CALLBACK_CHECK_INTERVAL_MS = 1000;
export const DAILY_SYNC_HOUR = 0;
