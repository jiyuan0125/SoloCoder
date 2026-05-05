export enum ErrorCode {
  BUDGET_NOT_FOUND = 'BUDGET_NOT_FOUND',
  DEPARTMENT_NOT_FOUND = 'DEPARTMENT_NOT_FOUND',
  INSUFFICIENT_BUDGET = 'INSUFFICIENT_BUDGET',
  TOTAL_EXCEEDS_ANNUAL = 'TOTAL_EXCEEDS_ANNUAL',
  BUDGET_LOWER_THAN_USED = 'BUDGET_LOWER_THAN_USED',
  CARRYOVER_EXCEEDS_LIMIT = 'CARRYOVER_EXCEEDS_LIMIT',
  INVALID_QUARTER_TRANSITION = 'INVALID_QUARTER_TRANSITION',
  INVALID_MONTH = 'INVALID_MONTH',
  INVALID_CATEGORY = 'INVALID_CATEGORY',
  INVALID_AMOUNT = 'INVALID_AMOUNT',
  BUDGET_ALREADY_EXISTS = 'BUDGET_ALREADY_EXISTS',
  BUDGET_NOT_ACTIVE = 'BUDGET_NOT_ACTIVE',
  BUDGET_NOT_DRAFT = 'BUDGET_NOT_DRAFT',
  INVALID_REQUEST = 'INVALID_REQUEST',
  INTERNAL_ERROR = 'INTERNAL_ERROR',
}

export interface ErrorResponse {
  success: false;
  error: {
    code: ErrorCode;
    message: string;
    details?: Record<string, unknown>;
  };
}

export interface SuccessResponse<T = unknown> {
  success: true;
  data: T;
}

export type ApiResponse<T = unknown> = SuccessResponse<T> | ErrorResponse;

export function createSuccessResponse<T>(data: T): SuccessResponse<T> {
  return {
    success: true,
    data,
  };
}

export function createErrorResponse(
  code: ErrorCode,
  message: string,
  details?: Record<string, unknown>
): ErrorResponse {
  return {
    success: false,
    error: {
      code,
      message,
      details,
    },
  };
}
