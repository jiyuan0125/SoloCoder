import { apiClient } from '../api-client';
import {
  formatStatus,
  formatDate,
  formatMoneyDisplay,
  printSection,
  printKeyValue,
  printSuccess,
  printTable,
} from '../formatter';
import type {
  GenerateBillResponse,
  GetBillsResponse,
  GetBillDetailResponse,
  PayBillResponse,
  RequestReviewResponse,
  BillStatus,
} from '@billing/shared';

export async function handleBillingCommand(args: string[]): Promise<void> {
  if (args.length === 0) {
    printBillingHelp();
    return;
  }

  const command = args[0];

  switch (command) {
    case 'generate':
      await handleGenerateBill(args.slice(1));
      break;
    case 'list':
      await handleListBills(args.slice(1));
      break;
    case 'show':
      await handleShowBill(args.slice(1));
      break;
    case 'pay':
      await handlePayBill(args.slice(1));
      break;
    case 'review':
      await handleRequestReview(args.slice(1));
      break;
    default:
      printBillingHelp();
  }
}

async function handleGenerateBill(args: string[]): Promise<void> {
  const customerId = args[0];
  const periodStart = args[1];
  const periodEnd = args[2];

  if (!customerId || !periodStart || !periodEnd) {
    console.error('用法: billing billing generate <customerId> <periodStart> <periodEnd>');
    console.error('  periodStart/periodEnd: ISO格式日期，如 2024-01-01T00:00:00Z');
    process.exit(1);
  }

  try {
    const result = await apiClient.post<GenerateBillResponse>('/bills', {
      customerId,
      periodStart,
      periodEnd,
    });
    printSuccess('账单生成成功');
    printSection('账单信息');
    printKeyValue('账单ID', result.id);
    printKeyValue('客户ID', result.customerId);
    printKeyValue('计费周期', `${formatDate(result.periodStart)} 至 ${formatDate(result.periodEnd)}`);
    console.log('');
    printKeyValue('基础月费', formatMoneyDisplay(result.baseFee));
    printKeyValue('超额费用', formatMoneyDisplay(result.overageFee));
    printKeyValue('折扣', formatMoneyDisplay(result.discount));
    printKeyValue('应付金额', formatMoneyDisplay(result.totalAmount));
    console.log('');
    printKeyValue('状态', formatStatus(result.status));
    printKeyValue('到期时间', formatDate(result.dueAt));
    printKeyValue('复核截止', formatDate(result.reviewDeadline));

    if (result.items && result.items.length > 0) {
      printSection('账单明细');
      const headers = ['类型', '描述', '金额'];
      const rows = result.items.map((item) => [
        item.type,
        item.description,
        formatMoneyDisplay(item.amount),
      ]);
      printTable(headers, rows);
    }
  } catch (error) {
    throw error;
  }
}

async function handleListBills(args: string[]): Promise<void> {
  const customerId = args[0];
  const statusFilter = args[1] as BillStatus | undefined;

  if (!customerId) {
    console.error('用法: billing billing list <customerId> [status]');
    console.error('  status: pending | paid | overdue | reviewing');
    process.exit(1);
  }

  try {
    let path = `/bills?customerId=${encodeURIComponent(customerId)}`;
    if (statusFilter) {
      path += `&status=${encodeURIComponent(statusFilter)}`;
    }

    const result = await apiClient.get<GetBillsResponse>(path);
    printSection(`账单列表 (共 ${result.total} 条)`);

    if (result.bills.length === 0) {
      console.log('暂无账单记录');
      return;
    }

    const headers = ['账单ID', '周期', '金额', '状态', '到期时间'];
    const rows = result.bills.map((bill) => {
      const periodStart = new Date(bill.periodStart).toLocaleDateString('zh-CN');
      const periodEnd = new Date(bill.periodEnd).toLocaleDateString('zh-CN');
      return [
        bill.id,
        `${periodStart} - ${periodEnd}`,
        formatMoneyDisplay(bill.totalAmount),
        formatStatus(bill.status as BillStatus),
        formatDate(bill.dueAt),
      ];
    });
    printTable(headers, rows);
  } catch (error) {
    throw error;
  }
}

async function handleShowBill(args: string[]): Promise<void> {
  const billId = args[0];

  if (!billId) {
    console.error('用法: billing billing show <billId>');
    process.exit(1);
  }

  try {
    const result = await apiClient.get<GetBillDetailResponse>(`/bills/${billId}`);
    printSection('账单详情');
    printKeyValue('账单ID', result.id);
    printKeyValue('客户ID', result.customerId);
    printKeyValue('订阅ID', result.subscriptionId);
    printKeyValue('计费周期', `${formatDate(result.periodStart)} 至 ${formatDate(result.periodEnd)}`);
    console.log('');
    printKeyValue('基础月费', formatMoneyDisplay(result.baseFee));
    printKeyValue('超额费用', formatMoneyDisplay(result.overageFee));
    printKeyValue('折扣', formatMoneyDisplay(result.discount));
    printKeyValue('应付金额', formatMoneyDisplay(result.totalAmount));
    console.log('');
    printKeyValue('状态', formatStatus(result.status));
    printKeyValue('到期时间', formatDate(result.dueAt));
    printKeyValue('支付时间', formatDate(result.paidAt));
    if (result.notes) {
      printKeyValue('备注', result.notes);
    }
    console.log('');
    printKeyValue('创建时间', formatDate(result.createdAt));
    printKeyValue('更新时间', formatDate(result.updatedAt));

    if (result.items && result.items.length > 0) {
      printSection('账单明细');
      const headers = ['类型', '描述', '数量', '单价', '金额'];
      const rows = result.items.map((item) => [
        item.type,
        item.description,
        item.units.toString(),
        formatMoneyDisplay(item.unitPrice),
        formatMoneyDisplay(item.amount),
      ]);
      printTable(headers, rows);
    }
  } catch (error) {
    throw error;
  }
}

async function handlePayBill(args: string[]): Promise<void> {
  const billId = args[0];
  const amountStr = args[1];

  if (!billId || !amountStr) {
    console.error('用法: billing billing pay <billId> <amount>');
    console.error('  amount: 支付金额 (单位: 分，如 9900 表示 99.00元)');
    process.exit(1);
  }

  const amount = parseInt(amountStr, 10);
  if (isNaN(amount) || amount <= 0) {
    console.error('金额必须是正整数');
    process.exit(1);
  }

  try {
    const result = await apiClient.post<PayBillResponse>('/bills/pay', {
      billId,
      amount,
    });
    printSuccess('支付成功');
    printSection('支付信息');
    printKeyValue('支付ID', result.paymentId);
    printKeyValue('账单ID', result.billId);
    printKeyValue('支付金额', formatMoneyDisplay(result.amount));
    printKeyValue('状态', formatStatus(result.status));
    printKeyValue('支付时间', formatDate(result.paidAt));
  } catch (error) {
    throw error;
  }
}

async function handleRequestReview(args: string[]): Promise<void> {
  const billId = args[0];
  const reason = args.slice(1).join(' ') || '账单有疑问';

  if (!billId) {
    console.error('用法: billing billing review <billId> [reason]');
    console.error('  账单生成后15天内可申请复核');
    process.exit(1);
  }

  try {
    const result = await apiClient.post<RequestReviewResponse>('/bills/review', {
      billId,
      reason,
    });
    printSuccess('复核申请已提交');
    printSection('复核信息');
    printKeyValue('账单ID', result.billId);
    printKeyValue('状态', formatStatus(result.status));
    printKeyValue('申请时间', formatDate(result.reviewRequestedAt));
    printKeyValue('复核截止', formatDate(result.reviewDeadline));
  } catch (error) {
    throw error;
  }
}

function printBillingHelp(): void {
  console.log(`
用法: billing billing <command> [options]

账单管理命令:
  generate <customerId> <periodStart> <periodEnd>
                          生成账单
  list <customerId> [status]
                          列出账单 (status: pending | paid | overdue | reviewing)
  show <billId>           显示账单详情
  pay <billId> <amount>   支付账单 (金额单位: 分)
  review <billId> [reason]
                          申请账单复核 (15天内)

示例:
  billing billing generate cus-abc123 "2024-01-01T00:00:00Z" "2024-01-31T23:59:59Z"
  billing billing list cus-abc123
  billing billing list cus-abc123 pending
  billing billing show bil-xyz
  billing billing pay bil-xyz 9900
  billing billing review bil-xyz "超额费用计算有误"
`);
}
