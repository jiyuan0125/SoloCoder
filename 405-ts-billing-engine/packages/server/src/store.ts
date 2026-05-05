import type {
  Customer,
  Subscription,
  UsageRecord,
  Bill,
  BillItem,
  Payment,
  Refund,
} from '@billing/shared';

export interface DataStore {
  customers: Map<string, Customer>;
  subscriptions: Map<string, Subscription>;
  usageRecords: Map<string, UsageRecord>;
  bills: Map<string, Bill>;
  billItems: Map<string, BillItem>;
  payments: Map<string, Payment>;
  refunds: Map<string, Refund>;
}

export const store: DataStore = {
  customers: new Map(),
  subscriptions: new Map(),
  usageRecords: new Map(),
  bills: new Map(),
  billItems: new Map(),
  payments: new Map(),
  refunds: new Map(),
};

export function findCustomerById(id: string): Customer | undefined {
  return store.customers.get(id);
}

export function findCustomerByEmail(email: string): Customer | undefined {
  for (const customer of store.customers.values()) {
    if (customer.email === email) {
      return customer;
    }
  }
  return undefined;
}

export function saveCustomer(customer: Customer): void {
  store.customers.set(customer.id, customer);
}

export function findSubscriptionById(id: string): Subscription | undefined {
  return store.subscriptions.get(id);
}

export function findActiveSubscriptionByCustomerId(customerId: string): Subscription | undefined {
  for (const subscription of store.subscriptions.values()) {
    if (subscription.customerId === customerId && subscription.isActive) {
      return subscription;
    }
  }
  return undefined;
}

export function findSubscriptionsByCustomerId(customerId: string): Subscription[] {
  const results: Subscription[] = [];
  for (const subscription of store.subscriptions.values()) {
    if (subscription.customerId === customerId) {
      results.push(subscription);
    }
  }
  return results;
}

export function saveSubscription(subscription: Subscription): void {
  store.subscriptions.set(subscription.id, subscription);
}

export function findUsageRecordsByCustomerId(
  customerId: string,
  periodStart?: string,
  periodEnd?: string
): UsageRecord[] {
  const results: UsageRecord[] = [];
  for (const record of store.usageRecords.values()) {
    if (record.customerId !== customerId) {
      continue;
    }
    if (periodStart && periodEnd) {
      if (record.recordedAt >= periodStart && record.recordedAt <= periodEnd) {
        results.push(record);
      }
    } else {
      results.push(record);
    }
  }
  return results.sort((a, b) => a.recordedAt.localeCompare(b.recordedAt));
}

export function saveUsageRecord(record: UsageRecord): void {
  store.usageRecords.set(record.id, record);
}

export function findBillById(id: string): Bill | undefined {
  return store.bills.get(id);
}

export function findBillsByCustomerId(customerId: string, status?: string): Bill[] {
  const results: Bill[] = [];
  for (const bill of store.bills.values()) {
    if (bill.customerId !== customerId) {
      continue;
    }
    if (status && bill.status !== status) {
      continue;
    }
    results.push(bill);
  }
  return results.sort((a, b) => b.createdAt.localeCompare(a.createdAt));
}

export function findOverdueBillsByCustomerId(customerId: string): Bill[] {
  const now = new Date().toISOString();
  const results: Bill[] = [];
  for (const bill of store.bills.values()) {
    if (
      bill.customerId === customerId &&
      (bill.status === 'pending' || bill.status === 'overdue') &&
      bill.dueAt < now
    ) {
      results.push(bill);
    }
  }
  return results;
}

export function saveBill(bill: Bill): void {
  store.bills.set(bill.id, bill);
}

export function findBillItemsByBillId(billId: string): BillItem[] {
  const results: BillItem[] = [];
  for (const item of store.billItems.values()) {
    if (item.billId === billId) {
      results.push(item);
    }
  }
  return results;
}

export function saveBillItem(item: BillItem): void {
  store.billItems.set(item.id, item);
}

export function findPaymentById(id: string): Payment | undefined {
  return store.payments.get(id);
}

export function findPaymentsByBillId(billId: string): Payment[] {
  const results: Payment[] = [];
  for (const payment of store.payments.values()) {
    if (payment.billId === billId) {
      results.push(payment);
    }
  }
  return results;
}

export function findPaymentsByCustomerId(customerId: string): Payment[] {
  const results: Payment[] = [];
  for (const payment of store.payments.values()) {
    if (payment.customerId === customerId) {
      results.push(payment);
    }
  }
  return results;
}

export function savePayment(payment: Payment): void {
  store.payments.set(payment.id, payment);
}

export function findRefundById(id: string): Refund | undefined {
  return store.refunds.get(id);
}

export function findRefundsBySubscriptionId(subscriptionId: string): Refund[] {
  const results: Refund[] = [];
  for (const refund of store.refunds.values()) {
    if (refund.subscriptionId === subscriptionId) {
      results.push(refund);
    }
  }
  return results;
}

export function saveRefund(refund: Refund): void {
  store.refunds.set(refund.id, refund);
}
