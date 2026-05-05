import {
  BillingError,
  ErrorCodes,
  generateId,
  getCurrentDate,
  addDays,
  isDateAfter,
  isDateBefore,
  getMonthsBetween,
  TRIAL_DAYS,
  FREEZE_AFTER_MONTHS_OVERDUE,
} from '@billing/shared';
import type { Customer, CustomerStatus, CreateCustomerRequest } from '@billing/shared';
import {
  findCustomerById,
  findCustomerByEmail,
  saveCustomer,
  findOverdueBillsByCustomerId,
} from '../store';

export function createCustomer(request: CreateCustomerRequest): Customer {
  const existingCustomer = findCustomerByEmail(request.email);
  if (existingCustomer) {
    throw new BillingError(ErrorCodes.CUSTOMER_ALREADY_EXISTS, '客户邮箱已存在', {
      email: request.email,
    });
  }

  const now = getCurrentDate();
  const trialEndAt = addDays(now, TRIAL_DAYS);

  const customer: Customer = {
    id: generateId(),
    name: request.name,
    email: request.email,
    status: 'trial',
    trialEndAt,
    frozenAt: null,
    createdAt: now,
    updatedAt: now,
  };

  saveCustomer(customer);
  return customer;
}

export function getCustomer(customerId: string): Customer {
  const customer = findCustomerById(customerId);
  if (!customer) {
    throw new BillingError(ErrorCodes.CUSTOMER_NOT_FOUND, '客户不存在', {
      customerId,
    });
  }
  return customer;
}

export function checkAndUpdateCustomerStatus(customerId: string): Customer {
  const customer = getCustomer(customerId);
  const now = getCurrentDate();

  if (customer.status === 'trial' && customer.trialEndAt && isDateAfter(now, customer.trialEndAt)) {
    customer.status = 'active';
    customer.trialEndAt = null;
    customer.updatedAt = now;
    saveCustomer(customer);
  }

  if (customer.status !== 'frozen') {
    const overdueBills = findOverdueBillsByCustomerId(customerId);
    if (overdueBills.length >= FREEZE_AFTER_MONTHS_OVERDUE) {
      const oldestBill = overdueBills.reduce((oldest, bill) =>
        getMonthsBetween(bill.dueAt, now) > getMonthsBetween(oldest.dueAt, now) ? bill : oldest
      );
      const monthsOverdue = getMonthsBetween(oldestBill.dueAt, now);
      if (monthsOverdue >= FREEZE_AFTER_MONTHS_OVERDUE) {
        customer.status = 'frozen';
        customer.frozenAt = now;
        customer.updatedAt = now;
        saveCustomer(customer);
      }
    }
  }

  return customer;
}

export function updateCustomerStatus(customerId: string, status: CustomerStatus): Customer {
  const customer = getCustomer(customerId);
  const wasFrozen = customer.status === 'frozen';
  customer.status = status;
  if (wasFrozen && status !== 'frozen') {
    customer.frozenAt = null;
  }
  customer.updatedAt = getCurrentDate();
  saveCustomer(customer);
  return customer;
}

export function isCustomerFrozen(customerId: string): boolean {
  const customer = getCustomer(customerId);
  return customer.status === 'frozen';
}

export function ensureCustomerActive(customerId: string): void {
  const customer = checkAndUpdateCustomerStatus(customerId);
  if (customer.status === 'frozen') {
    throw new BillingError(ErrorCodes.CUSTOMER_FROZEN, '客户已被冻结，请先结清欠款', {
      customerId,
    });
  }
}

export function isCustomerInTrial(customerId: string): boolean {
  const customer = getCustomer(customerId);
  if (customer.status !== 'trial' || !customer.trialEndAt) {
    return false;
  }
  const now = getCurrentDate();
  return isDateBefore(now, customer.trialEndAt);
}
