#!/usr/bin/env node

import { apiClient } from './api';
import {
  formatExpense,
  formatExpenseList,
  formatError,
  formatUser,
  formatHelp,
  formatOperationLogList,
  formatAuditLogList,
} from './format';

interface ParsedArgs {
  command: string | null;
  globalOptions: {
    user?: string;
    help: boolean;
  };
  options: Record<string, string | boolean | number>;
}

function parseArgs(args: string[]): ParsedArgs {
  const result: ParsedArgs = {
    command: null,
    globalOptions: {
      help: false,
    },
    options: {},
  };

  let i = 0;
  while (i < args.length) {
    const arg = args[i];
    if (arg === undefined) {
      i++;
      continue;
    }

    if (arg === '--help' || arg === '-h') {
      result.globalOptions.help = true;
      i++;
      continue;
    }

    if (arg === '--user') {
      if (i + 1 < args.length) {
        const userId = args[i + 1];
        if (userId !== undefined) {
          result.globalOptions.user = userId;
        }
        i += 2;
        continue;
      }
    }

    if (!result.command && !arg.startsWith('-')) {
      result.command = arg;
      i++;
      continue;
    }

    if (arg.startsWith('--')) {
      const key = arg.substring(2).replace(/-/g, '_');
      const nextArg = args[i + 1];
      if (i + 1 < args.length && nextArg !== undefined && !nextArg.startsWith('-')) {
        const value = nextArg;
        const numValue = parseFloat(value);
        if (!isNaN(numValue) && value === numValue.toString()) {
          result.options[key] = numValue;
        } else {
          result.options[key] = value;
        }
        i += 2;
      } else {
        result.options[key] = true;
        i++;
      }
      continue;
    }

    i++;
  }

  return result;
}

function parseAmountToCents(amountStr: string): number | null {
  const trimmed = amountStr.trim();
  const match = trimmed.match(/^(\d+)(?:\.(\d{1,2}))?$/);
  if (!match) {
    return null;
  }
  const yuan = parseInt(match[1] ?? '0', 10);
  const fenStr = (match[2] ?? '0').padEnd(2, '0');
  const fen = parseInt(fenStr, 10);
  return yuan * 100 + fen;
}

function getOptionAsString(options: Record<string, string | boolean | number>, key: string): string | undefined {
  const value = options[key];
  if (typeof value === 'string') {
    return value;
  }
  return undefined;
}

function getOptionAsNumber(options: Record<string, string | boolean | number>, key: string): number | undefined {
  const value = options[key];
  if (typeof value === 'number') {
    return value;
  }
  if (typeof value === 'string') {
    const num = parseFloat(value);
    if (!isNaN(num)) {
      return num;
    }
  }
  return undefined;
}

function getBoolean(obj: Record<string, unknown>, key: string): boolean {
  const value = obj[key];
  return value === true;
}

function getAny(obj: Record<string, unknown>, key: string): unknown {
  return obj[key];
}

async function handleCreate(options: Record<string, string | boolean | number>): Promise<void> {
  const employeeId = getOptionAsString(options, 'employeeId');
  const employeeName = getOptionAsString(options, 'employeeName');
  const date = getOptionAsString(options, 'date');
  const amountStr = getOptionAsString(options, 'amount');
  const category = getOptionAsString(options, 'category');
  const reason = getOptionAsString(options, 'reason');
  const voucher = getOptionAsString(options, 'voucher');

  if (!employeeId || !employeeName || !date || !amountStr || !category || !reason || !voucher) {
    console.error('错误: 缺少必要参数');
    console.log('用法: create --employeeId <id> --employeeName <name> --date <YYYY-MM-DD> --amount <yuan> --category <cat> --reason <text> --voucher <file>');
    process.exit(1);
  }

  const amountCents = parseAmountToCents(amountStr);
  if (amountCents === null) {
    console.error(`错误: 金额格式无效: ${amountStr}`);
    console.log('金额格式: 支持小数，如 100.50 表示100元50分');
    process.exit(1);
  }

  try {
    const response = await apiClient.createExpense({
      employeeId,
      employeeName,
      date,
      amount: amountCents,
      category,
      reason,
      voucherFileName: voucher,
    });

    const result = response as Record<string, unknown>;
    if (getBoolean(result, 'success') && getAny(result, 'data')) {
      console.log('✅ 报销单创建成功!');
      console.log(formatExpense(getAny(result, 'data') as Record<string, unknown>));
    } else if (getAny(result, 'error')) {
      console.log(formatError(getAny(result, 'error') as { code: number; message: string; details?: string }));
      process.exit(1);
    }
  } catch (err) {
    console.error('连接服务器失败，请确保服务器已启动');
    if (err instanceof Error) {
      console.error(`错误: ${err.message}`);
    }
    process.exit(1);
  }
}

async function handleUpdate(options: Record<string, string | boolean | number>): Promise<void> {
  const expenseId = getOptionAsString(options, 'id');
  if (!expenseId) {
    console.error('错误: 缺少报销单ID');
    console.log('用法: update --id <expenseId> [--date <date>] [--amount <yuan>] [--category <cat>] [--reason <text>] [--voucher <file>] [--supplementary <text>]');
    process.exit(1);
  }

  const params: Record<string, unknown> = { expenseId };

  const date = getOptionAsString(options, 'date');
  if (date) params['date'] = date;

  const amountStr = getOptionAsString(options, 'amount');
  if (amountStr) {
    const amountCents = parseAmountToCents(amountStr);
    if (amountCents === null) {
      console.error(`错误: 金额格式无效: ${amountStr}`);
      process.exit(1);
    }
    params['amount'] = amountCents;
  }

  const category = getOptionAsString(options, 'category');
  if (category) params['category'] = category;

  const reason = getOptionAsString(options, 'reason');
  if (reason) params['reason'] = reason;

  const voucher = getOptionAsString(options, 'voucher');
  if (voucher) params['voucherFileName'] = voucher;

  const supplementary = getOptionAsString(options, 'supplementary');
  if (supplementary) params['supplementaryInfo'] = supplementary;

  try {
    const response = await apiClient.updateExpense(params as {
      expenseId: string;
      date?: string;
      amount?: number;
      category?: string;
      reason?: string;
      voucherFileName?: string;
      supplementaryInfo?: string;
    });

    const result = response as Record<string, unknown>;
    if (getBoolean(result, 'success') && getAny(result, 'data')) {
      console.log('✅ 报销单更新成功!');
      console.log(formatExpense(getAny(result, 'data') as Record<string, unknown>));
    } else if (getAny(result, 'error')) {
      console.log(formatError(getAny(result, 'error') as { code: number; message: string; details?: string }));
      process.exit(1);
    }
  } catch (err) {
    console.error('连接服务器失败，请确保服务器已启动');
    if (err instanceof Error) {
      console.error(`错误: ${err.message}`);
    }
    process.exit(1);
  }
}

async function handleSubmit(options: Record<string, string | boolean | number>): Promise<void> {
  const expenseId = getOptionAsString(options, 'id');
  if (!expenseId) {
    console.error('错误: 缺少报销单ID');
    console.log('用法: submit --id <expenseId>');
    process.exit(1);
  }

  try {
    const response = await apiClient.submitExpense(expenseId);

    const result = response as Record<string, unknown>;
    if (getBoolean(result, 'success') && getAny(result, 'data')) {
      console.log('✅ 报销单提交成功!');
      console.log(formatExpense(getAny(result, 'data') as Record<string, unknown>));
    } else if (getAny(result, 'error')) {
      console.log(formatError(getAny(result, 'error') as { code: number; message: string; details?: string }));
      process.exit(1);
    }
  } catch (err) {
    console.error('连接服务器失败，请确保服务器已启动');
    if (err instanceof Error) {
      console.error(`错误: ${err.message}`);
    }
    process.exit(1);
  }
}

async function handleApprove(options: Record<string, string | boolean | number>): Promise<void> {
  const expenseId = getOptionAsString(options, 'id');
  if (!expenseId) {
    console.error('错误: 缺少报销单ID');
    console.log('用法: approve --id <expenseId> [--comment <text>]');
    process.exit(1);
  }

  const comment = getOptionAsString(options, 'comment');

  try {
    const approveParams: {
      expenseId: string;
      comment?: string;
    } = {
      expenseId,
    };
    if (comment !== undefined) {
      approveParams.comment = comment;
    }
    const response = await apiClient.approveExpense(approveParams);

    const result = response as Record<string, unknown>;
    if (getBoolean(result, 'success') && getAny(result, 'data')) {
      if (getBoolean(result, 'needsSecondLevel')) {
        console.log('✅ 部门经理审批通过! 需等待财务总监二级审批');
      } else {
        console.log('✅ 审批通过!');
      }
      console.log(formatExpense(getAny(result, 'data') as Record<string, unknown>));
    } else if (getAny(result, 'error')) {
      console.log(formatError(getAny(result, 'error') as { code: number; message: string; details?: string }));
      process.exit(1);
    }
  } catch (err) {
    console.error('连接服务器失败，请确保服务器已启动');
    if (err instanceof Error) {
      console.error(`错误: ${err.message}`);
    }
    process.exit(1);
  }
}

async function handleReject(options: Record<string, string | boolean | number>): Promise<void> {
  const expenseId = getOptionAsString(options, 'id');
  if (!expenseId) {
    console.error('错误: 缺少报销单ID');
    console.log('用法: reject --id <expenseId> [--comment <text>]');
    process.exit(1);
  }

  const comment = getOptionAsString(options, 'comment');

  try {
    const rejectParams: {
      expenseId: string;
      comment?: string;
    } = {
      expenseId,
    };
    if (comment !== undefined) {
      rejectParams.comment = comment;
    }
    const response = await apiClient.rejectExpense(rejectParams);

    const result = response as Record<string, unknown>;
    if (getBoolean(result, 'success') && getAny(result, 'data')) {
      console.log('✅ 报销单已驳回!');
      console.log(formatExpense(getAny(result, 'data') as Record<string, unknown>));
    } else if (getAny(result, 'error')) {
      console.log(formatError(getAny(result, 'error') as { code: number; message: string; details?: string }));
      process.exit(1);
    }
  } catch (err) {
    console.error('连接服务器失败，请确保服务器已启动');
    if (err instanceof Error) {
      console.error(`错误: ${err.message}`);
    }
    process.exit(1);
  }
}

async function handleBatchApprove(options: Record<string, string | boolean | number>): Promise<void> {
  const idsStr = getOptionAsString(options, 'ids');
  if (!idsStr) {
    console.error('错误: 缺少报销单ID列表');
    console.log('用法: batch-approve --ids <id1,id2,id3> [--comment <text>]');
    process.exit(1);
  }

  const expenseIds = idsStr.split(',').map(id => id.trim()).filter(id => id.length > 0);
  if (expenseIds.length === 0) {
    console.error('错误: 报销单ID列表为空');
    process.exit(1);
  }

  const comment = getOptionAsString(options, 'comment');

  try {
    const batchParams: {
      expenseIds: string[];
      comment?: string;
    } = {
      expenseIds,
    };
    if (comment !== undefined) {
      batchParams.comment = comment;
    }
    const response = await apiClient.batchApprove(batchParams);

    const result = response as Record<string, unknown>;
    if (getBoolean(result, 'success')) {
      const data = getAny(result, 'data') as Record<string, unknown>;
      const approved = (getAny(data, 'approved') as string[]) || [];
      const skipped = (getAny(data, 'skipped') as string[]) || [];
      const failed = (getAny(data, 'failed') as Array<Record<string, unknown>>) || [];

      console.log('');
      console.log('批量审批结果:');
      console.log('');
      console.log(`  ✅ 成功审批: ${approved.length} 笔`);
      if (approved.length > 0) {
        console.log(`     ${approved.join(', ')}`);
      }
      console.log(`  ⏭️  自动跳过: ${skipped.length} 笔 (大额报销需单独审批)`);
      if (skipped.length > 0) {
        console.log(`     ${skipped.join(', ')}`);
      }
      console.log(`  ❌ 失败: ${failed.length} 笔`);
      if (failed.length > 0) {
        for (const f of failed) {
          const fExpenseId = getAny(f, 'expenseId');
          const fMessage = getAny(f, 'message');
          console.log(`     ${fExpenseId}: ${fMessage}`);
        }
      }
      console.log('');
    } else if (getAny(result, 'error')) {
      console.log(formatError(getAny(result, 'error') as { code: number; message: string; details?: string }));
      process.exit(1);
    }
  } catch (err) {
    console.error('连接服务器失败，请确保服务器已启动');
    if (err instanceof Error) {
      console.error(`错误: ${err.message}`);
    }
    process.exit(1);
  }
}

async function handlePay(options: Record<string, string | boolean | number>): Promise<void> {
  const expenseId = getOptionAsString(options, 'id');
  if (!expenseId) {
    console.error('错误: 缺少报销单ID');
    console.log('用法: pay --id <expenseId>');
    process.exit(1);
  }

  try {
    const response = await apiClient.payExpense(expenseId);

    const result = response as Record<string, unknown>;
    if (getBoolean(result, 'success') && getAny(result, 'data')) {
      console.log('✅ 打款确认成功!');
      console.log(formatExpense(getAny(result, 'data') as Record<string, unknown>));
    } else if (getAny(result, 'error')) {
      console.log(formatError(getAny(result, 'error') as { code: number; message: string; details?: string }));
      process.exit(1);
    }
  } catch (err) {
    console.error('连接服务器失败，请确保服务器已启动');
    if (err instanceof Error) {
      console.error(`错误: ${err.message}`);
    }
    process.exit(1);
  }
}

async function handleResubmit(options: Record<string, string | boolean | number>): Promise<void> {
  const expenseId = getOptionAsString(options, 'id');
  if (!expenseId) {
    console.error('错误: 缺少报销单ID');
    console.log('用法: resubmit --id <expenseId>');
    process.exit(1);
  }

  try {
    const response = await apiClient.resubmitExpense(expenseId);

    const result = response as Record<string, unknown>;
    if (getBoolean(result, 'success') && getAny(result, 'data')) {
      console.log('✅ 报销单重新提交成功!');
      console.log(formatExpense(getAny(result, 'data') as Record<string, unknown>));
    } else if (getAny(result, 'error')) {
      console.log(formatError(getAny(result, 'error') as { code: number; message: string; details?: string }));
      process.exit(1);
    }
  } catch (err) {
    console.error('连接服务器失败，请确保服务器已启动');
    if (err instanceof Error) {
      console.error(`错误: ${err.message}`);
    }
    process.exit(1);
  }
}

async function handleDelete(options: Record<string, string | boolean | number>): Promise<void> {
  const expenseId = getOptionAsString(options, 'id');
  if (!expenseId) {
    console.error('错误: 缺少报销单ID');
    console.log('用法: delete --id <expenseId>');
    process.exit(1);
  }

  try {
    const response = await apiClient.deleteExpense(expenseId);

    const result = response as Record<string, unknown>;
    if (getBoolean(result, 'success')) {
      console.log('✅ 报销单删除成功!');
    } else if (getAny(result, 'error')) {
      console.log(formatError(getAny(result, 'error') as { code: number; message: string; details?: string }));
      process.exit(1);
    }
  } catch (err) {
    console.error('连接服务器失败，请确保服务器已启动');
    if (err instanceof Error) {
      console.error(`错误: ${err.message}`);
    }
    process.exit(1);
  }
}

async function handleGet(options: Record<string, string | boolean | number>): Promise<void> {
  const expenseId = getOptionAsString(options, 'id');
  if (!expenseId) {
    console.error('错误: 缺少报销单ID');
    console.log('用法: get --id <expenseId>');
    process.exit(1);
  }

  try {
    const response = await apiClient.getExpense(expenseId);

    const result = response as Record<string, unknown>;
    if (getBoolean(result, 'success') && getAny(result, 'data')) {
      console.log(formatExpense(getAny(result, 'data') as Record<string, unknown>));
    } else if (getAny(result, 'error')) {
      console.log(formatError(getAny(result, 'error') as { code: number; message: string; details?: string }));
      process.exit(1);
    }
  } catch (err) {
    console.error('连接服务器失败，请确保服务器已启动');
    if (err instanceof Error) {
      console.error(`错误: ${err.message}`);
    }
    process.exit(1);
  }
}

async function handleList(options: Record<string, string | boolean | number>): Promise<void> {
  const params: Record<string, unknown> = {};

  const startDate = getOptionAsString(options, 'start_date');
  if (startDate) params['startDate'] = startDate;

  const endDate = getOptionAsString(options, 'end_date');
  if (endDate) params['endDate'] = endDate;

  const status = getOptionAsString(options, 'status');
  if (status) params['status'] = status;

  const employeeId = getOptionAsString(options, 'employee_id');
  if (employeeId) params['employeeId'] = employeeId;

  const page = getOptionAsNumber(options, 'page');
  if (page !== undefined) params['page'] = page;

  const pageSize = getOptionAsNumber(options, 'page_size');
  if (pageSize !== undefined) params['pageSize'] = pageSize;

  try {
    const response = await apiClient.listExpenses(params as {
      startDate?: string;
      endDate?: string;
      status?: string;
      employeeId?: string;
      page?: number;
      pageSize?: number;
    });

    const result = response as Record<string, unknown>;
    if (getBoolean(result, 'success') && getAny(result, 'data')) {
      const data = getAny(result, 'data') as Record<string, unknown>;
      const items = (getAny(data, 'items') as Array<Record<string, unknown>>) || [];
      const total = getAny(data, 'total') as number;
      const pageNum = getAny(data, 'page') as number;
      const size = getAny(data, 'pageSize') as number;

      console.log('');
      console.log(`共 ${total} 条记录 (第 ${pageNum} 页，每页 ${size} 条)`);
      console.log('');
      console.log(formatExpenseList(items));
      console.log('');
    } else if (getAny(result, 'error')) {
      console.log(formatError(getAny(result, 'error') as { code: number; message: string; details?: string }));
      process.exit(1);
    }
  } catch (err) {
    console.error('连接服务器失败，请确保服务器已启动');
    if (err instanceof Error) {
      console.error(`错误: ${err.message}`);
    }
    process.exit(1);
  }
}

async function handleLogs(options: Record<string, string | boolean | number>): Promise<void> {
  const params: Record<string, unknown> = {};

  const startDate = getOptionAsString(options, 'start_date');
  if (startDate) params['startDate'] = startDate;

  const endDate = getOptionAsString(options, 'end_date');
  if (endDate) params['endDate'] = endDate;

  const userId = getOptionAsString(options, 'user_id');
  if (userId) params['userId'] = userId;

  const page = getOptionAsNumber(options, 'page');
  if (page !== undefined) params['page'] = page;

  try {
    const response = await apiClient.getOperationLogs(params as {
      startDate?: string;
      endDate?: string;
      userId?: string;
      page?: number;
      pageSize?: number;
    });

    const result = response as Record<string, unknown>;
    if (getBoolean(result, 'success') && getAny(result, 'data')) {
      const data = getAny(result, 'data') as Record<string, unknown>;
      const items = (getAny(data, 'items') as Array<Record<string, unknown>>) || [];
      const total = getAny(data, 'total') as number;

      console.log('');
      console.log(`共 ${total} 条操作日志`);
      console.log('');
      console.log(formatOperationLogList(items));
      console.log('');
    } else if (getAny(result, 'error')) {
      console.log(formatError(getAny(result, 'error') as { code: number; message: string; details?: string }));
      process.exit(1);
    }
  } catch (err) {
    console.error('连接服务器失败，请确保服务器已启动');
    if (err instanceof Error) {
      console.error(`错误: ${err.message}`);
    }
    process.exit(1);
  }
}

async function handleAudit(options: Record<string, string | boolean | number>): Promise<void> {
  const params: Record<string, unknown> = {};

  const startDate = getOptionAsString(options, 'start_date');
  if (startDate) params['startDate'] = startDate;

  const endDate = getOptionAsString(options, 'end_date');
  if (endDate) params['endDate'] = endDate;

  const userId = getOptionAsString(options, 'user_id');
  if (userId) params['userId'] = userId;

  const action = getOptionAsString(options, 'action');
  if (action) params['action'] = action;

  const page = getOptionAsNumber(options, 'page');
  if (page !== undefined) params['page'] = page;

  try {
    const response = await apiClient.getAuditLogs(params as {
      startDate?: string;
      endDate?: string;
      userId?: string;
      action?: string;
      page?: number;
      pageSize?: number;
    });

    const result = response as Record<string, unknown>;
    if (getBoolean(result, 'success') && getAny(result, 'data')) {
      const data = getAny(result, 'data') as Record<string, unknown>;
      const items = (getAny(data, 'items') as Array<Record<string, unknown>>) || [];
      const total = getAny(data, 'total') as number;

      console.log('');
      console.log(`共 ${total} 条审计日志`);
      console.log('');
      console.log(formatAuditLogList(items));
      console.log('');
    } else if (getAny(result, 'error')) {
      console.log(formatError(getAny(result, 'error') as { code: number; message: string; details?: string }));
      process.exit(1);
    }
  } catch (err) {
    console.error('连接服务器失败，请确保服务器已启动');
    if (err instanceof Error) {
      console.error(`错误: ${err.message}`);
    }
    process.exit(1);
  }
}

async function handleUser(options: Record<string, string | boolean | number>): Promise<void> {
  const userId = getOptionAsString(options, 'id');
  if (!userId) {
    console.error('错误: 缺少用户ID');
    console.log('用法: user --id <userId>');
    process.exit(1);
  }

  try {
    const response = await apiClient.getUser(userId);

    const result = response as Record<string, unknown>;
    if (getBoolean(result, 'success') && getAny(result, 'data')) {
      console.log(formatUser(getAny(result, 'data') as Record<string, unknown>));
    } else if (getAny(result, 'error')) {
      console.log(formatError(getAny(result, 'error') as { code: number; message: string; details?: string }));
      process.exit(1);
    }
  } catch (err) {
    console.error('连接服务器失败，请确保服务器已启动');
    if (err instanceof Error) {
      console.error(`错误: ${err.message}`);
    }
    process.exit(1);
  }
}

async function main(): Promise<void> {
  const args = process.argv.slice(2);
  const parsed = parseArgs(args);

  if (parsed.globalOptions.user) {
    apiClient.setUserId(parsed.globalOptions.user);
  }

  if (parsed.globalOptions.help || !parsed.command) {
    console.log(formatHelp());
    return;
  }

  const command = parsed.command;

  switch (command) {
    case 'create':
      await handleCreate(parsed.options);
      break;
    case 'update':
      await handleUpdate(parsed.options);
      break;
    case 'submit':
      await handleSubmit(parsed.options);
      break;
    case 'approve':
      await handleApprove(parsed.options);
      break;
    case 'reject':
      await handleReject(parsed.options);
      break;
    case 'batch-approve':
      await handleBatchApprove(parsed.options);
      break;
    case 'pay':
      await handlePay(parsed.options);
      break;
    case 'resubmit':
      await handleResubmit(parsed.options);
      break;
    case 'delete':
      await handleDelete(parsed.options);
      break;
    case 'get':
      await handleGet(parsed.options);
      break;
    case 'list':
      await handleList(parsed.options);
      break;
    case 'logs':
      await handleLogs(parsed.options);
      break;
    case 'audit':
      await handleAudit(parsed.options);
      break;
    case 'user':
      await handleUser(parsed.options);
      break;
    default:
      console.error(`未知命令: ${command}`);
      console.log(formatHelp());
      process.exit(1);
  }
}

main().catch((err) => {
  console.error('执行出错:', err);
  process.exit(1);
});
