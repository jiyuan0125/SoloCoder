export const ErrorCodes = {
  CUSTOMER_NOT_FOUND: 'CUSTOMER_NOT_FOUND',
  SUBSCRIPTION_NOT_FOUND: 'SUBSCRIPTION_NOT_FOUND',
  BILL_NOT_FOUND: 'BILL_NOT_FOUND',
  PAYMENT_NOT_FOUND: 'PAYMENT_NOT_FOUND',
  REFUND_NOT_FOUND: 'REFUND_NOT_FOUND',

  INVALID_PLAN_TYPE: 'INVALID_PLAN_TYPE',
  INVALID_BILLING_CYCLE: 'INVALID_BILLING_CYCLE',
  INVALID_USAGE_AMOUNT: 'INVALID_USAGE_AMOUNT',
  INVALID_PAYMENT_AMOUNT: 'INVALID_PAYMENT_AMOUNT',

  CUSTOMER_FROZEN: 'CUSTOMER_FROZEN',
  CUSTOMER_ALREADY_EXISTS: 'CUSTOMER_ALREADY_EXISTS',
  SUBSCRIPTION_ALREADY_ACTIVE: 'SUBSCRIPTION_ALREADY_ACTIVE',
  SUBSCRIPTION_INACTIVE: 'SUBSCRIPTION_INACTIVE',

  USAGE_OVER_QUOTA_LIMITED: 'USAGE_OVER_QUOTA_LIMITED',

  BILL_ALREADY_PAID: 'BILL_ALREADY_PAID',
  BILL_OVERDUE: 'BILL_OVERDUE',
  BILL_REVIEW_WINDOW_EXPIRED: 'BILL_REVIEW_WINDOW_EXPIRED',
  BILL_ALREADY_UNDER_REVIEW: 'BILL_ALREADY_UNDER_REVIEW',

  REFUND_NOT_ELIGIBLE: 'REFUND_NOT_ELIGIBLE',
  REFUND_ALREADY_PROCESSED: 'REFUND_ALREADY_PROCESSED',

  VALIDATION_ERROR: 'VALIDATION_ERROR',
  INTERNAL_ERROR: 'INTERNAL_ERROR',
  BAD_REQUEST: 'BAD_REQUEST',
} as const;

export type ErrorCode = typeof ErrorCodes[keyof typeof ErrorCodes];

export class BillingError extends Error {
  public readonly code: ErrorCode;
  public readonly details?: Record<string, unknown>;

  constructor(code: ErrorCode, message: string, details?: Record<string, unknown>) {
    super(message);
    this.code = code;
    if (details !== undefined) {
      this.details = details;
    }
    this.name = 'BillingError';
  }
}

export function isBillingError(error: unknown): error is BillingError {
  return error instanceof BillingError;
}
