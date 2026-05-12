export const ONE_MINUTE_MS = 60 * 1000;

export function now(): number {
  return Date.now();
}

export function getSlidingWindowStart(windowMs: number): number {
  return now() - windowMs;
}

export function addMinutes(timestamp: number, minutes: number): number {
  return timestamp + minutes * ONE_MINUTE_MS;
}

export function isExpired(timestamp?: number): boolean {
  if (timestamp === undefined || timestamp === null) {
    return false;
  }
  return now() > timestamp;
}

export function formatTimestamp(timestamp: number): string {
  return new Date(timestamp).toISOString();
}
