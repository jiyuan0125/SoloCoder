import {
  BillingError,
  ErrorCodes,
  generateId,
  getCurrentDate,
  addDays,
  getFirstDayOfMonth,
  getLastDayOfMonth,
  PLANS,
  Bill,
  BillItem,
  BillStatus,
  BILL_DUE_DAYS,
  REVIEW_WINDOW_DAYS,
  YEARLY_DISCOUNT_PERCENT,
  applyDiscount,
  calculateOverageFee,
} from '@billing/shared';
import type { GenerateBillRequest } from '@billing/shared';
import {
  findBillById,
  findBillsByCustomerId,
  saveBill,
  saveBillItem,
  findBillItemsByBillId,
  findUsageRecordsByCustomerId,
} from '../store';
import { getActiveSubscription } from './subscription-service';
import { getCustomer } from './customer-service';

export function generateBill(request: GenerateBillRequest): Bill {
  const subscription = getActiveSubscription(request.customerId);
  const plan = PLANS[subscription.planType];

  const now = getCurrentDate();
  const dueAt = addDays(now, BILL_DUE_DAYS);
  const reviewDeadline = addDays(now, REVIEW_WINDOW_DAYS);

  let baseFee = plan.monthlyFee;

  if (subscription.isYearlyPaid && subscription.yearlyPaymentDate) {
    const yearlyDiscountedFee = applyDiscount(plan.yearlyFee, YEARLY_DISCOUNT_PERCENT);
    baseFee = Math.floor(yearlyDiscountedFee / 12);
  }

  const usageRecords = findUsageRecordsByCustomerId(
    request.customerId,
    request.periodStart,
    request.periodEnd
  );
  const totalUsage = usageRecords.reduce((sum, r) => sum + r.units, 0);

  let overageFee = 0;
  if (plan.type !== 'free' && totalUsage > plan.monthlyQuota) {
    const overageUnits = totalUsage - plan.monthlyQuota;
    overageFee = calculateOverageFee(overageUnits, plan.overageFeePerUnit);
  }

  let discount = 0;
  if (subscription.isYearlyPaid && plan.type !== 'free') {
    const monthlyFullPrice = plan.monthlyFee;
    discount = monthlyFullPrice - baseFee;
  }

  const totalAmount = baseFee + overageFee - discount;

  const bill: Bill = {
    id: generateId(),
    customerId: request.customerId,
    subscriptionId: subscription.id,
    periodStart: request.periodStart,
    periodEnd: request.periodEnd,
    baseFee,
    overageFee,
    discount,
    totalAmount: Math.max(0, totalAmount),
    status: 'pending',
    dueAt,
    paidAt: null,
    reviewRequestedAt: null,
    reviewDeadline,
    notes: null,
    createdAt: now,
    updatedAt: now,
  };

  saveBill(bill);

  if (baseFee > 0) {
    const baseFeeItem: BillItem = {
      id: generateId(),
      billId: bill.id,
      type: 'base_fee',
      description: `${plan.name}套餐月费`,
      units: 1,
      unitPrice: baseFee,
      amount: baseFee,
    };
    saveBillItem(baseFeeItem);
  }

  if (overageFee > 0) {
    const overageUnits = Math.max(0, totalUsage - plan.monthlyQuota);
    const overageItem: BillItem = {
      id: generateId(),
      billId: bill.id,
      type: 'overage',
      description: '超额使用费用',
      units: overageUnits,
      unitPrice: plan.overageFeePerUnit,
      amount: overageFee,
    };
    saveBillItem(overageItem);
  }

  if (discount > 0) {
    const discountItem: BillItem = {
      id: generateId(),
      billId: bill.id,
      type: 'discount',
      description: '年度预付费优惠',
      units: 1,
      unitPrice: -discount,
      amount: -discount,
    };
    saveBillItem(discountItem);
  }

  return bill;
}

export function generateMonthlyBill(customerId: string): Bill {
  const now = getCurrentDate();
  const periodStart = getFirstDayOfMonth(now);
  const periodEnd = getLastDayOfMonth(now);

  return generateBill({
    customerId,
    periodStart,
    periodEnd,
  });
}

export function getBill(billId: string): Bill {
  const bill = findBillById(billId);
  if (!bill) {
    throw new BillingError(ErrorCodes.BILL_NOT_FOUND, '账单不存在', {
      billId,
    });
  }
  return bill;
}

export function getCustomerBills(customerId: string, status?: string): Bill[] {
  getCustomer(customerId);
  return findBillsByCustomerId(customerId, status);
}

export function getBillItems(billId: string): BillItem[] {
  return findBillItemsByBillId(billId);
}

export function updateBillStatus(billId: string, status: BillStatus): Bill {
  const bill = getBill(billId);
  bill.status = status;
  bill.updatedAt = getCurrentDate();
  saveBill(bill);
  return bill;
}

export function requestBillReview(billId: string, reason: string): Bill {
  const bill = getBill(billId);
  const now = getCurrentDate();

  if (bill.reviewDeadline && now > bill.reviewDeadline) {
    throw new BillingError(
      ErrorCodes.BILL_REVIEW_WINDOW_EXPIRED,
      '账单复核期已过',
      { billId, reviewDeadline: bill.reviewDeadline }
    );
  }

  if (bill.reviewRequestedAt) {
    throw new BillingError(
      ErrorCodes.BILL_ALREADY_UNDER_REVIEW,
      '账单已在复核中',
      { billId }
    );
  }

  bill.status = 'reviewing';
  bill.reviewRequestedAt = now;
  bill.notes = reason;
  bill.updatedAt = now;
  saveBill(bill);

  return bill;
}
