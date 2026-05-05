import {
  BillingError,
  ErrorCodes,
  generateId,
  getCurrentDate,
  PLANS,
  Payment,
} from '@billing/shared';
import type { PayBillRequest } from '@billing/shared';
import {
  savePayment,
  findPaymentsByBillId,
  findOverdueBillsByCustomerId,
  findUsageRecordsByCustomerId,
  saveBillItem,
  findPaymentById,
} from '../store';
import { getCustomer, updateCustomerStatus } from './customer-service';
import { getBill, updateBillStatus, generateBill, getBillItems } from './billing-service';
import { getActiveSubscription } from './subscription-service';

export function payBill(request: PayBillRequest): Payment {
  const bill = getBill(request.billId);

  if (bill.status === 'paid') {
    throw new BillingError(ErrorCodes.BILL_ALREADY_PAID, '账单已支付', {
      billId: request.billId,
    });
  }

  if (request.amount <= 0) {
    throw new BillingError(ErrorCodes.INVALID_PAYMENT_AMOUNT, '支付金额必须大于0', {
      amount: request.amount,
    });
  }

  const now = getCurrentDate();
  const payment: Payment = {
    id: generateId(),
    customerId: bill.customerId,
    billId: bill.id,
    amount: request.amount,
    status: 'completed',
    transactionId: generateId(),
    paidAt: now,
    createdAt: now,
  };

  savePayment(payment);

  const existingPayments = findPaymentsByBillId(bill.id);
  const totalPaid = existingPayments.reduce((sum, p) => sum + p.amount, 0);

  if (totalPaid >= bill.totalAmount) {
    bill.status = 'paid';
    bill.paidAt = now;
    bill.updatedAt = now;
    updateBillStatus(bill.id, 'paid');

    const customer = getCustomer(bill.customerId);
    if (customer.status === 'frozen') {
      const remainingOverdue = findOverdueBillsByCustomerId(bill.customerId).filter(
        (b) => b.id !== bill.id && (b.status === 'pending' || b.status === 'overdue')
      );

      if (remainingOverdue.length === 0) {
        updateCustomerStatus(bill.customerId, 'active');
        handleFrozenPeriodUsage(bill.customerId);
      }
    }
  }

  return payment;
}

function handleFrozenPeriodUsage(customerId: string): void {
  const subscription = getActiveSubscription(customerId);
  const plan = PLANS[subscription.planType];

  if (plan.type === 'free') {
    return;
  }

  const customer = getCustomer(customerId);
  const now = getCurrentDate();

  const recentUsage = findUsageRecordsByCustomerId(
    customerId,
    customer.updatedAt,
    now
  );

  if (recentUsage.length === 0) {
    return;
  }

  const totalUsage = recentUsage.reduce((sum, r) => sum + r.units, 0);

  if (totalUsage <= plan.monthlyQuota) {
    return;
  }

  const overageUnits = totalUsage - plan.monthlyQuota;
  const overageFee = overageUnits * plan.overageFeePerUnit;

  const nextBill = generateBill({
    customerId,
    periodStart: now,
    periodEnd: now,
  });

  const existingItems = getBillItems(nextBill.id);
  const existingFrozenItem = existingItems.find((i) => i.type === 'frozen_usage');

  if (!existingFrozenItem) {
    saveBillItem({
      id: generateId(),
      billId: nextBill.id,
      type: 'frozen_usage',
      description: '冻结期间超额使用费用',
      units: overageUnits,
      unitPrice: plan.overageFeePerUnit,
      amount: overageFee,
    });

    nextBill.overageFee += overageFee;
    nextBill.totalAmount += overageFee;
  }
}

export function getPayment(paymentId: string): Payment {
  const payment = findPaymentsByBillId(paymentId)[0];
  if (!payment) {
    throw new BillingError(ErrorCodes.PAYMENT_NOT_FOUND, '支付记录不存在', {
      paymentId,
    });
  }
  return payment;
}

export function getPaymentById(paymentId: string): Payment {
  const payment = findPaymentById(paymentId);
  if (!payment) {
    throw new BillingError(ErrorCodes.PAYMENT_NOT_FOUND, '支付记录不存在', {
      paymentId,
    });
  }
  return payment;
}
