interface RateLimitRecord {
  count: number;
  windowStart: number;
}

const WINDOW_SIZE_MS = 10 * 1000;
const MAX_REQUESTS_PER_WINDOW = 20;

const store = new Map<string, RateLimitRecord>();

export function checkRateLimit(userId: string): { allowed: boolean; remaining: number; resetTime: number } {
  const now = Date.now();
  let record = store.get(userId);

  if (!record || now - record.windowStart >= WINDOW_SIZE_MS) {
    record = { count: 0, windowStart: now };
    store.set(userId, record);
  }

  if (record.count >= MAX_REQUESTS_PER_WINDOW) {
    const resetTime = record.windowStart + WINDOW_SIZE_MS;
    return { allowed: false, remaining: 0, resetTime };
  }

  record.count++;
  const remaining = MAX_REQUESTS_PER_WINDOW - record.count;
  const resetTime = record.windowStart + WINDOW_SIZE_MS;

  return { allowed: true, remaining, resetTime };
}

export function cleanupExpired(): void {
  const now = Date.now();
  for (const [userId, record] of store.entries()) {
    if (now - record.windowStart >= WINDOW_SIZE_MS) {
      store.delete(userId);
    }
  }
}

setInterval(cleanupExpired, WINDOW_SIZE_MS);
