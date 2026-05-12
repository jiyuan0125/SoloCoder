import { store } from '../storage/store';

const ALERT_WINDOW_MS = 10 * 60 * 1000;
const FAILURE_RATE_THRESHOLD = 0.1;

export class AlertService {
  checkAndTrigger(channelId: string): boolean {
    const now = Date.now();
    const records = store.getChannelRecords(channelId);
    const windowRecords = records.filter(r => r.time > now - ALERT_WINDOW_MS);

    if (windowRecords.length === 0) return false;

    const failures = windowRecords.filter(r => !r.success).length;
    const failureRate = failures / windowRecords.length;

    if (failureRate >= FAILURE_RATE_THRESHOLD) {
      const alert = store.getAlert(channelId);
      if (!alert) {
        store.setAlert({
          channelId,
          startedAt: now,
          failureRate,
        });
        return true;
      }
    }
    return false;
  }

  getFailureRate(channelId: string): number {
    const now = Date.now();
    const records = store.getChannelRecords(channelId);
    const windowRecords = records.filter(r => r.time > now - ALERT_WINDOW_MS);
    if (windowRecords.length === 0) return 0;
    const failures = windowRecords.filter(r => !r.success).length;
    return failures / windowRecords.length;
  }
}

export const alertService = new AlertService();
