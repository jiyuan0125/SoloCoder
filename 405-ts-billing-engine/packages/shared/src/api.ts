import type { PlanType, BillingCycle, BillStatus, CustomerStatus, PaymentStatus, RefundStatus } from './types';

export interface CreateCustomerRequest {
  name: string;
  email: string;
}

export interface CreateCustomerResponse {
  id: string;
  name: string;
  email: string;
  status: CustomerStatus;
  trialEndAt: string | null;
  createdAt: string;
}

export interface GetCustomerResponse {
  id: string;
  name: string;
  email: string;
  status: CustomerStatus;
  trialEndAt: string | null;
  subscription?: {
    id: string;
    planType: PlanType;
    billingCycle: BillingCycle;
    startsAt: string;
    isActive: boolean;
  };
  createdAt: string;
  updatedAt: string;
}

export interface CreateSubscriptionRequest {
  customerId: string;
  planType: PlanType;
  billingCycle: BillingCycle;
}

export interface CreateSubscriptionResponse {
  id: string;
  customerId: string;
  planType: PlanType;
  billingCycle: BillingCycle;
  startsAt: string;
  isActive: boolean;
  yearlyDiscountApplied: boolean;
  createdAt: string;
}

export interface UpgradeSubscriptionRequest {
  customerId: string;
  subscriptionId: string;
  toPlanType: PlanType;
}

export interface UpgradeSubscriptionResponse {
  subscriptionId: string;
  customerId: string;
  fromPlan: PlanType;
  toPlan: PlanType;
  upgradeDate: string;
  proration: {
    daysOnOldPlan: number;
    daysOnNewPlan: number;
    oldPlanProratedFee: number;
    newPlanProratedFee: number;
    totalBaseFee: number;
  };
}

export interface RecordUsageRequest {
  customerId: string;
  units: number;
  timestamp?: string;
}

export interface RecordUsageResponse {
  id: string;
  customerId: string;
  units: number;
  recordedAt: string;
  isLimited: boolean;
}

export interface GetUsageRequest {
  customerId: string;
  periodStart?: string;
  periodEnd?: string;
}

export interface GetUsageResponse {
  customerId: string;
  subscriptionId: string;
  planType: PlanType;
  monthlyQuota: number;
  periodStart: string;
  periodEnd: string;
  totalUnits: number;
  quotaUsed: number;
  overageUnits: number;
  overageFee: number;
}

export interface GenerateBillRequest {
  customerId: string;
  periodStart: string;
  periodEnd: string;
}

export interface GenerateBillResponse {
  id: string;
  customerId: string;
  subscriptionId: string;
  periodStart: string;
  periodEnd: string;
  baseFee: number;
  overageFee: number;
  discount: number;
  totalAmount: number;
  status: BillStatus;
  dueAt: string;
  reviewDeadline: string;
  items: Array<{
    type: string;
    description: string;
    amount: number;
  }>;
}

export interface GetBillsRequest {
  customerId?: string;
  status?: BillStatus;
  limit?: number;
  offset?: number;
}

export interface GetBillsResponse {
  bills: Array<{
    id: string;
    customerId: string;
    periodStart: string;
    periodEnd: string;
    totalAmount: number;
    status: BillStatus;
    dueAt: string;
    createdAt: string;
  }>;
  total: number;
  limit: number;
  offset: number;
}

export interface GetBillDetailResponse {
  id: string;
  customerId: string;
  subscriptionId: string;
  periodStart: string;
  periodEnd: string;
  baseFee: number;
  overageFee: number;
  discount: number;
  totalAmount: number;
  status: BillStatus;
  dueAt: string;
  paidAt: string | null;
  reviewRequestedAt: string | null;
  reviewDeadline: string;
  notes: string | null;
  items: Array<{
    id: string;
    type: string;
    description: string;
    units: number;
    unitPrice: number;
    amount: number;
  }>;
  createdAt: string;
  updatedAt: string;
}

export interface PayBillRequest {
  billId: string;
  amount: number;
}

export interface PayBillResponse {
  paymentId: string;
  billId: string;
  amount: number;
  status: PaymentStatus;
  paidAt: string | null;
}

export interface RequestReviewRequest {
  billId: string;
  reason: string;
}

export interface RequestReviewResponse {
  billId: string;
  status: BillStatus;
  reviewRequestedAt: string;
  reviewDeadline: string;
}

export interface RequestRefundRequest {
  customerId: string;
  subscriptionId: string;
  reason: string;
}

export interface RequestRefundResponse {
  refundId: string;
  customerId: string;
  subscriptionId: string;
  amount: number;
  refundedMonths: number;
  status: RefundStatus;
  createdAt: string;
}

export interface CheckCustomerStatusRequest {
  customerId: string;
}

export interface CheckCustomerStatusResponse {
  customerId: string;
  status: CustomerStatus;
  isFrozen: boolean;
  freezeReason?: string;
  overdueBills: Array<{
    billId: string;
    amount: number;
    dueAt: string;
  }>;
}

export interface ApiSuccessResponse<T> {
  success: true;
  data: T;
}

export interface ApiErrorResponse {
  success: false;
  error: {
    code: string;
    message: string;
    details?: Record<string, unknown>;
  };
}

export type ApiResponse<T> = ApiSuccessResponse<T> | ApiErrorResponse;
