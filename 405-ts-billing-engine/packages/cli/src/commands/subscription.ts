import { apiClient } from '../api-client';
import {
  formatStatus,
  formatDate,
  formatPlanType,
  formatBillingCycle,
  formatMoneyDisplay,
  printSection,
  printKeyValue,
  printSuccess,
} from '../formatter';
import type {
  CreateSubscriptionResponse,
  UpgradeSubscriptionResponse,
  RequestRefundResponse,
  PlanType,
  BillingCycle,
} from '@billing/shared';

export async function handleSubscriptionCommand(args: string[]): Promise<void> {
  if (args.length === 0) {
    printSubscriptionHelp();
    return;
  }

  const command = args[0];

  switch (command) {
    case 'create':
      await handleCreateSubscription(args.slice(1));
      break;
    case 'upgrade':
      await handleUpgradeSubscription(args.slice(1));
      break;
    case 'refund':
      await handleRequestRefund(args.slice(1));
      break;
    default:
      printSubscriptionHelp();
  }
}

async function handleCreateSubscription(args: string[]): Promise<void> {
  const customerId = args[0];
  const planTypeArg = args[1]?.toLowerCase() as PlanType;
  const billingCycleArg = (args[2]?.toLowerCase() || 'monthly') as BillingCycle;

  if (!customerId || !planTypeArg) {
    console.error('用法: billing subscription create <customerId> <planType> [billingCycle]');
    console.error('  planType: free | basic | pro');
    console.error('  billingCycle: monthly | yearly (默认: monthly)');
    process.exit(1);
  }

  const validPlans: PlanType[] = ['free', 'basic', 'pro'];
  if (!validPlans.includes(planTypeArg)) {
    console.error('无效的套餐类型。可选: free, basic, pro');
    process.exit(1);
  }

  const validCycles: BillingCycle[] = ['monthly', 'yearly'];
  if (!validCycles.includes(billingCycleArg)) {
    console.error('无效的计费周期。可选: monthly, yearly');
    process.exit(1);
  }

  try {
    const result = await apiClient.post<CreateSubscriptionResponse>('/subscriptions', {
      customerId,
      planType: planTypeArg,
      billingCycle: billingCycleArg,
    });
    printSuccess('订阅创建成功');
    printSection('订阅信息');
    printKeyValue('订阅ID', result.id);
    printKeyValue('客户ID', result.customerId);
    printKeyValue('套餐', formatPlanType(result.planType as PlanType));
    printKeyValue('计费周期', formatBillingCycle(result.billingCycle as BillingCycle));
    printKeyValue('开始时间', formatDate(result.startsAt));
    printKeyValue('状态', result.isActive ? '活跃' : '不活跃');
    if (result.yearlyDiscountApplied) {
      printKeyValue('年度优惠', '已应用 (9折)');
    }
    printKeyValue('创建时间', formatDate(result.createdAt));
  } catch (error) {
    throw error;
  }
}

async function handleUpgradeSubscription(args: string[]): Promise<void> {
  const customerId = args[0];
  const subscriptionId = args[1];
  const toPlanType = args[2]?.toLowerCase() as PlanType;

  if (!customerId || !subscriptionId || !toPlanType) {
    console.error('用法: billing subscription upgrade <customerId> <subscriptionId> <toPlanType>');
    console.error('  toPlanType: free | basic | pro');
    process.exit(1);
  }

  try {
    const result = await apiClient.post<UpgradeSubscriptionResponse>('/subscriptions/upgrade', {
      customerId,
      subscriptionId,
      toPlanType,
    });
    printSuccess('订阅升级成功');
    printSection('升级信息');
    printKeyValue('原套餐', formatPlanType(result.fromPlan as PlanType));
    printKeyValue('新套餐', formatPlanType(result.toPlan as PlanType));
    printKeyValue('升级时间', formatDate(result.upgradeDate));

    printSection('按天折算');
    printKeyValue('原套餐使用天数', `${result.proration.daysOnOldPlan} 天`);
    printKeyValue('新套餐剩余天数', `${result.proration.daysOnNewPlan} 天`);
    printKeyValue('原套餐折算费用', formatMoneyDisplay(result.proration.oldPlanProratedFee));
    printKeyValue('新套餐折算费用', formatMoneyDisplay(result.proration.newPlanProratedFee));
    printKeyValue('本月基础费用合计', formatMoneyDisplay(result.proration.totalBaseFee));
  } catch (error) {
    throw error;
  }
}

async function handleRequestRefund(args: string[]): Promise<void> {
  const customerId = args[0];
  const subscriptionId = args[1];
  const reason = args.slice(2).join(' ') || '客户申请退款';

  if (!customerId || !subscriptionId) {
    console.error('用法: billing subscription refund <customerId> <subscriptionId> [reason]');
    process.exit(1);
  }

  try {
    const result = await apiClient.post<RequestRefundResponse>('/subscriptions/refund', {
      customerId,
      subscriptionId,
      reason,
    });
    printSuccess('退款申请已提交');
    printSection('退款信息');
    printKeyValue('退款ID', result.refundId);
    printKeyValue('退款月份数', `${result.refundedMonths} 个月`);
    printKeyValue('退款金额', formatMoneyDisplay(result.amount));
    printKeyValue('状态', formatStatus(result.status));
  } catch (error) {
    throw error;
  }
}

function printSubscriptionHelp(): void {
  console.log(`
用法: billing subscription <command> [options]

订阅管理命令:
  create <customerId> <planType> [billingCycle]
                          创建新订阅
                          planType: free | basic | pro
                          billingCycle: monthly | yearly (默认: monthly)
  upgrade <customerId> <subscriptionId> <toPlanType>
                          升级订阅套餐
  refund <customerId> <subscriptionId> [reason]
                          申请年度订阅退款 (第10个月后可申请)

示例:
  billing subscription create cus-abc123 pro yearly
  billing subscription upgrade cus-abc123 sub-xyz basic pro
  billing subscription refund cus-abc123 sub-xyz "不再需要服务"
`);
}
