import { formatMoney } from '@billing/shared';
import type {
  BillStatus,
  CustomerStatus,
  PlanType,
  BillingCycle,
  PaymentStatus,
  RefundStatus,
} from '@billing/shared';

export { formatMoney };

export type AllStatus = CustomerStatus | BillStatus | PaymentStatus | RefundStatus;

export const STATUS_COLORS: Record<string, string> = {
  active: '\x1b[32m',
  frozen: '\x1b[31m',
  trial: '\x1b[36m',
  pending: '\x1b[33m',
  paid: '\x1b[32m',
  overdue: '\x1b[31m',
  reviewing: '\x1b[35m',
  reviewed: '\x1b[36m',
  draft: '\x1b[90m',
  completed: '\x1b[32m',
  failed: '\x1b[31m',
  approved: '\x1b[32m',
  rejected: '\x1b[31m',
};

export const RESET = '\x1b[0m';
export const BOLD = '\x1b[1m';
export const DIM = '\x1b[2m';

export function colorize(text: string, status: string): string {
  const color = STATUS_COLORS[status] || '';
  return `${color}${text}${RESET}`;
}

const STATUS_NAMES: Record<string, string> = {
  active: '活跃',
  frozen: '冻结',
  trial: '试用',
  pending: '待支付',
  paid: '已支付',
  overdue: '已逾期',
  reviewing: '复核中',
  reviewed: '已复核',
  draft: '草稿',
  completed: '已完成',
  failed: '失败',
  approved: '已批准',
  rejected: '已拒绝',
};

export function formatStatus(status: AllStatus): string {
  const name = STATUS_NAMES[status] || status;
  return colorize(name.toUpperCase(), status);
}

export function formatPlanType(planType: PlanType): string {
  const names: Record<PlanType, string> = {
    free: '免费版',
    basic: '基础版',
    pro: '专业版',
  };
  return names[planType] || planType;
}

export function formatBillingCycle(cycle: BillingCycle): string {
  return cycle === 'monthly' ? '按月付费' : '按年付费';
}

export function formatMoneyDisplay(amount: number): string {
  return `¥${formatMoney(amount)}`;
}

export function formatDate(dateStr: string | null): string {
  if (!dateStr) {
    return '-';
  }
  const date = new Date(dateStr);
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  });
}

export function printTable(headers: string[], rows: string[][]): void {
  const colWidths = headers.map((h, i) => {
    const maxLen = Math.max(...rows.map((r) => (r[i] || '').length), h.length);
    return maxLen + 2;
  });

  const headerLine = headers.map((h, i) => h.padEnd((colWidths[i] ?? 0) - 2)).join('  ');
  console.log(`${BOLD}${headerLine}${RESET}`);
  console.log(DIM + '-'.repeat(colWidths.reduce((sum, w) => sum + w + 2, 0) - 2) + RESET);

  for (const row of rows) {
    const line = row.map((cell, i) => (cell || '').padEnd((colWidths[i] ?? 0) - 2)).join('  ');
    console.log(line);
  }
}

export function printSection(title: string): void {
  console.log(`\n${BOLD}=== ${title} ===${RESET}`);
}

export function printKeyValue(key: string, value: string): void {
  console.log(`${DIM}${key.padEnd(20)}${RESET}${value}`);
}

export function printError(message: string): void {
  console.error(`${STATUS_COLORS.frozen}错误: ${message}${RESET}`);
}

export function printSuccess(message: string): void {
  console.log(`${STATUS_COLORS.active}✓ ${message}${RESET}`);
}
