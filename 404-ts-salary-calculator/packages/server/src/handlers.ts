import { IncomingMessage, ServerResponse } from 'http';
import { parseBody, sendJson, createErrorResponse, createSuccessResponse } from './server.js';
import { calculateSalary, validateInput } from './calculator.js';
import { saveRecord, queryRecords, getComparison } from './storage.js';
import {
  ErrorCode,
  SalaryCalculationInput,
  SalaryCalculationResult,
} from '@salary/shared';

function getQueryParams(url: string): URLSearchParams {
  const queryStart = url.indexOf('?');
  if (queryStart === -1) {
    return new URLSearchParams();
  }
  return new URLSearchParams(url.slice(queryStart + 1));
}

export async function handleHealth(req: IncomingMessage, res: ServerResponse): Promise<void> {
  sendJson(
    res,
    200,
    createSuccessResponse({
      status: 'ok',
      timestamp: Date.now(),
    })
  );
}

export async function handleCalculate(req: IncomingMessage, res: ServerResponse): Promise<void> {
  const body = (await parseBody(req)) as Record<string, unknown>;
  
  const input: SalaryCalculationInput = {
    baseSalary: (body.baseSalary as number) ?? 0,
    positionAllowance: (body.positionAllowance as number) ?? 0,
    performanceCoefficient: (body.performanceCoefficient as number) ?? 0,
    socialInsuranceRate: (body.socialInsuranceRate as number) ?? 0,
    housingFundRate: (body.housingFundRate as number) ?? 0,
    taxThreshold: (body.taxThreshold as number) ?? 0,
    leaveDeduction: (body.leaveDeduction as number) ?? 0,
    overtimeHours: {
      weekday: (body.overtimeHours as { weekday?: number })?.weekday ?? 0,
      weekend: (body.overtimeHours as { weekend?: number })?.weekend ?? 0,
      holiday: (body.overtimeHours as { holiday?: number })?.holiday ?? 0,
    },
    yearEndBonus: (body.yearEndBonus as number) ?? 0,
    employeeId: (body.employeeId as string) ?? '',
    employeeName: (body.employeeName as string) ?? '',
    month: (body.month as string) ?? '',
  };
  
  const validationError = validateInput(input);
  if (validationError) {
    sendJson(res, 400, createErrorResponse(ErrorCode.INVALID_INPUT, validationError));
    return;
  }
  
  const result = calculateSalary(input);
  
  sendJson(res, 200, createSuccessResponse(result));
}

export async function handleSave(req: IncomingMessage, res: ServerResponse): Promise<void> {
  const body = (await parseBody(req)) as SalaryCalculationResult;
  
  if (!body.employeeId || !body.month) {
    sendJson(
      res,
      400,
      createErrorResponse(ErrorCode.INVALID_INPUT, '缺少必要字段: employeeId 或 month')
    );
    return;
  }
  
  const record = saveRecord(body);
  
  sendJson(res, 201, createSuccessResponse(record));
}

export async function handleQuery(req: IncomingMessage, res: ServerResponse): Promise<void> {
  const params = getQueryParams(req.url || '');
  const employeeId = params.get('employeeId') || undefined;
  const month = params.get('month') || undefined;
  
  const records = queryRecords(employeeId, month);
  
  sendJson(res, 200, createSuccessResponse(records));
}

export async function handleComparison(req: IncomingMessage, res: ServerResponse): Promise<void> {
  const params = getQueryParams(req.url || '');
  const employeeId = params.get('employeeId');
  const month = params.get('month');
  
  if (!employeeId || !month) {
    sendJson(
      res,
      400,
      createErrorResponse(ErrorCode.INVALID_INPUT, '缺少必要查询参数: employeeId 和 month')
    );
    return;
  }
  
  if (!/^\d{4}-\d{2}$/.test(month)) {
    sendJson(
      res,
      400,
      createErrorResponse(ErrorCode.INVALID_MONTH_FORMAT, '月份格式无效，应为 YYYY-MM')
    );
    return;
  }
  
  const comparison = getComparison(employeeId, month);
  
  if (!comparison) {
    sendJson(
      res,
      404,
      createErrorResponse(ErrorCode.RECORD_NOT_FOUND, `未找到员工 ${employeeId} 在 ${month} 的薪资记录`)
    );
    return;
  }
  
  sendJson(res, 200, createSuccessResponse(comparison));
}
