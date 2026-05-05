import {
  BillingError,
  ErrorCodes,
  generateId,
  getCurrentDate,
  getDaysInMonth,
  getDaysBetween,
  getFirstDayOfMonth,
  getLastDayOfMonth,
  getMonthsBetween,
  PLANS,
  PlanType,
  YEARLY_DISCOUNT_PERCENT,
  REFUND_ELIGIBLE_MONTH,
  REFUND_FEE_PERCENT,
  applyDiscount,
  calculateDailyRate,
  calculateProratedFee,
  calculateRefundAmount,
} from '@billing/shared';
import type {
  Subscription,
  UpgradeProration,
  CreateSubscriptionRequest,
  UpgradeSubscriptionRequest,
  RequestRefundRequest,
} from '@billing/shared';
import {
  findSubscriptionById,
  findActiveSubscriptionByCustomerId,
  saveSubscription,
  findRefundsBySubscriptionId,
  saveRefund,
} from '../store';
import { getCustomer, ensureCustomerActive } from './customer-service';

export function createSubscription(request: CreateSubscriptionRequest): {
  subscription: Subscription;
  yearlyDiscountApplied: boolean;
} {
  getCustomer(request.customerId);
  ensureCustomerActive(request.customerId);

  const existingSubscription = findActiveSubscriptionByCustomerId(request.customerId);
  if (existingSubscription) {
    throw new BillingError(ErrorCodes.SUBSCRIPTION_ALREADY_ACTIVE, '客户已有活跃订阅', {
      customerId: request.customerId,
      existingSubscriptionId: existingSubscription.id,
    });
  }

  const plan = PLANS[request.planType];
  if (!plan) {
    throw new BillingError(ErrorCodes.INVALID_PLAN_TYPE, '无效的套餐类型', {
      planType: request.planType,
    });
  }

  const now = getCurrentDate();
  let yearlyDiscountApplied = false;
  let yearlyPaymentAmount = 0;

  if (request.billingCycle === 'yearly' && request.planType !== 'free') {
    yearlyPaymentAmount = applyDiscount(plan.yearlyFee, YEARLY_DISCOUNT_PERCENT);
    yearlyDiscountApplied = true;
  }

  const subscription: Subscription = {
    id: generateId(),
    customerId: request.customerId,
    planType: request.planType,
    billingCycle: request.billingCycle,
    startsAt: now,
    endsAt: null,
    isActive: true,
    isYearlyPaid: request.billingCycle === 'yearly' && request.planType !== 'free',
    yearlyPaymentAmount,
    yearlyPaymentDate: request.billingCycle === 'yearly' && request.planType !== 'free' ? now : null,
    createdAt: now,
    updatedAt: now,
  };

  saveSubscription(subscription);
  return { subscription, yearlyDiscountApplied };
}

export function getActiveSubscription(customerId: string): Subscription {
  const subscription = findActiveSubscriptionByCustomerId(customerId);
  if (!subscription) {
    throw new BillingError(ErrorCodes.SUBSCRIPTION_NOT_FOUND, '客户无活跃订阅', {
      customerId,
    });
  }
  return subscription;
}

export function getSubscription(subscriptionId: string): Subscription {
  const subscription = findSubscriptionById(subscriptionId);
  if (!subscription) {
    throw new BillingError(ErrorCodes.SUBSCRIPTION_NOT_FOUND, '订阅不存在', {
      subscriptionId,
    });
  }
  return subscription;
}

export function calculateUpgradeProration(
  fromPlan: PlanType,
  toPlan: PlanType,
  upgradeDate: string
): UpgradeProration {
  const periodStart = getFirstDayOfMonth(upgradeDate);
  const periodEnd = getLastDayOfMonth(upgradeDate);

  const daysInPeriod = getDaysInMonth(upgradeDate);
  const daysOnOldPlan = getDaysBetween(periodStart, upgradeDate);
  const daysOnNewPlan = daysInPeriod - daysOnOldPlan;

  const fromPlanConfig = PLANS[fromPlan];
  const toPlanConfig = PLANS[toPlan];

  const oldPlanDailyRate = calculateDailyRate(fromPlanConfig.monthlyFee, daysInPeriod);
  const newPlanDailyRate = calculateDailyRate(toPlanConfig.monthlyFee, daysInPeriod);

  const oldPlanProratedFee = calculateProratedFee(oldPlanDailyRate, daysOnOldPlan);
  const newPlanProratedFee = calculateProratedFee(newPlanDailyRate, daysOnNewPlan);

  return {
    fromPlan,
    toPlan,
    upgradeDate,
    periodStart,
    periodEnd,
    daysInPeriod,
    daysOnOldPlan,
    daysOnNewPlan,
    oldPlanDailyRate,
    newPlanDailyRate,
    oldPlanProratedFee,
    newPlanProratedFee,
    totalBaseFee: oldPlanProratedFee + newPlanProratedFee,
  };
}

export function upgradeSubscription(
  request: UpgradeSubscriptionRequest
): { subscription: Subscription; proration: UpgradeProration } {
  const oldSubscription = getSubscription(request.subscriptionId);

  if (!oldSubscription.isActive) {
    throw new BillingError(ErrorCodes.SUBSCRIPTION_INACTIVE, '订阅已不活跃', {
      subscriptionId: request.subscriptionId,
    });
  }

  if (oldSubscription.planType === request.toPlanType) {
    throw new BillingError(ErrorCodes.BAD_REQUEST, '不能升级到相同的套餐', {
      subscriptionId: request.subscriptionId,
      currentPlan: oldSubscription.planType,
    });
  }

  const now = getCurrentDate();
  const proration = calculateUpgradeProration(
    oldSubscription.planType,
    request.toPlanType,
    now
  );

  oldSubscription.isActive = false;
  oldSubscription.endsAt = now;
  oldSubscription.updatedAt = now;
  saveSubscription(oldSubscription);

  const newSubscription: Subscription = {
    id: generateId(),
    customerId: request.customerId,
    planType: request.toPlanType,
    billingCycle: oldSubscription.billingCycle,
    startsAt: now,
    endsAt: null,
    isActive: true,
    isYearlyPaid: oldSubscription.isYearlyPaid,
    yearlyPaymentAmount: oldSubscription.yearlyPaymentAmount,
    yearlyPaymentDate: oldSubscription.yearlyPaymentDate,
    createdAt: now,
    updatedAt: now,
  };

  saveSubscription(newSubscription);

  return { subscription: newSubscription, proration };
}

export function isSubscriptionEligibleForRefund(subscription: Subscription): {
  eligible: boolean;
  reason?: string;
  remainingMonths?: number;
} {
  if (!subscription.isActive) {
    return { eligible: false, reason: '订阅已不活跃' };
  }

  if (!subscription.isYearlyPaid || !subscription.yearlyPaymentDate) {
    return { eligible: false, reason: '非年度预付费订阅' };
  }

  const now = getCurrentDate();
  const monthsSincePayment =
    getMonthsBetween(subscription.yearlyPaymentDate, now) + 1;

  if (monthsSincePayment < REFUND_ELIGIBLE_MONTH) {
    return {
      eligible: false,
      reason: `需在第${REFUND_ELIGIBLE_MONTH}个月后才能申请退款`,
    };
  }

  if (monthsSincePayment >= 12) {
    return { eligible: false, reason: '年度订阅已到期' };
  }

  const existingRefunds = findRefundsBySubscriptionId(subscription.id);
  if (existingRefunds.some((r) => r.status === 'approved' || r.status === 'pending')) {
    return { eligible: false, reason: '已有退款申请' };
  }

  const remainingMonths = 12 - monthsSincePayment;
  return { eligible: true, remainingMonths };
}

export function requestRefund(request: RequestRefundRequest): {
  refundId: string;
  amount: number;
  refundedMonths: number;
} {
  const subscription = getSubscription(request.subscriptionId);

  if (subscription.customerId !== request.customerId) {
    throw new BillingError(ErrorCodes.BAD_REQUEST, '订阅不属于该客户', {
      subscriptionId: request.subscriptionId,
      customerId: request.customerId,
    });
  }

  const eligibility = isSubscriptionEligibleForRefund(subscription);
  if (!eligibility.eligible || !eligibility.remainingMonths) {
    throw new BillingError(ErrorCodes.REFUND_NOT_ELIGIBLE, eligibility.reason || '不符合退款条件', {
      subscriptionId: request.subscriptionId,
    });
  }

  const plan = PLANS[subscription.planType];
  const refundCalculation = calculateRefundAmount(
    eligibility.remainingMonths,
    plan.monthlyFee,
    REFUND_FEE_PERCENT
  );

  const now = getCurrentDate();
  const refund = {
    id: generateId(),
    customerId: request.customerId,
    subscriptionId: request.subscriptionId,
    amount: refundCalculation.refundAmount,
    refundedMonths: eligibility.remainingMonths,
    status: 'pending' as const,
    reason: request.reason,
    processedAt: null,
    createdAt: now,
  };

  saveRefund(refund);

  return {
    refundId: refund.id,
    amount: refund.amount,
    refundedMonths: refund.refundedMonths,
  };
}
