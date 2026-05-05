export const ErrorCodes = {
  INVALID_INPUT: 'INVALID_INPUT',
  SALESPERSON_NOT_FOUND: 'SALESPERSON_NOT_FOUND',
  SALESPERSON_ALREADY_EXISTS: 'SALESPERSON_ALREADY_EXISTS',
  SALESPERSON_ALREADY_RESIGNED: 'SALESPERSON_ALREADY_RESIGNED',
  ORDER_NOT_FOUND: 'ORDER_NOT_FOUND',
  ORDER_ALREADY_REFUNDED: 'ORDER_ALREADY_REFUNDED',
  ORDER_LOCKED: 'ORDER_LOCKED',
  SETTLEMENT_NOT_FOUND: 'SETTLEMENT_NOT_FOUND',
  INVALID_DATE: 'INVALID_DATE',
  INVALID_AMOUNT: 'INVALID_AMOUNT',
  INVALID_SALES_CONTRIBUTION: 'INVALID_SALES_CONTRIBUTION',
  NO_PRIMARY_SALESPERSON: 'NO_PRIMARY_SALESPERSON',
  MULTIPLE_PRIMARY_SALESPERSONS: 'MULTIPLE_PRIMARY_SALESPERSONS',
  INSUFFICIENT_COMMISSION_BALANCE: 'INSUFFICIENT_COMMISSION_BALANCE',
  INVALID_RANKING_DIMENSION: 'INVALID_RANKING_DIMENSION',
  INVALID_QUARTER: 'INVALID_QUARTER',
  INVALID_MONTH: 'INVALID_MONTH',
  INTERNAL_ERROR: 'INTERNAL_ERROR',
} as const;

export type ErrorCode = typeof ErrorCodes[keyof typeof ErrorCodes];

export class BusinessError extends Error {
  readonly code: ErrorCode;
  readonly details?: Record<string, unknown>;

  constructor(code: ErrorCode, message: string, details?: Record<string, unknown>) {
    super(message);
    this.code = code;
    this.details = details;
    this.name = 'BusinessError';
  }
}

export function createError(code: ErrorCode, message: string, details?: Record<string, unknown>): BusinessError {
  return new BusinessError(code, message, details);
}
