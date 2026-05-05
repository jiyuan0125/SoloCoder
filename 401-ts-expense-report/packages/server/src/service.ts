import {
  ErrorCode,
  DEFAULT_PAGE,
  DEFAULT_PAGE_SIZE,
  type ExpenseItem,
  type ExpenseStatus,
  type ExpenseCategory,
  type UserRole,
  type AuditActionType,
  type OperationLog,
  type AuditLog,
  type ApprovalRecord,
  type User,
} from '@expense-report/shared';
import {
  generateId,
  getCurrentTime,
  isLargeAmount,
  checkExpenseLimit,
  isDateInRange,
} from './utils';
import {
  getExpenseById,
  getAllExpenses,
  saveExpense,
  deleteExpense,
  addApprovalRecord,
  clearApprovalRecords,
  addOperationLog,
  addAuditLog,
  getUserById,
  getAllOperationLogs,
  getAllAuditLogs,
} from './store';
import {
  validateCreateExpenseParams,
  validateUpdateExpenseParams,
  validateApproveParams,
  validateBatchApproveParams,
  validateListParams,
  validateExpenseStatusForApproval,
  validateSupplementaryInfo,
  validateUserId,
  validateExpenseId,
} from './validators';

interface ServiceResult<T> {
  success: boolean;
  data?: T;
  error?: {
    code: number;
    message: string;
  };
}

interface BatchApproveResult {
  approved: string[];
  skipped: string[];
  failed: Array<{
    expenseId: string;
    code: number;
    message: string;
  }>;
}

function recordOperation(
  userId: string,
  userName: string,
  action: string,
  targetType: 'expense' | 'user' | 'system',
  targetId?: string,
  details?: Record<string, unknown>
): void {
  const log: OperationLog = {
    id: generateId(),
    userId,
    userName,
    action,
    targetType,
    createdAt: getCurrentTime(),
  };
  if (targetId !== undefined) {
    log.targetId = targetId;
  }
  if (details !== undefined) {
    log.details = details;
  }
  addOperationLog(log);
}

function recordAudit(
  userId: string,
  userName: string,
  action: AuditActionType,
  targetType: 'expense' | 'audit_log' | 'system',
  targetId?: string,
  oldValue?: Record<string, unknown>,
  newValue?: Record<string, unknown>,
  ipAddress?: string
): void {
  const log: AuditLog = {
    id: generateId(),
    userId,
    userName,
    action,
    targetType,
    createdAt: getCurrentTime(),
  };
  if (targetId !== undefined) {
    log.targetId = targetId;
  }
  if (oldValue !== undefined) {
    log.oldValue = oldValue;
  }
  if (newValue !== undefined) {
    log.newValue = newValue;
  }
  if (ipAddress !== undefined) {
    log.ipAddress = ipAddress;
  }
  addAuditLog(log);
}

function createError(code: ErrorCode, message?: string): ServiceResult<never> {
  const errorMessages: Record<ErrorCode, string> = {
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
    [ErrorCode.EXPENSE_LOCKED]: '报销单已锁定',
    [ErrorCode.APPROVAL_INSUFFICIENT_LEVEL]: '审批级别不足',
    [ErrorCode.APPROVAL_ALREADY_COMPLETED]: '审批已完成',
    [ErrorCode.APPROVAL_NEEDS_SECOND_LEVEL]: '需要二级审批',
    [ErrorCode.APPROVAL_LARGE_AMOUNT_REQUIRES_INDIVIDUAL]: '大额报销需单独审批',
    [ErrorCode.UNAUTHORIZED]: '未授权',
    [ErrorCode.FORBIDDEN]: '禁止访问',
    [ErrorCode.INSUFFICIENT_PERMISSION]: '权限不足',
    [ErrorCode.INTERNAL_ERROR]: '内部错误',
  };
  return {
    success: false,
    error: {
      code,
      message: message || errorMessages[code] || '未知错误',
    },
  };
}

export function createExpense(params: {
  employeeId: string;
  employeeName: string;
  date: string;
  amount: number;
  category: ExpenseCategory;
  reason: string;
  voucherFileName: string;
}): ServiceResult<ExpenseItem> {
  const validation = validateCreateExpenseParams(params);
  if (!validation.valid) {
    return createError(validation.errorCode || ErrorCode.INVALID_PARAMETER, validation.errorMessage);
  }

  const limitCheck = checkExpenseLimit(params.category, params.amount);

  const now = getCurrentTime();
  const expense: ExpenseItem = {
    id: generateId(),
    employeeId: params.employeeId,
    employeeName: params.employeeName,
    date: params.date,
    amount: params.amount,
    category: params.category,
    reason: params.reason,
    voucherFileName: params.voucherFileName,
    status: 'pending',
    needsSupplementaryInfo: limitCheck.needsSupplementary,
    createdAt: now,
    updatedAt: now,
  };

  saveExpense(expense);

  recordOperation(
    params.employeeId,
    params.employeeName,
    '创建报销单',
    'expense',
    expense.id,
    { category: params.category, amount: params.amount }
  );

  recordAudit(
    params.employeeId,
    params.employeeName,
    'create_expense',
    'expense',
    expense.id,
    undefined,
    { ...expense }
  );

  return { success: true, data: expense };
}

export function updateExpense(params: {
  expenseId: string;
  date?: string;
  amount?: number;
  category?: ExpenseCategory;
  reason?: string;
  voucherFileName?: string;
  supplementaryInfo?: string;
  operatorId: string;
  operatorName: string;
}): ServiceResult<ExpenseItem> {
  const idValidation = validateExpenseId(params.expenseId);
  if (!idValidation.valid) {
    return createError(idValidation.errorCode || ErrorCode.INVALID_PARAMETER, idValidation.errorMessage);
  }

  const validation = validateUpdateExpenseParams({
    date: params.date,
    amount: params.amount,
    category: params.category,
    reason: params.reason,
    voucherFileName: params.voucherFileName,
    supplementaryInfo: params.supplementaryInfo,
  });
  if (!validation.valid) {
    return createError(validation.errorCode || ErrorCode.INVALID_PARAMETER, validation.errorMessage);
  }

  const expense = getExpenseById(params.expenseId);
  if (!expense) {
    return createError(ErrorCode.EXPENSE_NOT_FOUND);
  }

  if (expense.status !== 'pending') {
    return createError(ErrorCode.EXPENSE_LOCKED, '报销单已提交或不在草稿状态，无法修改');
  }

  const oldValue = { ...expense };
  let hasAmountChange = false;

  if (params.date !== undefined) {
    expense.date = params.date;
  }
  if (params.amount !== undefined) {
    expense.amount = params.amount;
    hasAmountChange = true;
  }
  if (params.category !== undefined) {
    expense.category = params.category;
  }
  if (params.reason !== undefined) {
    expense.reason = params.reason;
  }
  if (params.voucherFileName !== undefined) {
    expense.voucherFileName = params.voucherFileName;
  }
  if (params.supplementaryInfo !== undefined) {
    expense.supplementaryInfo = params.supplementaryInfo;
    if (params.supplementaryInfo.trim() !== '') {
      expense.needsSupplementaryInfo = false;
    }
  }

  if (params.category !== undefined || params.amount !== undefined) {
    const limitCheck = checkExpenseLimit(
      expense.category,
      expense.amount
    );
    if (limitCheck.needsSupplementary && !expense.supplementaryInfo) {
      expense.needsSupplementaryInfo = true;
    }
  }

  expense.updatedAt = getCurrentTime();
  saveExpense(expense);

  recordOperation(
    params.operatorId,
    params.operatorName,
    '更新报销单',
    'expense',
    expense.id,
    { updatedFields: Object.keys(params).filter(k => k !== 'expenseId' && k !== 'operatorId' && k !== 'operatorName') }
  );

  const auditAction: AuditActionType = hasAmountChange ? 'update_amount' : 'update_expense';
  recordAudit(
    params.operatorId,
    params.operatorName,
    auditAction,
    'expense',
    expense.id,
    oldValue,
    { ...expense }
  );

  return { success: true, data: expense };
}

export function submitExpense(params: {
  expenseId: string;
  operatorId: string;
  operatorName: string;
}): ServiceResult<ExpenseItem> {
  const idValidation = validateExpenseId(params.expenseId);
  if (!idValidation.valid) {
    return createError(idValidation.errorCode || ErrorCode.INVALID_PARAMETER, idValidation.errorMessage);
  }

  const expense = getExpenseById(params.expenseId);
  if (!expense) {
    return createError(ErrorCode.EXPENSE_NOT_FOUND);
  }

  const suppValidation = validateSupplementaryInfo(
    expense.needsSupplementaryInfo,
    expense.supplementaryInfo
  );
  if (!suppValidation.valid) {
    return createError(suppValidation.errorCode || ErrorCode.EXPENSE_NEEDS_SUPPLEMENT, suppValidation.errorMessage);
  }

  recordOperation(
    params.operatorId,
    params.operatorName,
    '提交报销单',
    'expense',
    expense.id
  );

  recordAudit(
    params.operatorId,
    params.operatorName,
    'submit_expense',
    'expense',
    expense.id,
    { status: 'pending' },
    { status: 'pending' }
  );

  return { success: true, data: expense };
}

export function approveExpense(params: {
  expenseId: string;
  approverId: string;
  approverName: string;
  approverRole: UserRole;
  comment?: string;
}): ServiceResult<ExpenseItem> & { needsSecondLevel?: boolean } {
  const validation = validateApproveParams({
    expenseId: params.expenseId,
    approverId: params.approverId,
    approverName: params.approverName,
  });
  if (!validation.valid) {
    return createError(validation.errorCode || ErrorCode.INVALID_PARAMETER, validation.errorMessage);
  }

  const expense = getExpenseById(params.expenseId);
  if (!expense) {
    return createError(ErrorCode.EXPENSE_NOT_FOUND);
  }

  const statusValidation = validateExpenseStatusForApproval(expense.status);
  if (!statusValidation.valid) {
    return createError(statusValidation.errorCode || ErrorCode.EXPENSE_NOT_PENDING, statusValidation.errorMessage);
  }

  const isLarge = isLargeAmount(expense.amount);
  const currentLevel = expense.approvalLevel;

  if (isLarge) {
    if (params.approverRole === 'dept_manager') {
      if (currentLevel === undefined) {
        const record: ApprovalRecord = {
          id: generateId(),
          expenseId: expense.id,
          approverId: params.approverId,
          approverName: params.approverName,
          approvalLevel: 'dept_manager',
          action: 'approve',
          createdAt: getCurrentTime(),
        };
        if (params.comment !== undefined) {
          record.comment = params.comment;
        }
        addApprovalRecord(record);
        expense.approvalLevel = 'dept_manager';
        expense.updatedAt = getCurrentTime();
        saveExpense(expense);

        recordOperation(
          params.approverId,
          params.approverName,
          '部门经理审批通过',
          'expense',
          expense.id
        );

        recordAudit(
          params.approverId,
          params.approverName,
          'approve_expense',
          'expense',
          expense.id
        );

        return { success: true, data: expense, needsSecondLevel: true };
      } else if (currentLevel === 'dept_manager') {
        return createError(ErrorCode.APPROVAL_ALREADY_COMPLETED, '部门经理已审批，等待财务总监审批');
      }
    } else if (params.approverRole === 'finance_director') {
      if (currentLevel === undefined) {
        return createError(ErrorCode.APPROVAL_NEEDS_SECOND_LEVEL, '大额报销需先由部门经理审批');
      } else if (currentLevel === 'dept_manager') {
        const record: ApprovalRecord = {
          id: generateId(),
          expenseId: expense.id,
          approverId: params.approverId,
          approverName: params.approverName,
          approvalLevel: 'finance_director',
          action: 'approve',
          createdAt: getCurrentTime(),
        };
        if (params.comment !== undefined) {
          record.comment = params.comment;
        }
        addApprovalRecord(record);
        expense.status = 'approved';
        delete (expense as unknown as Record<string, unknown>)['approvalLevel'];
        expense.updatedAt = getCurrentTime();
        saveExpense(expense);

        recordOperation(
          params.approverId,
          params.approverName,
          '财务总监审批通过',
          'expense',
          expense.id
        );

        recordAudit(
          params.approverId,
          params.approverName,
          'approve_expense',
          'expense',
          expense.id
        );

        return { success: true, data: expense };
      }
    } else if (params.approverRole === 'admin') {
      const record: ApprovalRecord = {
        id: generateId(),
        expenseId: expense.id,
        approverId: params.approverId,
        approverName: params.approverName,
        approvalLevel: 'dept_manager',
        action: 'approve',
        createdAt: getCurrentTime(),
      };
      if (params.comment !== undefined) {
        record.comment = params.comment;
      }
      addApprovalRecord(record);
      expense.status = 'approved';
      delete (expense as unknown as Record<string, unknown>)['approvalLevel'];
      expense.updatedAt = getCurrentTime();
      saveExpense(expense);

      recordOperation(
        params.approverId,
        params.approverName,
        '管理员审批通过',
        'expense',
        expense.id
      );

      recordAudit(
        params.approverId,
        params.approverName,
        'approve_expense',
        'expense',
        expense.id
      );

      return { success: true, data: expense };
    } else {
      return createError(ErrorCode.INSUFFICIENT_PERMISSION, '普通员工无审批权限');
    }
  } else {
    if (params.approverRole === 'employee') {
      return createError(ErrorCode.INSUFFICIENT_PERMISSION, '普通员工无审批权限');
    }

    const record: ApprovalRecord = {
      id: generateId(),
      expenseId: expense.id,
      approverId: params.approverId,
      approverName: params.approverName,
      approvalLevel: 'dept_manager',
      action: 'approve',
      createdAt: getCurrentTime(),
    };
    if (params.comment !== undefined) {
      record.comment = params.comment;
    }
    addApprovalRecord(record);
    expense.status = 'approved';
    expense.updatedAt = getCurrentTime();
    saveExpense(expense);

    recordOperation(
      params.approverId,
      params.approverName,
      '审批通过',
      'expense',
      expense.id
    );

    recordAudit(
      params.approverId,
      params.approverName,
      'approve_expense',
      'expense',
      expense.id
    );

    return { success: true, data: expense };
  }

  return createError(ErrorCode.INTERNAL_ERROR);
}

export function rejectExpense(params: {
  expenseId: string;
  approverId: string;
  approverName: string;
  approverRole: UserRole;
  comment?: string;
}): ServiceResult<ExpenseItem> {
  const validation = validateApproveParams({
    expenseId: params.expenseId,
    approverId: params.approverId,
    approverName: params.approverName,
  });
  if (!validation.valid) {
    return createError(validation.errorCode || ErrorCode.INVALID_PARAMETER, validation.errorMessage);
  }

  const expense = getExpenseById(params.expenseId);
  if (!expense) {
    return createError(ErrorCode.EXPENSE_NOT_FOUND);
  }

  if (params.approverRole === 'employee') {
    return createError(ErrorCode.INSUFFICIENT_PERMISSION, '普通员工无驳回权限');
  }

  const oldValue = { ...expense };
  expense.status = 'rejected';
  delete (expense as unknown as Record<string, unknown>)['approvalLevel'];
  expense.updatedAt = getCurrentTime();

  clearApprovalRecords(expense.id);

  saveExpense(expense);

  const operationDetails: Record<string, unknown> = {};
  if (params.comment !== undefined) {
    operationDetails['comment'] = params.comment;
  }
  recordOperation(
    params.approverId,
    params.approverName,
    '驳回报销单',
    'expense',
    expense.id,
    Object.keys(operationDetails).length > 0 ? operationDetails : undefined
  );

  recordAudit(
    params.approverId,
    params.approverName,
    'reject_expense',
    'expense',
    expense.id,
    oldValue
  );

  return { success: true, data: expense };
}

export function batchApprove(params: {
  expenseIds: string[];
  approverId: string;
  approverName: string;
  approverRole: UserRole;
  comment?: string;
}): ServiceResult<BatchApproveResult> {
  const validation = validateBatchApproveParams({
    expenseIds: params.expenseIds,
    approverId: params.approverId,
    approverName: params.approverName,
  });
  if (!validation.valid) {
    return createError(validation.errorCode || ErrorCode.INVALID_PARAMETER, validation.errorMessage);
  }

  const result: BatchApproveResult = {
    approved: [],
    skipped: [],
    failed: [],
  };

  for (const expenseId of params.expenseIds) {
    const expense = getExpenseById(expenseId);
    if (!expense) {
      result.failed.push({
        expenseId,
        code: ErrorCode.EXPENSE_NOT_FOUND,
        message: '报销单不存在',
      });
      continue;
    }

    if (isLargeAmount(expense.amount)) {
      result.skipped.push(expenseId);
      continue;
    }

    if (expense.status !== 'pending') {
      result.failed.push({
        expenseId,
        code: ErrorCode.EXPENSE_NOT_PENDING,
        message: '报销单不在待审批状态',
      });
      continue;
    }

    const approveParams: {
      expenseId: string;
      approverId: string;
      approverName: string;
      approverRole: UserRole;
      comment?: string;
    } = {
      expenseId,
      approverId: params.approverId,
      approverName: params.approverName,
      approverRole: params.approverRole,
    };
    if (params.comment !== undefined) {
      approveParams.comment = params.comment;
    }
    const approveResult = approveExpense(approveParams);

    if (approveResult.success) {
      result.approved.push(expenseId);
    } else {
      result.failed.push({
        expenseId,
        code: approveResult.error?.code || ErrorCode.INTERNAL_ERROR,
        message: approveResult.error?.message || '审批失败',
      });
    }
  }

  recordOperation(
    params.approverId,
    params.approverName,
    '批量审批',
    'system',
    undefined,
    {
      total: params.expenseIds.length,
      approved: result.approved.length,
      skipped: result.skipped.length,
      failed: result.failed.length,
    }
  );

  recordAudit(
    params.approverId,
    params.approverName,
    'batch_approve',
    'system',
    undefined,
    undefined,
    {
      expenseIds: params.expenseIds,
      result,
    }
  );

  return { success: true, data: result };
}

export function payExpense(params: {
  expenseId: string;
  operatorId: string;
  operatorName: string;
  operatorRole: UserRole;
}): ServiceResult<ExpenseItem> {
  const idValidation = validateExpenseId(params.expenseId);
  if (!idValidation.valid) {
    return createError(idValidation.errorCode || ErrorCode.INVALID_PARAMETER, idValidation.errorMessage);
  }

  const expense = getExpenseById(params.expenseId);
  if (!expense) {
    return createError(ErrorCode.EXPENSE_NOT_FOUND);
  }

  if (expense.status !== 'approved') {
    return createError(ErrorCode.EXPENSE_NOT_PENDING, '报销单不在已批准状态，无法打款');
  }

  if (params.operatorRole !== 'finance_director' && params.operatorRole !== 'admin') {
    return createError(ErrorCode.INSUFFICIENT_PERMISSION, '只有财务总监或管理员可以执行打款操作');
  }

  const oldValue = { ...expense };
  expense.status = 'paid';
  expense.updatedAt = getCurrentTime();
  saveExpense(expense);

  recordOperation(
    params.operatorId,
    params.operatorName,
    '打款确认',
    'expense',
    expense.id
  );

  recordAudit(
    params.operatorId,
    params.operatorName,
    'pay_expense',
    'expense',
    expense.id,
    oldValue,
    { status: 'paid' }
  );

  return { success: true, data: expense };
}

export function resubmitExpense(params: {
  expenseId: string;
  operatorId: string;
  operatorName: string;
}): ServiceResult<ExpenseItem> {
  const idValidation = validateExpenseId(params.expenseId);
  if (!idValidation.valid) {
    return createError(idValidation.errorCode || ErrorCode.INVALID_PARAMETER, idValidation.errorMessage);
  }

  const expense = getExpenseById(params.expenseId);
  if (!expense) {
    return createError(ErrorCode.EXPENSE_NOT_FOUND);
  }

  if (expense.status !== 'rejected') {
    return createError(ErrorCode.INVALID_STATUS, '只有已驳回的报销单可以重新提交');
  }

  const oldValue = { ...expense };
  expense.status = 'pending';
  delete (expense as unknown as Record<string, unknown>)['approvalLevel'];
  expense.updatedAt = getCurrentTime();

  clearApprovalRecords(expense.id);

  saveExpense(expense);

  recordOperation(
    params.operatorId,
    params.operatorName,
    '重新提交报销单',
    'expense',
    expense.id
  );

  recordAudit(
    params.operatorId,
    params.operatorName,
    'submit_expense',
    'expense',
    expense.id,
    oldValue
  );

  return { success: true, data: expense };
}

export function deleteExpenseService(params: {
  expenseId: string;
  operatorId: string;
  operatorName: string;
  operatorRole: UserRole;
}): ServiceResult<void> {
  const idValidation = validateExpenseId(params.expenseId);
  if (!idValidation.valid) {
    return createError(idValidation.errorCode || ErrorCode.INVALID_PARAMETER, idValidation.errorMessage);
  }

  if (params.operatorRole !== 'admin') {
    return createError(ErrorCode.INSUFFICIENT_PERMISSION, '只有管理员可以删除报销单');
  }

  const expense = getExpenseById(params.expenseId);
  if (!expense) {
    return createError(ErrorCode.EXPENSE_NOT_FOUND);
  }

  const oldValue = { ...expense };
  clearApprovalRecords(params.expenseId);
  deleteExpense(params.expenseId);

  recordOperation(
    params.operatorId,
    params.operatorName,
    '删除报销单',
    'expense',
    params.expenseId
  );

  recordAudit(
    params.operatorId,
    params.operatorName,
    'delete_expense',
    'expense',
    params.expenseId,
    oldValue,
    undefined
  );

  return { success: true };
}

export function getExpense(expenseId: string): ServiceResult<ExpenseItem> {
  const idValidation = validateExpenseId(expenseId);
  if (!idValidation.valid) {
    return createError(idValidation.errorCode || ErrorCode.INVALID_PARAMETER, idValidation.errorMessage);
  }

  const expense = getExpenseById(expenseId);
  if (!expense) {
    return createError(ErrorCode.EXPENSE_NOT_FOUND);
  }

  return { success: true, data: expense };
}

export function listExpenses(params: {
  startDate?: string;
  endDate?: string;
  status?: ExpenseStatus;
  employeeId?: string;
  page?: number;
  pageSize?: number;
}): ServiceResult<{
  items: ExpenseItem[];
  total: number;
  page: number;
  pageSize: number;
}> {
  const validation = validateListParams(params);
  if (!validation.valid) {
    return createError(validation.errorCode || ErrorCode.INVALID_PARAMETER, validation.errorMessage);
  }

  const page = params.page ?? DEFAULT_PAGE;
  const pageSize = params.pageSize ?? DEFAULT_PAGE_SIZE;

  let allExpenses = getAllExpenses();

  if (params.startDate || params.endDate) {
    allExpenses = allExpenses.filter(e =>
      isDateInRange(e.date, params.startDate, params.endDate)
    );
  }

  if (params.status) {
    allExpenses = allExpenses.filter(e => e.status === params.status);
  }

  if (params.employeeId) {
    allExpenses = allExpenses.filter(e => e.employeeId === params.employeeId);
  }

  allExpenses.sort((a, b) =>
    new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
  );

  const total = allExpenses.length;
  const startIndex = (page - 1) * pageSize;
  const endIndex = startIndex + pageSize;
  const items = allExpenses.slice(startIndex, endIndex);

  return {
    success: true,
    data: {
      items,
      total,
      page,
      pageSize,
    },
  };
}

export function listOperationLogs(params: {
  startDate?: string;
  endDate?: string;
  userId?: string;
  page?: number;
  pageSize?: number;
}): ServiceResult<{
  items: OperationLog[];
  total: number;
  page: number;
  pageSize: number;
}> {
  const page = params.page ?? DEFAULT_PAGE;
  const pageSize = params.pageSize ?? DEFAULT_PAGE_SIZE;

  let logs = getAllOperationLogs();

  if (params.startDate || params.endDate) {
    logs = logs.filter(log =>
      isDateInRange(log.createdAt, params.startDate, params.endDate)
    );
  }

  if (params.userId) {
    logs = logs.filter(log => log.userId === params.userId);
  }

  logs.sort((a, b) =>
    new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
  );

  const total = logs.length;
  const startIndex = (page - 1) * pageSize;
  const endIndex = startIndex + pageSize;
  const items = logs.slice(startIndex, endIndex);

  return {
    success: true,
    data: {
      items,
      total,
      page,
      pageSize,
    },
  };
}

export function listAuditLogs(params: {
  startDate?: string;
  endDate?: string;
  userId?: string;
  action?: string;
  page?: number;
  pageSize?: number;
}): ServiceResult<{
  items: AuditLog[];
  total: number;
  page: number;
  pageSize: number;
}> {
  const page = params.page ?? DEFAULT_PAGE;
  const pageSize = params.pageSize ?? DEFAULT_PAGE_SIZE;

  let logs = getAllAuditLogs();

  if (params.startDate || params.endDate) {
    logs = logs.filter(log =>
      isDateInRange(log.createdAt, params.startDate, params.endDate)
    );
  }

  if (params.userId) {
    logs = logs.filter(log => log.userId === params.userId);
  }

  if (params.action) {
    logs = logs.filter(log => log.action === params.action);
  }

  logs.sort((a, b) =>
    new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
  );

  const total = logs.length;
  const startIndex = (page - 1) * pageSize;
  const endIndex = startIndex + pageSize;
  const items = logs.slice(startIndex, endIndex);

  return {
    success: true,
    data: {
      items,
      total,
      page,
      pageSize,
    },
  };
}

export function getUser(userId: string): ServiceResult<User> {
  const validation = validateUserId(userId);
  if (!validation.valid) {
    return createError(validation.errorCode || ErrorCode.INVALID_PARAMETER, validation.errorMessage);
  }

  const user = getUserById(userId);
  if (!user) {
    return createError(ErrorCode.UNAUTHORIZED, '用户不存在');
  }

  return { success: true, data: user };
}
