import { apiClient } from '../client';
import {
  Salesperson,
  CommissionCalculation,
  BalanceInfo,
  ResignResult,
} from '@commission-tracker/shared';
import {
  formatSalesperson,
  formatSalespersonList,
} from '../utils/formatter';

export async function handleCreateSalesperson(args: string[]): Promise<void> {
  if (args.length < 2) {
    console.log('用法: commission salesperson create <姓名> <入职日期>');
    console.log('示例: commission salesperson create "张三" "2026-01-01"');
    return;
  }

  const name = args[0];
  const joinDate = args[1];

  const response = await apiClient.post('/api/salespeople', {
    name,
    joinDate,
  });

  if (!response.success) {
    console.error('创建销售失败:', response.error?.message || '未知错误');
    return;
  }

  const data = response.data as Salesperson;
  console.log('销售创建成功:');
  console.log(formatSalesperson(data));
}

export async function handleGetSalesperson(args: string[]): Promise<void> {
  if (args.length < 1) {
    console.log('用法: commission salesperson get <销售ID>');
    return;
  }

  const id = args[0];
  const response = await apiClient.get(`/api/salespeople/${id}`);

  if (!response.success) {
    console.error('获取销售失败:', response.error?.message || '未知错误');
    return;
  }

  const data = response.data as Salesperson;
  console.log(formatSalesperson(data));
}

export async function handleListSalespeople(): Promise<void> {
  const response = await apiClient.get('/api/salespeople');

  if (!response.success) {
    console.error('获取销售列表失败:', response.error?.message || '未知错误');
    return;
  }

  const data = response.data as Salesperson[];
  console.log(formatSalespersonList(data));
}

export async function handleResignSalesperson(args: string[]): Promise<void> {
  if (args.length < 1) {
    console.log('用法: commission salesperson resign <销售ID>');
    return;
  }

  const id = args[0];
  const response = await apiClient.post(`/api/salespeople/${id}/resign`);

  if (!response.success) {
    console.error('销售离职失败:', response.error?.message || '未知错误');
    return;
  }

  console.log('销售离职成功，已完成离职结算:');
  console.log(JSON.stringify(response.data, null, 2));
}

export async function handleGetBalance(args: string[]): Promise<void> {
  if (args.length < 1) {
    console.log('用法: commission salesperson balance <销售ID>');
    return;
  }

  const id = args[0];
  const response = await apiClient.get(`/api/salespeople/${id}/balance`);

  if (!response.success) {
    console.error('获取余额失败:', response.error?.message || '未知错误');
    return;
  }

  const data = response.data as BalanceInfo;
  console.log(`销售ID: ${data.salespersonId}`);
  console.log(`实时佣金余额: ${new Intl.NumberFormat('zh-CN', { style: 'currency', currency: 'CNY' }).format(data.balance / 100)}`);
  console.log(`本月已结算金额: ${new Intl.NumberFormat('zh-CN', { style: 'currency', currency: 'CNY' }).format(data.thisMonthSettled / 100)}`);
}

export async function handleGetCommission(args: string[]): Promise<void> {
  if (args.length < 1) {
    console.log('用法: commission salesperson commission <销售ID> [月份] [年份]');
    console.log('示例: commission salesperson commission SP-xxx 5 2026');
    return;
  }

  const id = args[0];
  let path = `/api/salespeople/${id}/commission`;
  
  const params: string[] = [];
  if (args.length >= 2) {
    params.push(`month=${args[1]}`);
  }
  if (args.length >= 3) {
    params.push(`year=${args[2]}`);
  }
  
  if (params.length > 0) {
    path += `?${params.join('&')}`;
  }

  const response = await apiClient.get(path);

  if (!response.success) {
    console.error('获取佣金计算失败:', response.error?.message || '未知错误');
    return;
  }

  const data = response.data as CommissionCalculation;
  console.log(JSON.stringify(data, null, 2));
}
