import {
  ErrorCode,
  EXPENSE_CATEGORIES,
  type ExpenseStatus,
} from '@expense-report/shared';
import { isValidDate, isValidAmount, isValidCategory, isValidStatus } from './utils';

interface ValidationResult {
  valid: boolean;
  errorCode?: ErrorCode;
  errorMessage?: string;
}

export function validateCreateExpenseParams(params: {
  employeeId: unknown;
  employeeName: unknown;
  date: unknown;
  amount: unknown;
  category: unknown;
  reason: unknown;
  voucherFileName: unknown;
}): ValidationResult {
  if (typeof params.employeeId !== 'string' || params.employeeId.trim() === '') {
    return { valid: false, errorCode: ErrorCode.MISSING_PARAMETER, errorMessage: '缺少 employeeId 或参数无效' };
  }
  if (typeof params.employeeName !== 'string' || params.employeeName.trim() === '') {
    return { valid: false, errorCode: ErrorCode.MISSING_PARAMETER, errorMessage: '缺少 employeeName 或参数无效' };
  }
  if (typeof params.date !== 'string' || !isValidDate(params.date)) {
    return { valid: false, errorCode: ErrorCode.INVALID_DATE, errorMessage: '日期格式无效' };
  }
  if (!isValidAmount(params.amount)) {
    return { valid: false, errorCode: ErrorCode.INVALID_AMOUNT, errorMessage: '金额无效，必须是非负整数（单位：分）' };
  }
  if (typeof params.category !== 'string' || !isValidCategory(params.category)) {
    return { valid: false, errorCode: ErrorCode.INVALID_CATEGORY, errorMessage: `费用类别无效，有效值：${EXPENSE_CATEGORIES.join(', ')}` };
  }
  if (typeof params.reason !== 'string' || params.reason.trim() === '') {
    return { valid: false, errorCode: ErrorCode.MISSING_PARAMETER, errorMessage: '缺少事由或参数无效' };
  }
  if (typeof params.voucherFileName !== 'string' || params.voucherFileName.trim() === '') {
    return { valid: false, errorCode: ErrorCode.MISSING_PARAMETER, errorMessage: '缺少凭证文件名或参数无效' };
  }
  return { valid: true };
}

export function validateUpdateExpenseParams(params: {
  date?: unknown;
  amount?: unknown;
  category?: unknown;
  reason?: unknown;
  voucherFileName?: unknown;
  supplementaryInfo?: unknown;
}): ValidationResult {
  if (params.date !== undefined) {
    if (typeof params.date !== 'string' || !isValidDate(params.date)) {
      return { valid: false, errorCode: ErrorCode.INVALID_DATE, errorMessage: '日期格式无效' };
    }
  }
  if (params.amount !== undefined) {
    if (!isValidAmount(params.amount)) {
      return { valid: false, errorCode: ErrorCode.INVALID_AMOUNT, errorMessage: '金额无效，必须是非负整数（单位：分）' };
    }
  }
  if (params.category !== undefined) {
    if (typeof params.category !== 'string' || !isValidCategory(params.category)) {
      return { valid: false, errorCode: ErrorCode.INVALID_CATEGORY, errorMessage: `费用类别无效，有效值：${EXPENSE_CATEGORIES.join(', ')}` };
    }
  }
  if (params.reason !== undefined) {
    if (typeof params.reason !== 'string' || params.reason.trim() === '') {
      return { valid: false, errorCode: ErrorCode.MISSING_PARAMETER, errorMessage: '事由不能为空' };
    }
  }
  if (params.voucherFileName !== undefined) {
    if (typeof params.voucherFileName !== 'string' || params.voucherFileName.trim() === '') {
      return { valid: false, errorCode: ErrorCode.MISSING_PARAMETER, errorMessage: '凭证文件名不能为空' };
    }
  }
  if (params.supplementaryInfo !== undefined) {
    if (typeof params.supplementaryInfo !== 'string') {
      return { valid: false, errorCode: ErrorCode.INVALID_PARAMETER, errorMessage: '补充说明参数无效' };
    }
  }
  return { valid: true };
}

export function validateApproveParams(params: {
  expenseId: unknown;
  approverId: unknown;
  approverName: unknown;
}): ValidationResult {
  if (typeof params.expenseId !== 'string' || params.expenseId.trim() === '') {
    return { valid: false, errorCode: ErrorCode.MISSING_PARAMETER, errorMessage: '缺少 expenseId 或参数无效' };
  }
  if (typeof params.approverId !== 'string' || params.approverId.trim() === '') {
    return { valid: false, errorCode: ErrorCode.MISSING_PARAMETER, errorMessage: '缺少 approverId 或参数无效' };
  }
  if (typeof params.approverName !== 'string' || params.approverName.trim() === '') {
    return { valid: false, errorCode: ErrorCode.MISSING_PARAMETER, errorMessage: '缺少 approverName 或参数无效' };
  }
  return { valid: true };
}

export function validateBatchApproveParams(params: {
  expenseIds: unknown;
  approverId: unknown;
  approverName: unknown;
}): ValidationResult {
  if (!Array.isArray(params.expenseIds) || params.expenseIds.length === 0) {
    return { valid: false, errorCode: ErrorCode.MISSING_PARAMETER, errorMessage: 'expenseIds 必须是非空数组' };
  }
  for (const id of params.expenseIds) {
    if (typeof id !== 'string' || id.trim() === '') {
      return { valid: false, errorCode: ErrorCode.INVALID_PARAMETER, errorMessage: 'expenseIds 中的每个元素必须是有效字符串' };
    }
  }
  if (typeof params.approverId !== 'string' || params.approverId.trim() === '') {
    return { valid: false, errorCode: ErrorCode.MISSING_PARAMETER, errorMessage: '缺少 approverId 或参数无效' };
  }
  if (typeof params.approverName !== 'string' || params.approverName.trim() === '') {
    return { valid: false, errorCode: ErrorCode.MISSING_PARAMETER, errorMessage: '缺少 approverName 或参数无效' };
  }
  return { valid: true };
}

export function validateListParams(params: {
  startDate?: unknown;
  endDate?: unknown;
  status?: unknown;
  page?: unknown;
  pageSize?: unknown;
}): ValidationResult {
  if (params.startDate !== undefined) {
    if (typeof params.startDate !== 'string' || !isValidDate(params.startDate)) {
      return { valid: false, errorCode: ErrorCode.INVALID_DATE, errorMessage: 'startDate 格式无效' };
    }
  }
  if (params.endDate !== undefined) {
    if (typeof params.endDate !== 'string' || !isValidDate(params.endDate)) {
      return { valid: false, errorCode: ErrorCode.INVALID_DATE, errorMessage: 'endDate 格式无效' };
    }
  }
  if (params.status !== undefined) {
    if (typeof params.status !== 'string' || !isValidStatus(params.status)) {
      return { valid: false, errorCode: ErrorCode.INVALID_STATUS, errorMessage: 'status 参数无效' };
    }
  }
  if (params.page !== undefined) {
    if (typeof params.page !== 'number' || !Number.isInteger(params.page) || params.page < 1) {
      return { valid: false, errorCode: ErrorCode.INVALID_PARAMETER, errorMessage: 'page 必须是大于等于1的整数' };
    }
  }
  if (params.pageSize !== undefined) {
    if (typeof params.pageSize !== 'number' || !Number.isInteger(params.pageSize) || params.pageSize < 1 || params.pageSize > 100) {
      return { valid: false, errorCode: ErrorCode.INVALID_PARAMETER, errorMessage: 'pageSize 必须是1-100之间的整数' };
    }
  }
  return { valid: true };
}

export function validateExpenseStatusForSubmission(status: ExpenseStatus): ValidationResult {
  if (status !== 'pending') {
    return { valid: false, errorCode: ErrorCode.EXPENSE_ALREADY_SUBMITTED, errorMessage: '报销单已提交或不在可提交状态' };
  }
  return { valid: true };
}

export function validateExpenseStatusForApproval(status: ExpenseStatus): ValidationResult {
  if (status !== 'pending') {
    return { valid: false, errorCode: ErrorCode.EXPENSE_NOT_PENDING, errorMessage: '报销单不在待审批状态' };
  }
  return { valid: true };
}

export function validateSupplementaryInfo(
  needsSupplementary: boolean,
  supplementaryInfo?: string
): ValidationResult {
  if (needsSupplementary && (!supplementaryInfo || supplementaryInfo.trim() === '')) {
    return { valid: false, errorCode: ErrorCode.EXPENSE_NEEDS_SUPPLEMENT, errorMessage: '超出限额，需要补充说明' };
  }
  return { valid: true };
}

export function validateUserId(userId: unknown): ValidationResult {
  if (typeof userId !== 'string' || userId.trim() === '') {
    return { valid: false, errorCode: ErrorCode.MISSING_PARAMETER, errorMessage: '缺少 userId 或参数无效' };
  }
  return { valid: true };
}

export function validateExpenseId(expenseId: unknown): ValidationResult {
  if (typeof expenseId !== 'string' || expenseId.trim() === '') {
    return { valid: false, errorCode: ErrorCode.MISSING_PARAMETER, errorMessage: '缺少 expenseId 或参数无效' };
  }
  return { valid: true };
}
