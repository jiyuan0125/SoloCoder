import {
  BillingError,
  ErrorCodes,
  generateId,
  getCurrentDate,
  getFirstDayOfMonth,
  getLastDayOfMonth,
  PLANS,
  MonthlyUsage,
  RecordUsageRequest,
} from '@billing/shared';
import type { UsageRecord } from '@billing/shared';
import {
  findUsageRecordsByCustomerId,
  saveUsageRecord,
} from '../store';
import { isCustomerFrozen, isCustomerInTrial } from './customer-service';
import { getActiveSubscription } from './subscription-service';

export function recordUsage(request: RecordUsageRequest): {
  record: UsageRecord;
  isLimited: boolean;
} {
  if (isCustomerFrozen(request.customerId)) {
    throw new BillingError(ErrorCodes.CUSTOMER_FROZEN, '客户已被冻结，无法使用服务', {
      customerId: request.customerId,
    });
  }

  if (request.units <= 0) {
    throw new BillingError(ErrorCodes.INVALID_USAGE_AMOUNT, '用量必须大于0', {
      units: request.units,
    });
  }

  const subscription = getActiveSubscription(request.customerId);
  const plan = PLANS[subscription.planType];
  const now = request.timestamp || getCurrentDate();
  const periodStart = getFirstDayOfMonth(now);
  const periodEnd = getLastDayOfMonth(now);

  const inTrial = isCustomerInTrial(request.customerId);

  const currentUsage = calculateMonthlyUsage(
    request.customerId,
    subscription.id,
    periodStart,
    periodEnd
  );

  const totalUnitsAfter = currentUsage.totalUnits + request.units;
  const isFreeQuota = inTrial || totalUnitsAfter <= plan.monthlyQuota;

  let isLimited = false;
  let actualUnits = request.units;

  if (plan.type === 'free' && totalUnitsAfter > plan.monthlyQuota) {
    const overage = totalUnitsAfter - plan.monthlyQuota;
    if (overage > 0) {
      actualUnits = plan.monthlyQuota - currentUsage.totalUnits;
      if (actualUnits < 0) actualUnits = 0;
      isLimited = actualUnits === 0;
    }
  }

  if (actualUnits > 0) {
    const record: UsageRecord = {
      id: generateId(),
      customerId: request.customerId,
      units: actualUnits,
      recordedAt: now,
      billingPeriodStart: periodStart,
      billingPeriodEnd: periodEnd,
      isFreeQuota,
    };
    saveUsageRecord(record);
    return { record, isLimited };
  }

  return {
    record: {
      id: generateId(),
      customerId: request.customerId,
      units: 0,
      recordedAt: now,
      billingPeriodStart: periodStart,
      billingPeriodEnd: periodEnd,
      isFreeQuota: true,
    },
    isLimited: true,
  };
}

export function calculateMonthlyUsage(
  customerId: string,
  subscriptionId: string,
  periodStart: string,
  periodEnd: string
): MonthlyUsage {
  const subscription = getActiveSubscription(customerId);
  const plan = PLANS[subscription.planType];

  const records = findUsageRecordsByCustomerId(customerId, periodStart, periodEnd);
  const totalUnits = records.reduce((sum, r) => sum + r.units, 0);

  const quotaUsed = Math.min(totalUnits, plan.monthlyQuota);
  const overageUnits = Math.max(0, totalUnits - plan.monthlyQuota);

  return {
    customerId,
    subscriptionId,
    periodStart,
    periodEnd,
    totalUnits,
    quotaUsed,
    overageUnits,
  };
}

export function getCurrentMonthUsage(customerId: string): MonthlyUsage {
  const subscription = getActiveSubscription(customerId);
  const now = getCurrentDate();
  const periodStart = getFirstDayOfMonth(now);
  const periodEnd = getLastDayOfMonth(now);

  return calculateMonthlyUsage(customerId, subscription.id, periodStart, periodEnd);
}
