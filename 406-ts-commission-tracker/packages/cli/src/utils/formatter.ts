import {
  Salesperson,
  Order,
  Settlement,
  CommissionCalculation,
  Ranking,
  AmountInCents,
  SalesContribution,
} from '@commission-tracker/shared';

export function formatCents(amount: AmountInCents): string {
  const yuan = Math.floor(Math.abs(amount) / 100);
  const cents = Math.abs(amount) % 100;
  const sign = amount < 0 ? '-' : '';
  return `${sign}¥${yuan.toLocaleString()}.${String(cents).padStart(2, '0')}`;
}

export function formatPercentage(value: number): string {
  return `${(value * 100).toFixed(0)}%`;
}

export function formatDate(dateString: string): string {
  const date = new Date(dateString);
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  });
}

export function formatSalesperson(salesperson: Salesperson): string {
  const lines: string[] = [];
  lines.push(`销售ID: ${salesperson.id}`);
  lines.push(`姓名: ${salesperson.name}`);
  lines.push(`入职日期: ${salesperson.joinDate}`);
  lines.push(`状态: ${salesperson.status === 'active' ? '在职' : '已离职'}`);
  lines.push(`佣金余额: ${formatCents(salesperson.balance)}`);
  return lines.join('\n');
}

export function formatSalespersonList(salespeople: Salesperson[]): string {
  if (salespeople.length === 0) {
    return '暂无销售数据';
  }

  const lines: string[] = [];
  lines.push('销售列表:');
  lines.push('='.repeat(80));
  
  for (const sp of salespeople) {
    lines.push(`[${sp.id}] ${sp.name}`);
    lines.push(`  状态: ${sp.status === 'active' ? '在职' : '已离职'}`);
    lines.push(`  入职日期: ${sp.joinDate}`);
    lines.push(`  佣金余额: ${formatCents(sp.balance)}`);
    lines.push('');
  }

  return lines.join('\n');
}

export function formatOrder(order: Order): string {
  const lines: string[] = [];
  lines.push(`订单ID: ${order.id}`);
  lines.push(`金额: ${formatCents(order.amount)}`);
  lines.push(`日期: ${order.date}`);
  lines.push(`状态: ${order.status === 'active' ? '有效' : '已退款'}`);
  lines.push(`是否锁定: ${order.locked ? '是' : '否'}`);
  lines.push(`销售贡献:`);
  
  for (const contrib of order.salesContributions) {
    lines.push(`  - ${contrib.salespersonId}${contrib.isPrimary ? ' (主销售)' : ''}`);
  }

  return lines.join('\n');
}

export function formatOrderList(orders: Order[]): string {
  if (orders.length === 0) {
    return '暂无订单数据';
  }

  const lines: string[] = [];
  lines.push('订单列表:');
  lines.push('='.repeat(80));
  
  for (const order of orders) {
    const salesNames = order.salesContributions
      .map((c: SalesContribution) => `${c.salespersonId}${c.isPrimary ? '(主)' : ''}`)
      .join(', ');
    
    lines.push(`[${order.id}] 金额: ${formatCents(order.amount)}`);
    lines.push(`  日期: ${order.date} | 状态: ${order.status === 'active' ? '有效' : '已退款'} | 锁定: ${order.locked ? '是' : '否'}`);
    lines.push(`  销售: ${salesNames}`);
    lines.push('');
  }

  return lines.join('\n');
}

export function formatCommissionCalculation(calc: CommissionCalculation): string {
  const lines: string[] = [];
  lines.push(`销售ID: ${calc.salespersonId}`);
  lines.push(`月份: ${calc.year}年${calc.month}月`);
  lines.push(`总销售额: ${formatCents(calc.totalSales)}`);
  lines.push(`折扣前佣金: ${formatCents(calc.commissionBeforeDiscount)}`);
  lines.push(`折扣率: ${formatPercentage(calc.discountRate)}`);
  lines.push(`最终佣金: ${formatCents(calc.finalCommission)}`);
  lines.push('');
  lines.push('阶梯明细:');
  lines.push('-'.repeat(60));
  
  for (const tier of calc.tierBreakdown) {
    lines.push(`  ${tier.tier}: 销售额 ${formatCents(tier.salesAmount)} → 佣金 ${formatCents(tier.commission)}`);
  }

  return lines.join('\n');
}

export function formatSettlement(settlement: Settlement): string {
  const lines: string[] = [];
  lines.push(`结算单ID: ${settlement.id}`);
  lines.push(`销售ID: ${settlement.salespersonId}`);
  lines.push(`结算月份: ${settlement.year}年${settlement.month}月`);
  lines.push(`结算状态: ${settlement.status === 'completed' ? '已完成' : settlement.status}`);
  lines.push(`结算时间: ${formatDate(settlement.settledAt)}`);
  lines.push('');
  lines.push('汇总信息:');
  lines.push('-'.repeat(60));
  lines.push(`  总销售额: ${formatCents(settlement.summary.totalSales)}`);
  lines.push(`  折扣前佣金: ${formatCents(settlement.summary.totalCommissionBeforeDiscount)}`);
  lines.push(`  折扣金额: ${formatCents(settlement.summary.totalDiscount)}`);
  lines.push(`  退款扣除: ${formatCents(settlement.summary.totalRefundDeduction)}`);
  lines.push(`  余额调整: ${formatCents(settlement.summary.balanceAdjustment)}`);
  lines.push(`  最终结算金额: ${formatCents(settlement.summary.finalSettlementAmount)}`);
  
  if (settlement.details.orders.length > 0) {
    lines.push('');
    lines.push('订单明细:');
    lines.push('-'.repeat(60));
    for (const item of settlement.details.orders) {
      lines.push(`  订单 ${item.orderId}: 金额 ${formatCents(item.amount)} → 佣金 ${formatCents(item.commission)}`);
    }
  }
  
  if (settlement.details.refunds.length > 0) {
    lines.push('');
    lines.push('退款明细:');
    lines.push('-'.repeat(60));
    for (const item of settlement.details.refunds) {
      lines.push(`  订单 ${item.orderId}: 金额 ${formatCents(item.amount)} → 扣回佣金 ${formatCents(item.commissionDeducted)}`);
    }
  }

  return lines.join('\n');
}

export function formatSettlementList(settlements: Settlement[]): string {
  if (settlements.length === 0) {
    return '暂无结算记录';
  }

  const lines: string[] = [];
  lines.push('结算记录列表:');
  lines.push('='.repeat(80));
  
  for (const s of settlements) {
    lines.push(`[${s.id}] ${s.year}年${s.month}月`);
    lines.push(`  销售ID: ${s.salespersonId}`);
    lines.push(`  结算金额: ${formatCents(s.summary.finalSettlementAmount)}`);
    lines.push(`  状态: ${s.status === 'completed' ? '已完成' : s.status}`);
    lines.push(`  时间: ${formatDate(s.settledAt)}`);
    lines.push('');
  }

  return lines.join('\n');
}

export function formatRankingList(rankings: Ranking[]): string {
  if (rankings.length === 0) {
    return '暂无排名数据';
  }

  const lines: string[] = [];
  lines.push('销售业绩排名:');
  lines.push('='.repeat(80));
  lines.push(`  排名  销售ID      姓名        销售额          佣金`);
  lines.push('-'.repeat(80));
  
  for (const r of rankings) {
    const rankStr = String(r.rank).padStart(4, ' ');
    const idStr = r.salespersonId.padEnd(10, ' ');
    const nameStr = r.salespersonName.padEnd(10, ' ');
    const salesStr = formatCents(r.totalSales).padStart(14, ' ');
    const commissionStr = formatCents(r.totalCommission).padStart(14, ' ');
    
    lines.push(`  ${rankStr}  ${idStr}${nameStr}${salesStr}${commissionStr}`);
  }

  return lines.join('\n');
}
