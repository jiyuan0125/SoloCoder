import { apiClient } from '../api-client';
import {
  formatStatus,
  formatDate,
  printSection,
  printKeyValue,
  printSuccess,
  printTable,
  formatPlanType,
  formatBillingCycle,
} from '../formatter';
import type {
  CreateCustomerResponse,
  GetCustomerResponse,
  CheckCustomerStatusResponse,
  PlanType,
  BillingCycle,
} from '@billing/shared';

export async function handleCustomerCommand(args: string[]): Promise<void> {
  if (args.length === 0) {
    printCustomerHelp();
    return;
  }

  const command = args[0];

  switch (command) {
    case 'create':
      await handleCreateCustomer(args.slice(1));
      break;
    case 'get':
      await handleGetCustomer(args.slice(1));
      break;
    case 'status':
      await handleCustomerStatus(args.slice(1));
      break;
    default:
      printCustomerHelp();
  }
}

async function handleCreateCustomer(args: string[]): Promise<void> {
  const name = args[0];
  const email = args[1];

  if (!name || !email) {
    console.error('用法: billing customer create <name> <email>');
    process.exit(1);
  }

  try {
    const result = await apiClient.post<CreateCustomerResponse>('/customers', { name, email });
    printSuccess('客户创建成功');
    printSection('客户信息');
    printKeyValue('ID', result.id);
    printKeyValue('姓名', result.name);
    printKeyValue('邮箱', result.email);
    printKeyValue('状态', formatStatus(result.status));
    printKeyValue('试用期结束', formatDate(result.trialEndAt));
    printKeyValue('创建时间', formatDate(result.createdAt));
  } catch (error) {
    throw error;
  }
}

async function handleGetCustomer(args: string[]): Promise<void> {
  const customerId = args[0];

  if (!customerId) {
    console.error('用法: billing customer get <customerId>');
    process.exit(1);
  }

  try {
    const result = await apiClient.get<GetCustomerResponse>(`/customers/${customerId}`);
    printSection('客户信息');
    printKeyValue('ID', result.id);
    printKeyValue('姓名', result.name);
    printKeyValue('邮箱', result.email);
    printKeyValue('状态', formatStatus(result.status));
    printKeyValue('试用期结束', formatDate(result.trialEndAt));
    printKeyValue('创建时间', formatDate(result.createdAt));

    if (result.subscription) {
      printSection('订阅信息');
      printKeyValue('订阅ID', result.subscription.id);
      printKeyValue('套餐', formatPlanType(result.subscription.planType as PlanType));
      printKeyValue('计费周期', formatBillingCycle(result.subscription.billingCycle as BillingCycle));
      printKeyValue('开始时间', formatDate(result.subscription.startsAt));
      printKeyValue('状态', result.subscription.isActive ? '活跃' : '不活跃');
    }
  } catch (error) {
    throw error;
  }
}

async function handleCustomerStatus(args: string[]): Promise<void> {
  const customerId = args[0];

  if (!customerId) {
    console.error('用法: billing customer status <customerId>');
    process.exit(1);
  }

  try {
    const result = await apiClient.get<CheckCustomerStatusResponse>(`/customers/${customerId}/status`);
    printSection('客户状态');
    printKeyValue('客户ID', result.customerId);
    printKeyValue('状态', formatStatus(result.status));
    printKeyValue('是否冻结', result.isFrozen ? '是' : '否');
    if (result.freezeReason) {
      printKeyValue('冻结原因', result.freezeReason);
    }

    if (result.overdueBills && result.overdueBills.length > 0) {
      printSection('欠费账单');
      const headers = ['账单ID', '金额', '到期时间'];
      const rows = result.overdueBills.map((bill) => [
        bill.billId,
        `¥${(bill.amount as number) / 100}`,
        formatDate(bill.dueAt),
      ]);
      printTable(headers, rows);
    }
  } catch (error) {
    throw error;
  }
}

function printCustomerHelp(): void {
  console.log(`
用法: billing customer <command> [options]

客户管理命令:
  create <name> <email>  创建新客户
  get <customerId>       获取客户信息
  status <customerId>    检查客户状态

示例:
  billing customer create "张三" zhangsan@example.com
  billing customer get cus-abc123
  billing customer status cus-abc123
`);
}
