import { Bill, PlanId, Plan } from '../types';
import { store } from '../data/store';
import { planService } from './planService';
import { subscriptionService } from './subscriptionService';
import { usageService } from './UsageService';
import {
  getFirstDayOfMonth,
  getLastDayOfMonth,
  getDaysInMonth,
  addDays,
  getStartOfDay,
} from '../utils/date';

export class BillingService {
  generateMonthlyBills(month?: Date): Bill[] {
    const targetDate = month || new Date();
    const currentMonth = new Date(targetDate.getFullYear(), targetDate.getMonth(), 1);

    const subscriptions = store.subscriptions.filter((s) => s.status !== 'cancelled');
    const bills: Bill[] = [];

    for (const sub of subscriptions) {
      const periodStart = getFirstDayOfMonth(currentMonth);
      const periodEnd = getLastDayOfMonth(currentMonth);

      const existingBill = store.bills.find(
        (b) => b.subscriptionId === sub.id && b.periodStart.getTime() === periodStart.getTime()
      );
      if (existingBill) continue;

      const bill = this.calculateBillForPeriod(sub.id, periodStart, periodEnd);
      if (bill) {
        bills.push(bill);
      }
    }

    return bills;
  }

  private calculateBillForPeriod(
    subscriptionId: string,
    periodStart: Date,
    periodEnd: Date
  ): Bill | null {
    const subscription = store.subscriptions.find((s) => s.id === subscriptionId);
    if (!subscription) return null;

    const baseFeeCents = this.calculateBaseFee(subscription, periodStart, periodEnd);
    const usageFeeCents = this.calculateUsageFee(subscription, periodStart, periodEnd);
    const totalAmountCents = baseFeeCents + usageFeeCents;

    const generatedAt = new Date();
    const dueDate = addDays(generatedAt, 15);

    const bill: Bill = {
      id: store.generateId(),
      subscriptionId,
      periodStart,
      periodEnd,
      baseFeeCents,
      usageFeeCents,
      totalAmountCents,
      status: 'pending',
      generatedAt,
      dueDate,
      paidAt: null,
    };

    store.bills.push(bill);
    return bill;
  }

  private calculateBaseFee(
    subscription: typeof store.subscriptions[0],
    periodStart: Date,
    periodEnd: Date
  ): number {
    const totalDays = getDaysInMonth(periodStart);
    let totalBaseFee = 0;

    const changes = [...subscription.planChangeHistory].sort(
      (a, b) => a.effectiveDate.getTime() - b.effectiveDate.getTime()
    );

    const segments: { planId: PlanId; start: Date; end: Date }[] = [];

    for (let i = 0; i < changes.length; i++) {
      const current = changes[i];
      const next = changes[i + 1];

      const segmentStart = new Date(
        Math.max(current.effectiveDate.getTime(), periodStart.getTime())
      );
      const segmentEnd = next
        ? new Date(Math.min(next.effectiveDate.getTime(), periodEnd.getTime() + 1))
        : new Date(periodEnd.getTime() + 1);

      if (segmentStart < segmentEnd) {
        segments.push({
          planId: current.toPlanId,
          start: getStartOfDay(segmentStart),
          end: getStartOfDay(segmentEnd),
        });
      }
    }

    for (const segment of segments) {
      const plan = planService.getPlanById(segment.planId);
      if (!plan) continue;

      const diffTime = Math.abs(segment.end.getTime() - segment.start.getTime());
      const segmentDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));

      if (segmentDays > 0) {
        const ratio = segmentDays / totalDays;
        const proratedFee = Math.round(plan.priceMonthlyCents * ratio);
        totalBaseFee += proratedFee;
      }
    }

    if (segments.length === 0) {
      const plan = planService.getPlanById(subscription.planId);
      if (plan) {
        totalBaseFee = plan.priceMonthlyCents;
      }
    }

    return totalBaseFee;
  }

  private calculateUsageFee(
    subscription: typeof store.subscriptions[0],
    periodStart: Date,
    periodEnd: Date
  ): number {
    const usage = usageService.getUsageForPeriod(
      subscription.id,
      periodStart,
      periodEnd
    );
    if (!usage) return 0;

    const currentPlan = planService.getPlanById(subscription.planId);
    if (!currentPlan) return 0;

    let usageFee = 0;

    const overageApiCalls = Math.max(0, usage.apiCalls - currentPlan.apiCallsLimit);
    usageFee += overageApiCalls * currentPlan.overageApiCallCostCents;

    const overageStorage = Math.max(0, usage.storageGB - currentPlan.storageLimitGB);
    usageFee += overageStorage * currentPlan.overageStorageCostCentsPerGB;

    return Math.round(usageFee);
  }

  payBill(billId: string): { success: boolean; message?: string } {
    const bill = store.bills.find((b) => b.id === billId);
    if (!bill) {
      return { success: false, message: '账单不存在' };
    }

    if (bill.status === 'paid') {
      return { success: false, message: '账单已支付' };
    }

    bill.status = 'paid';
    bill.paidAt = new Date();
    return { success: true };
  }

  getBillsBySubscription(subscriptionId: string): Bill[] {
    return store.bills.filter((b) => b.subscriptionId === subscriptionId);
  }

  getBillsByCustomer(customerId: string): Bill[] {
    const customerSubs = store.subscriptions.filter((s) => s.customerId === customerId);
    const subIds = new Set(customerSubs.map((s) => s.id));
    return store.bills.filter((b) => subIds.has(b.subscriptionId));
  }

  processOverdueBills(): void {
    const now = new Date();

    for (const bill of store.bills) {
      if (bill.status === 'paid') continue;

      if (bill.dueDate < now) {
        bill.status = 'overdue';

        const subscription = store.subscriptions.find((s) => s.id === bill.subscriptionId);
        if (subscription && subscription.status === 'active') {
          subscriptionService.pauseSubscription(subscription.id);
        }
      }
    }

    this.checkAutoCancel();
  }

  private checkAutoCancel(): void {
    const now = new Date();

    for (const bill of store.bills) {
      if (bill.status === 'paid') continue;

      const pauseDate = addDays(bill.dueDate, 0);
      const cancelDate = addDays(pauseDate, 30);

      if (cancelDate < now) {
        const subscription = store.subscriptions.find((s) => s.id === bill.subscriptionId);
        if (subscription && subscription.status !== 'cancelled') {
          subscriptionService.cancelSubscription(subscription.id);
        }
      }
    }
  }
}

export const billingService = new BillingService();
