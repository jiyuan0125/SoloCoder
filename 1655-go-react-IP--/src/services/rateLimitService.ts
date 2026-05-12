import { requestLogDAO } from '../daos/requestLogDAO';
import { RateLimitResult } from '../types';
import { getSlidingWindowStart, now } from '../utils/timeUtils';

const WINDOW_MS = 60 * 1000;
const MAX_REQUESTS = 60;

export const rateLimitService = {
  checkRateLimit(ip: string): RateLimitResult {
    const windowStart = getSlidingWindowStart(WINDOW_MS);
    const currentCount = requestLogDAO.getRequestsInWindow(ip, windowStart);
    const allowed = currentCount < MAX_REQUESTS;
    const remaining = Math.max(0, MAX_REQUESTS - currentCount - 1);
    const reset = windowStart + WINDOW_MS;

    if (allowed) {
      requestLogDAO.logRequest(ip, '', '');
    }

    this.cleanupOldLogs();

    return {
      allowed,
      remaining,
      reset,
      currentCount
    };
  },

  getCurrentCount(ip: string): number {
    const windowStart = getSlidingWindowStart(WINDOW_MS);
    return requestLogDAO.getRequestsInWindow(ip, windowStart);
  },

  getMaxRequests(): number {
    return MAX_REQUESTS;
  },

  getWindowMs(): number {
    return WINDOW_MS;
  },

  cleanupOldLogs(): void {
    const twoHoursAgo = now() - (2 * 60 * 60 * 1000);
    requestLogDAO.cleanupOldLogs(twoHoursAgo);
  }
};
