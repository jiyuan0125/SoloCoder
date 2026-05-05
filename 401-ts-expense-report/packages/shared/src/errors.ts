export enum ErrorCode {
  SUCCESS = 0,
  UNKNOWN_ERROR = 1000,
  INVALID_PARAMETER = 1001,
  MISSING_PARAMETER = 1002,
  INVALID_AMOUNT = 1003,
  INVALID_CATEGORY = 1004,
  INVALID_STATUS = 1005,
  INVALID_DATE = 1006,

  EXPENSE_NOT_FOUND = 2001,
  EXPENSE_ALREADY_SUBMITTED = 2002,
  EXPENSE_NOT_PENDING = 2003,
  EXPENSE_NEEDS_SUPPLEMENT = 2004,
  EXPENSE_LIMIT_EXCEEDED = 2005,
  EXPENSE_LOCKED = 2006,

  APPROVAL_INSUFFICIENT_LEVEL = 3001,
  APPROVAL_ALREADY_COMPLETED = 3002,
  APPROVAL_NEEDS_SECOND_LEVEL = 3003,
  APPROVAL_LARGE_AMOUNT_REQUIRES_INDIVIDUAL = 3004,

  UNAUTHORIZED = 4001,
  FORBIDDEN = 4002,
  INSUFFICIENT_PERMISSION = 4003,

  INTERNAL_ERROR = 5001,
}

export interface ErrorResponse {
  code: ErrorCode;
  message: string;
  details?: string;
}

export const ERROR_MESSAGES: Readonly<Record<ErrorCode, string>> = {
  [ErrorCode.SUCCESS]: '操作成功',
  [ErrorCode.UNKNOWN_ERROR]: '未知错误',
  [ErrorCode.INVALID_PARAMETER]: '参数无效',
  [ErrorCode.MISSING_PARAMETER]: '缺少必要参数',
  [ErrorCode.INVALID_AMOUNT]: '金额无效',
  [ErrorCode.INVALID_CATEGORY]: '费用类别无效',
  [ErrorCode.INVALID_STATUS]: '状态无效',
  [ErrorCode.INVALID_DATE]: '日期格式无效',
  [ErrorCode.EXPENSE_NOT_FOUND]: '报销单不存在',
  [ErrorCode.EXPENSE_ALREADY_SUBMITTED]: '报销单已提交',
  [ErrorCode.EXPENSE_NOT_PENDING]: '报销单不在待审批状态',
  [ErrorCode.EXPENSE_NEEDS_SUPPLEMENT]: '需要补充说明',
  [ErrorCode.EXPENSE_LIMIT_EXCEEDED]: '超出限额',
  [ErrorCode.EXPENSE_LOCKED]: '报销单已锁定，无法修改',
  [ErrorCode.APPROVAL_INSUFFICIENT_LEVEL]: '审批级别不足',
  [ErrorCode.APPROVAL_ALREADY_COMPLETED]: '审批已完成',
  [ErrorCode.APPROVAL_NEEDS_SECOND_LEVEL]: '需要二级审批',
  [ErrorCode.APPROVAL_LARGE_AMOUNT_REQUIRES_INDIVIDUAL]: '大额报销需单独审批',
  [ErrorCode.UNAUTHORIZED]: '未授权',
  [ErrorCode.FORBIDDEN]: '禁止访问',
  [ErrorCode.INSUFFICIENT_PERMISSION]: '权限不足',
  [ErrorCode.INTERNAL_ERROR]: '内部错误',
};

export function createError(code: ErrorCode, details?: string): ErrorResponse {
  const result: ErrorResponse = {
    code,
    message: ERROR_MESSAGES[code] || ERROR_MESSAGES[ErrorCode.UNKNOWN_ERROR],
  };
  if (details !== undefined) {
    result.details = details;
  }
  return result;
}
