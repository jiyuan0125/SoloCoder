import { Usage } from '../types';
import { store } from '../data/store';
import { getFirstDayOfMonth, getLastDayOfMonth } from '../utils/date';

export class UsageService {
  recordApiCall(subscriptionId: string, count: number = 1): Usage {
    return this.recordUsage(subscriptionId, count, 0);
  }

  recordStorage(subscriptionId: string, storageGB: number): Usage {
    return this.recordUsage(subscriptionId, 0, storageGB);
  }

  private recordUsage(subscriptionId: string, apiCalls: number, storageGB: number): Usage {
    const now = new Date();
    const periodStart = getFirstDayOfMonth(now);
    const periodEnd = getLastDayOfMonth(now);

    let usage = store.usages.find(
      (u) =>
        u.subscriptionId === subscriptionId &&
        u.periodStart.getTime() === periodStart.getTime() &&
        !u.billed
    );

    if (!usage) {
      usage = {
        id: store.generateId(),
        subscriptionId,
        periodStart,
        periodEnd,
        apiCalls: 0,
        storageGB: 0,
        billed: false,
      };
      store.usages.push(usage);
    }

    usage.apiCalls += apiCalls;
    if (storageGB > 0) {
      usage.storageGB = Math.max(usage.storageGB, storageGB);
    }

    return usage;
  }

  getCurrentUsage(subscriptionId: string): Usage | null {
    const now = new Date();
    const periodStart = getFirstDayOfMonth(now);

    return (
      store.usages.find(
        (u) =>
          u.subscriptionId === subscriptionId &&
          u.periodStart.getTime() === periodStart.getTime() &&
          !u.billed
      ) || null
    );
  }

  getUsageForPeriod(subscriptionId: string, periodStart: Date, periodEnd: Date): Usage | null {
    return (
      store.usages.find(
        (u) =>
          u.subscriptionId === subscriptionId &&
          u.periodStart.getTime() === periodStart.getTime() &&
          u.periodEnd.getTime() === periodEnd.getTime()
      ) || null
    );
  }

  markAsBilled(usageId: string): void {
    const usage = store.usages.find((u) => u.id === usageId);
    if (usage) {
      usage.billed = true;
    }
  }
}

export const usageService = new UsageService();
