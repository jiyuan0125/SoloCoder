import bcrypt from 'bcryptjs';
import jwt from 'jsonwebtoken';
import crypto from 'crypto';
import { totp } from 'otplib';
import { encode as base32Encode } from 'hi-base32';
import { config } from '../config';

export function hashPassword(password: string): string {
  return bcrypt.hashSync(password, 10);
}

export function verifyPassword(password: string, hash: string): boolean {
  return bcrypt.compareSync(password, hash);
}

export interface JwtPayload {
  userId: number;
  username: string;
  iat: number;
  exp: number;
}

export function generateToken(userId: number, username: string): string {
  const now = Math.floor(Date.now() / 1000);
  const payload: JwtPayload = {
    userId,
    username,
    iat: now,
    exp: now + config.jwtExpiresInSeconds
  };
  return jwt.sign(payload, config.jwtSecret, { algorithm: 'HS256' });
}

export type TokenVerifyResult = 
  | { valid: true; payload: JwtPayload }
  | { valid: false; error: 'expired' | 'invalid' | 'malformed' };

export function verifyToken(token: string): TokenVerifyResult {
  try {
    const payload = jwt.verify(token, config.jwtSecret, { algorithms: ['HS256'] }) as JwtPayload;
    return { valid: true, payload };
  } catch (err) {
    if (err instanceof jwt.TokenExpiredError) {
      return { valid: false, error: 'expired' };
    }
    if (err instanceof jwt.JsonWebTokenError) {
      return { valid: false, error: 'invalid' };
    }
    return { valid: false, error: 'malformed' };
  }
}

export function generateMfaSecret(): string {
  const bytes = crypto.randomBytes(20);
  return base32Encode(bytes).replace(/=/g, '');
}

export function generateMfaUri(username: string, secret: string, issuer: string = 'AuthCenter'): string {
  return `otpauth://totp/${encodeURIComponent(issuer)}:${encodeURIComponent(username)}?secret=${secret}&issuer=${encodeURIComponent(issuer)}`;
}

export function verifyTotp(secret: string, token: string): boolean {
  const customTotp = totp.clone();
  customTotp.options = {
    window: config.totpWindow,
    step: 30,
    digits: 6
  };
  return customTotp.verify({ secret, token });
}

export function generateRandomId(length: number = 32): string {
  return crypto.randomBytes(length).toString('hex');
}

export function hmacSign(data: string, secret: string): string {
  return crypto.createHmac('sha256', secret).update(data).digest('hex');
}

export function hmacVerify(data: string, signature: string, secret: string): boolean {
  const expected = crypto.createHmac('sha256', secret).update(data).digest('hex');
  return crypto.timingSafeEqual(Buffer.from(expected, 'hex'), Buffer.from(signature, 'hex'));
}

export function nowSeconds(): number {
  return Math.floor(Date.now() / 1000);
}

export function formatMinutes(seconds: number): string {
  const mins = Math.ceil(seconds / 60);
  return `${mins} 分钟`;
}
