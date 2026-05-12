import { getOrCreateHealthMetric, recordHealthMetric, getAllChannels, updateChannelDegraded, getHealthMetricsSince } from './db';

const TEN_MINUTES_MS = 10 * 60 * 1000;
const MIN_SUCCESS_RATE = 0.95;

function getCurrentHour(): number {
  const now = new Date();
  return new Date(now.getFullYear(), now.getMonth(), now.getDate(), now.getHours()).getTime();
}

export function recordChannelRequest(channelId: string, success: boolean, responseTimeMs: number): void {
  const hour = getCurrentHour();
  getOrCreateHealthMetric(channelId, hour);
  recordHealthMetric(channelId, hour, success, responseTimeMs);
}

export function checkAndUpdateDegradation(): void {
  const channels = getAllChannels();
  const now = Date.now();
  const since = now - TEN_MINUTES_MS;
  const sinceHour = new Date(new Date(since).setMinutes(0, 0, 0)).getTime();

  for (const channel of channels) {
    const recentRate = getRecentSuccessRate(channel.id, sinceHour, since);
    const shouldBeDegraded = recentRate !== null && recentRate < MIN_SUCCESS_RATE;

    if (shouldBeDegraded !== channel.isDegraded) {
      updateChannelDegraded(channel.id, shouldBeDegraded);
    }
  }
}

export function getRecentSuccessRate(channelId: string, sinceHour: number, since: number): number | null {
  const metrics = getHealthMetricsSince(channelId, sinceHour);

  if (metrics.length === 0) return null;

  const tenMinAgo = since;
  const currentHourStart = getCurrentHour();
  let totalRequests = 0;
  let successCount = 0;

  for (const m of metrics) {
    if (m.hour < currentHourStart) {
      totalRequests += m.totalRequests;
      successCount += m.successCount;
    } else {
      const pctSince = Math.max(0, (Date.now() - tenMinAgo) / (1000 * 60 * 60));
      totalRequests += Math.floor(m.totalRequests * pctSince);
      successCount += Math.floor(m.successCount * pctSince);
    }
  }

  if (totalRequests === 0) return null;
  return successCount / totalRequests;
}

export function startDegradationChecker(): void {
  setInterval(checkAndUpdateDegradation, 60 * 1000);
}
