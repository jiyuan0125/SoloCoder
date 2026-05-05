import { ParsedArgs } from './args.js';
import { httpGet, httpPost } from './http.js';
import { formatSalaryResult, formatRecords, formatComparison } from './formatter.js';
import { API_ENDPOINTS } from '@salary/shared';
import { SalaryCalculationInput, SalaryCalculationResult, SalaryRecord, SalaryComparisonResult } from '@salary/shared';

let lastCalculationResult: SalaryCalculationResult | null = null;

function buildUrl(server: string, endpoint: string, params?: Record<string, string>): string {
  let url = server + endpoint;
  if (params && Object.keys(params).length > 0) {
    const searchParams = new URLSearchParams(params);
    url += '?' + searchParams.toString();
  }
  return url;
}

export async function handleCalculate(args: ParsedArgs): Promise<void> {
  const server = args.server as string;
  
  const input: SalaryCalculationInput = {
    employeeId: (args.employeeId as string) ?? '',
    employeeName: (args.employeeName as string) ?? '',
    month: (args.month as string) ?? '',
    baseSalary: (args.baseSalary as number) ?? 0,
    positionAllowance: (args.positionAllowance as number) ?? 0,
    performanceCoefficient: (args.performanceCoefficient as number) ?? 0,
    socialInsuranceRate: (args.socialInsuranceRate as number) ?? 0,
    housingFundRate: (args.housingFundRate as number) ?? 0,
    taxThreshold: (args.taxThreshold as number) ?? 0,
    leaveDeduction: (args.leaveDeduction as number) ?? 0,
    overtimeHours: {
      weekday: (args.overtimeWeekday as number) ?? 0,
      weekend: (args.overtimeWeekend as number) ?? 0,
      holiday: (args.overtimeHoliday as number) ?? 0,
    },
    yearEndBonus: (args.yearEndBonus as number) ?? 0,
  };
  
  const url = buildUrl(server, API_ENDPOINTS.CALCULATE);
  const response = await httpPost<SalaryCalculationResult>(url, input);
  
  if (!response.success || !response.data) {
    console.error('计算失败:', response.error?.message ?? '未知错误');
    process.exit(1);
  }
  
  lastCalculationResult = response.data;
  console.log(formatSalaryResult(response.data));
}

export async function handleSave(args: ParsedArgs): Promise<void> {
  const server = args.server as string;
  
  if (!lastCalculationResult) {
    console.error('没有可保存的计算结果。请先使用 calculate 命令计算薪资。');
    process.exit(1);
  }
  
  const url = buildUrl(server, API_ENDPOINTS.SAVE);
  const response = await httpPost<SalaryRecord>(url, lastCalculationResult);
  
  if (!response.success || !response.data) {
    console.error('保存失败:', response.error?.message ?? '未知错误');
    process.exit(1);
  }
  
  console.log('保存成功!');
  console.log('记录ID:', response.data.id);
}

export async function handleQuery(args: ParsedArgs): Promise<void> {
  const server = args.server as string;
  const params: Record<string, string> = {};
  
  if (args.employeeId) {
    params.employeeId = args.employeeId as string;
  }
  if (args.month) {
    params.month = args.month as string;
  }
  
  const url = buildUrl(server, API_ENDPOINTS.QUERY, params);
  const response = await httpGet<SalaryRecord[]>(url);
  
  if (!response.success || !response.data) {
    console.error('查询失败:', response.error?.message ?? '未知错误');
    process.exit(1);
  }
  
  console.log(formatRecords(response.data));
}

export async function handleComparison(args: ParsedArgs): Promise<void> {
  const server = args.server as string;
  const employeeId = args.employeeId as string;
  const month = args.month as string;
  
  if (!employeeId || !month) {
    console.error('请提供 --employeeId 和 --month 参数');
    process.exit(1);
  }
  
  const url = buildUrl(server, API_ENDPOINTS.COMPARISON, { employeeId, month });
  const response = await httpGet<SalaryComparisonResult>(url);
  
  if (!response.success || !response.data) {
    console.error('查询失败:', response.error?.message ?? '未知错误');
    process.exit(1);
  }
  
  console.log(formatComparison(response.data));
}
