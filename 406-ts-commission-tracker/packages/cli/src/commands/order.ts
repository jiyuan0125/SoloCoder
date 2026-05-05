import { apiClient } from '../client';
import {
  Order,
  SalesContribution,
} from '@commission-tracker/shared';
import {
  formatOrder,
  formatOrderList,
} from '../utils/formatter';

export async function handleCreateOrder(args: string[]): Promise<void> {
  if (args.length < 3) {
    console.log('用法: commission order create <金额(分)> <日期> <销售贡献...>');
    console.log('销售贡献格式: <销售ID>[:primary]');
    console.log('示例: commission order create 1000000 "2026-05-01" SP-xxx:primary SP-yyy');
    return;
  }

  const amount = parseInt(args[0], 10);
  const date = args[1];
  const contributionArgs = args.slice(2);

  if (isNaN(amount) || amount <= 0) {
    console.error('金额必须是大于0的整数(单位:分)');
    return;
  }

  const salesContributions: SalesContribution[] = contributionArgs.map(arg => {
    const parts = arg.split(':');
    return {
      salespersonId: parts[0],
      isPrimary: parts[1] === 'primary',
    };
  });

  const response = await apiClient.post('/api/orders', {
    amount,
    date,
    salesContributions,
  });

  if (!response.success) {
    console.error('创建订单失败:', response.error?.message || '未知错误');
    return;
  }

  const data = response.data as Order;
  console.log('订单创建成功:');
  console.log(formatOrder(data));
}

export async function handleGetOrder(args: string[]): Promise<void> {
  if (args.length < 1) {
    console.log('用法: commission order get <订单ID>');
    return;
  }

  const id = args[0];
  const response = await apiClient.get(`/api/orders/${id}`);

  if (!response.success) {
    console.error('获取订单失败:', response.error?.message || '未知错误');
    return;
  }

  const data = response.data as Order;
  console.log(formatOrder(data));
}

export async function handleListOrders(args: string[]): Promise<void> {
  let path = '/api/orders';
  const params: string[] = [];

  if (args.length >= 1) {
    params.push(`salespersonId=${args[0]}`);
  }
  if (args.length >= 2) {
    params.push(`fromDate=${args[1]}`);
  }
  if (args.length >= 3) {
    params.push(`toDate=${args[2]}`);
  }

  if (params.length > 0) {
    path += `?${params.join('&')}`;
  }

  const response = await apiClient.get(path);

  if (!response.success) {
    console.error('获取订单列表失败:', response.error?.message || '未知错误');
    return;
  }

  const data = response.data as Order[];
  console.log(formatOrderList(data));
}

export async function handleRefundOrder(args: string[]): Promise<void> {
  if (args.length < 1) {
    console.log('用法: commission order refund <订单ID>');
    return;
  }

  const id = args[0];
  const response = await apiClient.post(`/api/orders/${id}/refund`);

  if (!response.success) {
    console.error('订单退款失败:', response.error?.message || '未知错误');
    return;
  }

  const data = response.data as Order;
  console.log('订单退款成功:');
  console.log(formatOrder(data));
}
