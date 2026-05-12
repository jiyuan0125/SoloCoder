import { randomBytes } from 'crypto';

export function generateRandomString(length: number = 64): string {
  return randomBytes(length).toString('hex');
}

export function generateClientId(): string {
  return generateRandomString(32);
}

export function generateClientSecret(): string {
  return generateRandomString(48);
}

export function generateAuthorizationCode(): string {
  return generateRandomString(64);
}

export function generateAccessToken(): string {
  return generateRandomString(48);
}

export function generateRefreshToken(): string {
  return generateRandomString(64);
}
