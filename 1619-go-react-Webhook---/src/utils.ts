import crypto from 'crypto';
import { v4 as uuidv4 } from 'uuid';
import { SUPPORTED_EVENT_TYPES, RETRY_INTERVALS, MAX_RETRIES } from './types';

export function generateSecret(): string {
  return uuidv4().replace(/-/g, '') + uuidv4().replace(/-/g, '');
}

export function isValidEventTypes(eventType: string): boolean {
  return SUPPORTED_EVENT_TYPES.includes(eventType);
}

export function isValidCallbackUrl(url: string): boolean {
  try {
    const parsed = new URL(url);
    if (parsed.protocol === 'https:') return true;
    if (parsed.hostname === 'localhost' || parsed.hostname === '127.0.0.1') {
      return true;
    }
    return false;
  } catch {
    return false;
  }
}

export function computeSignature(timestamp: number, payloadString: string, secret: string): string {
  const signatureString = `${timestamp}${payloadString}`;
  return crypto
    .createHmac('sha256', secret)
    .update(signatureString)
    .digest('hex');
}

export function getNextRetryInterval(retryCount: number): number | null {
  if (retryCount >= MAX_RETRIES) return null;
  return RETRY_INTERVALS[retryCount] || null;
}

export function addMinutes(isoString: string | null, milliseconds: number): string {
  const base = isoString ? new Date(isoString) : new Date();
  return new Date(base.getTime() + milliseconds).toISOString();
}
