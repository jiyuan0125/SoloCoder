import { apiClient } from '../client';
import {
  formatSettlement,
  formatSettlementList,
} from '../utils/formatter';

export async function handleTriggerSettlement(args: string[]): Promise<void> {
  const body: Record<string, number> = {};
  
  if (args.length >= 1) {
    body.month = parseInt(args[0], 10);
  }
  if (args.length >= 2) {
    body.year = parseInt(args[1], 10);
  }

  const response = await apiClient.post('/api/settlements/trigger', body);

  if (!response.success) {
    console.error('触发结算失败:', response.error?.message || '未知错误');
    return;
  }

  const data = response.data as any;
  console.log(`结算完成，共 ${data.settlements.length} 笔结算记录`);
  console.log(`总结算金额: ¥${(data.totalSettled / 100).toFixed(2)}`);
  console.log('');
  console.log(formatSettlementList(data.settlements));
}

export async function handleListSettlements(args: string[]): Promise<void> {
  let path = '/api/settlements';
  const params: string[] = [];

  if (args.length >= 1) {
    params.push(`salespersonId=${args[0]}`);
  }
  if (args.length >= 2) {
    params.push(`year=${args[1]}`);
  }
  if (args.length >= 3) {
    params.push(`month=${args[2]}`);
  }

  if (params.length > 0) {
    path += `?${params.join('&')}`;
  }

  const response = await apiClient.get(path);

  if (!response.success) {
    console.error('获取结算记录失败:', response.error?.message || '未知错误');
    return;
  }

  console.log(formatSettlementList(response.data as any[]));
}

export async function handleGetSettlement(args: string[]): Promise<void> {
  if (args.length < 1) {
    console.log('用法: commission settlement get <结算单ID>');
    return;
  }

  const id = args[0];
  const response = await apiClient.get(`/api/settlements/${id}`);

  if (!response.success) {
    console.error('获取结算单失败:', response.error?.message || '未知错误');
    return;
  }

  console.log(formatSettlement(response.data as any));
}
