import { apiClient } from '../api-client';
import {
  formatDate,
  formatMoneyDisplay,
  printSection,
  printKeyValue,
  printSuccess,
  formatPlanType,
} from '../formatter';
import type {
  RecordUsageResponse,
  GetUsageResponse,
  PlanType,
} from '@billing/shared';

export async function handleUsageCommand(args: string[]): Promise<void> {
  if (args.length === 0) {
    printUsageHelp();
    return;
  }

  const command = args[0];

  switch (command) {
    case 'record':
      await handleRecordUsage(args.slice(1));
      break;
    case 'query':
      await handleQueryUsage(args.slice(1));
      break;
    default:
      printUsageHelp();
  }
}

async function handleRecordUsage(args: string[]): Promise<void> {
  const customerId = args[0];
  const units = parseInt(args[1] || '0', 10);

  if (!customerId || !args[1] || isNaN(units) || units <= 0) {
    console.error('用法: billing usage record <customerId> <units>');
    console.error('  units: 用量单位数 (正整数)');
    process.exit(1);
  }

  try {
    const result = await apiClient.post<RecordUsageResponse>('/usage', {
      customerId,
      units,
    });
    if (result.isLimited) {
      console.warn('⚠️ 用量已超出免费版配额，已被限流');
    }
    printSuccess('用量记录成功');
    printSection('用量信息');
    printKeyValue('记录ID', result.id);
    printKeyValue('客户ID', result.customerId);
    printKeyValue('用量单位', `${result.units} 单位`);
    printKeyValue('记录时间', formatDate(result.recordedAt));
    if (result.isLimited) {
      printKeyValue('是否限流', '是');
    }
  } catch (error) {
    throw error;
  }
}

async function handleQueryUsage(args: string[]): Promise<void> {
  const customerId = args[0];

  if (!customerId) {
    console.error('用法: billing usage query <customerId> [periodStart] [periodEnd]');
    process.exit(1);
  }

  try {
    let path = `/usage/${customerId}`;
    const periodStart = args[1];
    const periodEnd = args[2];

    if (periodStart && periodEnd) {
      path += `?periodStart=${encodeURIComponent(periodStart)}&periodEnd=${encodeURIComponent(periodEnd)}`;
    }

    const result = await apiClient.get<GetUsageResponse>(path);
    printSection('用量统计');
    printKeyValue('客户ID', result.customerId);
    printKeyValue('套餐', formatPlanType(result.planType as PlanType));
    printKeyValue('月度配额', `${result.monthlyQuota.toLocaleString()} 单位`);
    printKeyValue('统计周期', `${formatDate(result.periodStart)} 至 ${formatDate(result.periodEnd)}`);
    console.log('');
    printKeyValue('总用量', `${result.totalUnits.toLocaleString()} 单位`);
    printKeyValue('已用配额', `${result.quotaUsed.toLocaleString()} 单位`);
    printKeyValue('超额用量', `${result.overageUnits.toLocaleString()} 单位`);
    if (result.overageFee > 0) {
      printKeyValue('超额费用', formatMoneyDisplay(result.overageFee));
    }

    const remaining = result.monthlyQuota - result.quotaUsed;
    const percentage = Math.round((result.quotaUsed / result.monthlyQuota) * 100);
    console.log('');
    printKeyValue('剩余配额', `${remaining.toLocaleString()} 单位`);
    printKeyValue('使用比例', `${percentage}%`);
  } catch (error) {
    throw error;
  }
}

function printUsageHelp(): void {
  console.log(`
用法: billing usage <command> [options]

用量管理命令:
  record <customerId> <units>
                          记录用量
  query <customerId> [periodStart] [periodEnd]
                          查询用量统计 (默认查询本月)

示例:
  billing usage record cus-abc123 1500
  billing usage query cus-abc123
  billing usage query cus-abc123 "2024-01-01T00:00:00Z" "2024-01-31T23:59:59Z"
`);
}
