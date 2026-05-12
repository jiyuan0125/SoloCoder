import { ChannelType } from '../types';
import { store } from '../storage/store';

const ONE_MINUTE = 60 * 1000;
const ONE_HOUR = 60 * ONE_MINUTE;
const ONE_DAY = 24 * ONE_HOUR;

export interface RateLimitResult {
  allowed: boolean;
  waitSeconds?: number;
}

export class RateLimiter {
  check(
    userId: string,
    channelType: ChannelType,
    limits: { perMinute: number; perHour: number; perDay: number },
    now: number = Date.now()
  ): RateLimitResult {
    const records = store.getSendRecords(userId, channelType);

    const minuteRecords = records.filter(t => t > now - ONE_MINUTE);
    if (minuteRecords.length >= limits.perMinute) {
      const oldestInMinute = Math.min(...minuteRecords);
      const waitMs = oldestInMinute + ONE_MINUTE - now;
      return { allowed: false, waitSeconds: Math.ceil(waitMs / 1000) };
    }

    const hourRecords = records.filter(t => t > now - ONE_HOUR);
    if (hourRecords.length >= limits.perHour) {
      const oldestInHour = Math.min(...hourRecords);
      const waitMs = oldestInHour + ONE_HOUR - now;
      return { allowed: false, waitSeconds: Math.ceil(waitMs / 1000) };
    }

    const dayRecords = records.filter(t => t > now - ONE_DAY);
    if (dayRecords.length >= limits.perDay) {
      const oldestInDay = Math.min(...dayRecords);
      const waitMs = oldestInDay + ONE_DAY - now;
      return { allowed: false, waitSeconds: Math.ceil(waitMs / 1000) };
    }

    return { allowed: true };
  }

  record(userId: string, channelType: ChannelType): void {
    store.addSendRecord(userId, channelType, Date.now());
  }
}

export const rateLimiter = new RateLimiter();
