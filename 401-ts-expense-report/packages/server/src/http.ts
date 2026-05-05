import type { IncomingMessage, ServerResponse } from 'http';
import {
  SERVER_PORT,
} from '@expense-report/shared';
import {
  createExpense,
  updateExpense,
  submitExpense,
  approveExpense,
  rejectExpense,
  batchApprove,
  payExpense,
  resubmitExpense,
  deleteExpenseService,
  getExpense,
  listExpenses,
  listOperationLogs,
  listAuditLogs,
  getUser,
} from './service';

function parseJson<T>(body: string): T | null {
  try {
    return JSON.parse(body) as T;
  } catch {
    return null;
  }
}

function sendJson(res: ServerResponse, statusCode: number, data: unknown): void {
  const responseBody = JSON.stringify(data, null, 2);
  res.statusCode = statusCode;
  res.setHeader('Content-Type', 'application/json; charset=utf-8');
  res.setHeader('Content-Length', Buffer.byteLength(responseBody));
  res.end(responseBody);
}

function sendError(res: ServerResponse, statusCode: number, error: { code: number; message: string }): void {
  sendJson(res, statusCode, {
    success: false,
    error,
  });
}

function sendSuccess<T>(res: ServerResponse, data: T, extra?: Record<string, unknown>): void {
  const response: Record<string, unknown> = {
    success: true,
    data,
  };
  if (extra !== undefined) {
    Object.assign(response, extra);
  }
  sendJson(res, 200, response);
}

async function readBody(req: IncomingMessage): Promise<string> {
  return new Promise((resolve, reject) => {
    const chunks: Buffer[] = [];
    req.on('data', (chunk: Buffer) => {
      chunks.push(chunk);
    });
    req.on('end', () => {
      resolve(Buffer.concat(chunks).toString('utf-8'));
    });
    req.on('error', (err) => {
      reject(err);
    });
  });
}

function getUserId(req: IncomingMessage): string | null {
  const userIdHeader = req.headers['x-user-id'];
  if (typeof userIdHeader === 'string') {
    return userIdHeader;
  }
  if (Array.isArray(userIdHeader) && userIdHeader.length > 0) {
    const firstId = userIdHeader[0];
    return firstId !== undefined ? firstId : null;
  }
  return null;
}

function parseQueryParams(url: string): URLSearchParams {
  const urlObj = new URL(url, `http://localhost:${SERVER_PORT}`);
  return urlObj.searchParams;
}

export async function handleCreateExpense(req: IncomingMessage, res: ServerResponse): Promise<void> {
  const userId = getUserId(req);
  if (!userId) {
    sendError(res, 401, { code: 4001, message: '缺少用户标识，请在请求头中提供 X-User-Id' });
    return;
  }

  const bodyStr = await readBody(req);
  const body = parseJson<{
    employeeId: string;
    employeeName: string;
    date: string;
    amount: number;
    category: string;
    reason: string;
    voucherFileName: string;
  }>(bodyStr);

  if (!body) {
    sendError(res, 400, { code: 1001, message: '请求体格式无效' });
    return;
  }

  const result = createExpense({
    employeeId: body.employeeId,
    employeeName: body.employeeName,
    date: body.date,
    amount: body.amount,
    category: body.category as 'transport' | 'dining' | 'accommodation' | 'office' | 'other',
    reason: body.reason,
    voucherFileName: body.voucherFileName,
  });

  if (result.success && result.data) {
    sendSuccess(res, result.data);
  } else if (result.error) {
    sendError(res, 400, result.error);
  } else {
    sendError(res, 500, { code: 5001, message: '内部错误' });
  }
}

export async function handleUpdateExpense(req: IncomingMessage, res: ServerResponse): Promise<void> {
  const userId = getUserId(req);
  if (!userId) {
    sendError(res, 401, { code: 4001, message: '缺少用户标识，请在请求头中提供 X-User-Id' });
    return;
  }

  const bodyStr = await readBody(req);
  const body = parseJson<{
    expenseId: string;
    date?: string;
    amount?: number;
    category?: string;
    reason?: string;
    voucherFileName?: string;
    supplementaryInfo?: string;
    operatorName?: string;
  }>(bodyStr);

  if (!body || !body.expenseId) {
    sendError(res, 400, { code: 1001, message: '请求体格式无效或缺少 expenseId' });
    return;
  }

  const userResult = getUser(userId);
  const operatorName = body.operatorName || (userResult.success ? userResult.data?.name : userId);

  const updateParams: {
    expenseId: string;
    date?: string;
    amount?: number;
    category?: 'transport' | 'dining' | 'accommodation' | 'office' | 'other';
    reason?: string;
    voucherFileName?: string;
    supplementaryInfo?: string;
    operatorId: string;
    operatorName: string;
  } = {
    expenseId: body.expenseId,
    operatorId: userId,
    operatorName: operatorName || userId,
  };
  if (body.date !== undefined) {
    updateParams.date = body.date;
  }
  if (body.amount !== undefined) {
    updateParams.amount = body.amount;
  }
  if (body.category !== undefined) {
    updateParams.category = body.category as 'transport' | 'dining' | 'accommodation' | 'office' | 'other';
  }
  if (body.reason !== undefined) {
    updateParams.reason = body.reason;
  }
  if (body.voucherFileName !== undefined) {
    updateParams.voucherFileName = body.voucherFileName;
  }
  if (body.supplementaryInfo !== undefined) {
    updateParams.supplementaryInfo = body.supplementaryInfo;
  }

  const result = updateExpense(updateParams);

  if (result.success && result.data) {
    sendSuccess(res, result.data);
  } else if (result.error) {
    sendError(res, 400, result.error);
  } else {
    sendError(res, 500, { code: 5001, message: '内部错误' });
  }
}

export async function handleSubmitExpense(req: IncomingMessage, res: ServerResponse): Promise<void> {
  const userId = getUserId(req);
  if (!userId) {
    sendError(res, 401, { code: 4001, message: '缺少用户标识，请在请求头中提供 X-User-Id' });
    return;
  }

  const bodyStr = await readBody(req);
  const body = parseJson<{
    expenseId: string;
    operatorName?: string;
  }>(bodyStr);

  if (!body || !body.expenseId) {
    sendError(res, 400, { code: 1001, message: '请求体格式无效或缺少 expenseId' });
    return;
  }

  const userResult = getUser(userId);
  const operatorName = body.operatorName || (userResult.success ? userResult.data?.name : userId);

  const result = submitExpense({
    expenseId: body.expenseId,
    operatorId: userId,
    operatorName: operatorName || userId,
  });

  if (result.success && result.data) {
    sendSuccess(res, result.data);
  } else if (result.error) {
    sendError(res, 400, result.error);
  } else {
    sendError(res, 500, { code: 5001, message: '内部错误' });
  }
}

export async function handleApproveExpense(req: IncomingMessage, res: ServerResponse): Promise<void> {
  const userId = getUserId(req);
  if (!userId) {
    sendError(res, 401, { code: 4001, message: '缺少用户标识，请在请求头中提供 X-User-Id' });
    return;
  }

  const userResult = getUser(userId);
  if (!userResult.success || !userResult.data) {
    sendError(res, 401, { code: 4001, message: '用户不存在' });
    return;
  }

  const bodyStr = await readBody(req);
  const body = parseJson<{
    expenseId: string;
    approverName?: string;
    comment?: string;
  }>(bodyStr);

  if (!body || !body.expenseId) {
    sendError(res, 400, { code: 1001, message: '请求体格式无效或缺少 expenseId' });
    return;
  }

  const approverName = body.approverName || userResult.data.name;

  const approveParams: {
    expenseId: string;
    approverId: string;
    approverName: string;
    approverRole: 'employee' | 'dept_manager' | 'finance_director' | 'admin';
    comment?: string;
  } = {
    expenseId: body.expenseId,
    approverId: userId,
    approverName,
    approverRole: userResult.data.role,
  };
  if (body.comment !== undefined) {
    approveParams.comment = body.comment;
  }
  const result = approveExpense(approveParams);

  if (result.success && result.data) {
    if (result.needsSecondLevel) {
      sendSuccess(res, result.data, { needsSecondLevel: true });
    } else {
      sendSuccess(res, result.data);
    }
  } else if (result.error) {
    sendError(res, 400, result.error);
  } else {
    sendError(res, 500, { code: 5001, message: '内部错误' });
  }
}

export async function handleRejectExpense(req: IncomingMessage, res: ServerResponse): Promise<void> {
  const userId = getUserId(req);
  if (!userId) {
    sendError(res, 401, { code: 4001, message: '缺少用户标识，请在请求头中提供 X-User-Id' });
    return;
  }

  const userResult = getUser(userId);
  if (!userResult.success || !userResult.data) {
    sendError(res, 401, { code: 4001, message: '用户不存在' });
    return;
  }

  const bodyStr = await readBody(req);
  const body = parseJson<{
    expenseId: string;
    approverName?: string;
    comment?: string;
  }>(bodyStr);

  if (!body || !body.expenseId) {
    sendError(res, 400, { code: 1001, message: '请求体格式无效或缺少 expenseId' });
    return;
  }

  const approverName = body.approverName || userResult.data.name;

  const rejectParams: {
    expenseId: string;
    approverId: string;
    approverName: string;
    approverRole: 'employee' | 'dept_manager' | 'finance_director' | 'admin';
    comment?: string;
  } = {
    expenseId: body.expenseId,
    approverId: userId,
    approverName,
    approverRole: userResult.data.role,
  };
  if (body.comment !== undefined) {
    rejectParams.comment = body.comment;
  }
  const result = rejectExpense(rejectParams);

  if (result.success && result.data) {
    sendSuccess(res, result.data);
  } else if (result.error) {
    sendError(res, 400, result.error);
  } else {
    sendError(res, 500, { code: 5001, message: '内部错误' });
  }
}

export async function handleBatchApprove(req: IncomingMessage, res: ServerResponse): Promise<void> {
  const userId = getUserId(req);
  if (!userId) {
    sendError(res, 401, { code: 4001, message: '缺少用户标识，请在请求头中提供 X-User-Id' });
    return;
  }

  const userResult = getUser(userId);
  if (!userResult.success || !userResult.data) {
    sendError(res, 401, { code: 4001, message: '用户不存在' });
    return;
  }

  const bodyStr = await readBody(req);
  const body = parseJson<{
    expenseIds: string[];
    approverName?: string;
    comment?: string;
  }>(bodyStr);

  if (!body || !Array.isArray(body.expenseIds)) {
    sendError(res, 400, { code: 1001, message: '请求体格式无效或缺少 expenseIds 数组' });
    return;
  }

  const approverName = body.approverName || userResult.data.name;

  const batchParams: {
    expenseIds: string[];
    approverId: string;
    approverName: string;
    approverRole: 'employee' | 'dept_manager' | 'finance_director' | 'admin';
    comment?: string;
  } = {
    expenseIds: body.expenseIds,
    approverId: userId,
    approverName,
    approverRole: userResult.data.role,
  };
  if (body.comment !== undefined) {
    batchParams.comment = body.comment;
  }
  const result = batchApprove(batchParams);

  if (result.success && result.data) {
    sendSuccess(res, result.data);
  } else if (result.error) {
    sendError(res, 400, result.error);
  } else {
    sendError(res, 500, { code: 5001, message: '内部错误' });
  }
}

export async function handlePayExpense(req: IncomingMessage, res: ServerResponse): Promise<void> {
  const userId = getUserId(req);
  if (!userId) {
    sendError(res, 401, { code: 4001, message: '缺少用户标识，请在请求头中提供 X-User-Id' });
    return;
  }

  const userResult = getUser(userId);
  if (!userResult.success || !userResult.data) {
    sendError(res, 401, { code: 4001, message: '用户不存在' });
    return;
  }

  const bodyStr = await readBody(req);
  const body = parseJson<{
    expenseId: string;
    operatorName?: string;
  }>(bodyStr);

  if (!body || !body.expenseId) {
    sendError(res, 400, { code: 1001, message: '请求体格式无效或缺少 expenseId' });
    return;
  }

  const operatorName = body.operatorName || userResult.data.name;

  const result = payExpense({
    expenseId: body.expenseId,
    operatorId: userId,
    operatorName,
    operatorRole: userResult.data.role,
  });

  if (result.success && result.data) {
    sendSuccess(res, result.data);
  } else if (result.error) {
    sendError(res, 400, result.error);
  } else {
    sendError(res, 500, { code: 5001, message: '内部错误' });
  }
}

export async function handleResubmitExpense(req: IncomingMessage, res: ServerResponse): Promise<void> {
  const userId = getUserId(req);
  if (!userId) {
    sendError(res, 401, { code: 4001, message: '缺少用户标识，请在请求头中提供 X-User-Id' });
    return;
  }

  const bodyStr = await readBody(req);
  const body = parseJson<{
    expenseId: string;
    operatorName?: string;
  }>(bodyStr);

  if (!body || !body.expenseId) {
    sendError(res, 400, { code: 1001, message: '请求体格式无效或缺少 expenseId' });
    return;
  }

  const userResult = getUser(userId);
  const operatorName = body.operatorName || (userResult.success ? userResult.data?.name : userId);

  const result = resubmitExpense({
    expenseId: body.expenseId,
    operatorId: userId,
    operatorName: operatorName || userId,
  });

  if (result.success && result.data) {
    sendSuccess(res, result.data);
  } else if (result.error) {
    sendError(res, 400, result.error);
  } else {
    sendError(res, 500, { code: 5001, message: '内部错误' });
  }
}

export async function handleDeleteExpense(req: IncomingMessage, res: ServerResponse): Promise<void> {
  const userId = getUserId(req);
  if (!userId) {
    sendError(res, 401, { code: 4001, message: '缺少用户标识，请在请求头中提供 X-User-Id' });
    return;
  }

  const userResult = getUser(userId);
  if (!userResult.success || !userResult.data) {
    sendError(res, 401, { code: 4001, message: '用户不存在' });
    return;
  }

  const bodyStr = await readBody(req);
  const body = parseJson<{
    expenseId: string;
    operatorName?: string;
  }>(bodyStr);

  if (!body || !body.expenseId) {
    sendError(res, 400, { code: 1001, message: '请求体格式无效或缺少 expenseId' });
    return;
  }

  const operatorName = body.operatorName || userResult.data.name;

  const result = deleteExpenseService({
    expenseId: body.expenseId,
    operatorId: userId,
    operatorName,
    operatorRole: userResult.data.role,
  });

  if (result.success) {
    sendSuccess(res, { deleted: true });
  } else if (result.error) {
    sendError(res, 400, result.error);
  } else {
    sendError(res, 500, { code: 5001, message: '内部错误' });
  }
}

export async function handleGetExpense(req: IncomingMessage, res: ServerResponse): Promise<void> {
  const params = parseQueryParams(req.url || '');
  const expenseId = params.get('id');

  if (!expenseId) {
    sendError(res, 400, { code: 1002, message: '缺少 expenseId 参数' });
    return;
  }

  const result = getExpense(expenseId);

  if (result.success && result.data) {
    sendSuccess(res, result.data);
  } else if (result.error) {
    sendError(res, 404, result.error);
  } else {
    sendError(res, 500, { code: 5001, message: '内部错误' });
  }
}

export async function handleListExpenses(req: IncomingMessage, res: ServerResponse): Promise<void> {
  const params = parseQueryParams(req.url || '');

  const startDateRaw = params.get('startDate');
  const endDateRaw = params.get('endDate');
  const statusRaw = params.get('status');
  const employeeIdRaw = params.get('employeeId');
  const pageRaw = params.has('page') ? parseInt(params.get('page') || '1', 10) : undefined;
  const pageSizeRaw = params.has('pageSize') ? parseInt(params.get('pageSize') || '20', 10) : undefined;

  const listParams: {
    startDate?: string;
    endDate?: string;
    status?: 'pending' | 'approved' | 'paid' | 'rejected';
    employeeId?: string;
    page?: number;
    pageSize?: number;
  } = {};
  if (startDateRaw) {
    listParams.startDate = startDateRaw;
  }
  if (endDateRaw) {
    listParams.endDate = endDateRaw;
  }
  if (statusRaw) {
    listParams.status = statusRaw as 'pending' | 'approved' | 'paid' | 'rejected';
  }
  if (employeeIdRaw) {
    listParams.employeeId = employeeIdRaw;
  }
  if (pageRaw !== undefined) {
    listParams.page = pageRaw;
  }
  if (pageSizeRaw !== undefined) {
    listParams.pageSize = pageSizeRaw;
  }

  const result = listExpenses(listParams);

  if (result.success && result.data) {
    sendSuccess(res, result.data);
  } else if (result.error) {
    sendError(res, 400, result.error);
  } else {
    sendError(res, 500, { code: 5001, message: '内部错误' });
  }
}

export async function handleGetOperationLogs(req: IncomingMessage, res: ServerResponse): Promise<void> {
  const params = parseQueryParams(req.url || '');

  const startDateRaw = params.get('startDate');
  const endDateRaw = params.get('endDate');
  const userIdRaw = params.get('userId');
  const pageRaw = params.has('page') ? parseInt(params.get('page') || '1', 10) : undefined;
  const pageSizeRaw = params.has('pageSize') ? parseInt(params.get('pageSize') || '20', 10) : undefined;

  const listParams: {
    startDate?: string;
    endDate?: string;
    userId?: string;
    page?: number;
    pageSize?: number;
  } = {};
  if (startDateRaw) {
    listParams.startDate = startDateRaw;
  }
  if (endDateRaw) {
    listParams.endDate = endDateRaw;
  }
  if (userIdRaw) {
    listParams.userId = userIdRaw;
  }
  if (pageRaw !== undefined) {
    listParams.page = pageRaw;
  }
  if (pageSizeRaw !== undefined) {
    listParams.pageSize = pageSizeRaw;
  }

  const result = listOperationLogs(listParams);

  if (result.success && result.data) {
    sendSuccess(res, result.data);
  } else if (result.error) {
    sendError(res, 400, result.error);
  } else {
    sendError(res, 500, { code: 5001, message: '内部错误' });
  }
}

export async function handleGetAuditLogs(req: IncomingMessage, res: ServerResponse): Promise<void> {
  const params = parseQueryParams(req.url || '');

  const startDateRaw = params.get('startDate');
  const endDateRaw = params.get('endDate');
  const userIdRaw = params.get('userId');
  const actionRaw = params.get('action');
  const pageRaw = params.has('page') ? parseInt(params.get('page') || '1', 10) : undefined;
  const pageSizeRaw = params.has('pageSize') ? parseInt(params.get('pageSize') || '20', 10) : undefined;

  const listParams: {
    startDate?: string;
    endDate?: string;
    userId?: string;
    action?: string;
    page?: number;
    pageSize?: number;
  } = {};
  if (startDateRaw) {
    listParams.startDate = startDateRaw;
  }
  if (endDateRaw) {
    listParams.endDate = endDateRaw;
  }
  if (userIdRaw) {
    listParams.userId = userIdRaw;
  }
  if (actionRaw) {
    listParams.action = actionRaw;
  }
  if (pageRaw !== undefined) {
    listParams.page = pageRaw;
  }
  if (pageSizeRaw !== undefined) {
    listParams.pageSize = pageSizeRaw;
  }

  const result = listAuditLogs(listParams);

  if (result.success && result.data) {
    sendSuccess(res, result.data);
  } else if (result.error) {
    sendError(res, 400, result.error);
  } else {
    sendError(res, 500, { code: 5001, message: '内部错误' });
  }
}

export async function handleGetUser(req: IncomingMessage, res: ServerResponse): Promise<void> {
  const params = parseQueryParams(req.url || '');
  const userId = params.get('id');

  if (!userId) {
    sendError(res, 400, { code: 1002, message: '缺少 id 参数' });
    return;
  }

  const result = getUser(userId);

  if (result.success && result.data) {
    sendSuccess(res, result.data);
  } else if (result.error) {
    sendError(res, 404, result.error);
  } else {
    sendError(res, 500, { code: 5001, message: '内部错误' });
  }
}

export function handleNotFound(_req: IncomingMessage, res: ServerResponse): void {
  sendError(res, 404, { code: 1000, message: '接口不存在' });
}

export function handleMethodNotAllowed(_req: IncomingMessage, res: ServerResponse): void {
  sendError(res, 405, { code: 1000, message: '方法不允许' });
}
