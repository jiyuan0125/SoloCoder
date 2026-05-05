import { Refund, RefundStatus, Order, Config } from '@refund/shared';

const statusLabels: Record<RefundStatus, string> = {
  pending: '待处理',
  processing: '处理中',
  waiting_for_return: '待退货',
  waiting_for_warehouse_confirm: '待仓库确认',
  refunded: '已退款',
  rejected: '已拒绝'
};

function formatAmount(amount: number): string {
  const yuan = Math.floor(amount / 100);
  const fen = amount % 100;
  return `${yuan}.${fen.toString().padStart(2, '0')} 元`;
}

function formatDate(date: Date): string {
  const d = new Date(date);
  const year = d.getFullYear();
  const month = (d.getMonth() + 1).toString().padStart(2, '0');
  const day = d.getDate().toString().padStart(2, '0');
  const hours = d.getHours().toString().padStart(2, '0');
  const minutes = d.getMinutes().toString().padStart(2, '0');
  const seconds = d.getSeconds().toString().padStart(2, '0');
  return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`;
}

export function formatRefund(refund: Refund): string {
  const lines: string[] = [];
  
  lines.push('╔══════════════════════════════════════════════════════════════╗');
  lines.push('║                        退款详情                                ║');
  lines.push('╠══════════════════════════════════════════════════════════════╣');
  lines.push(`║ 退款ID:   ${refund.refundId.padEnd(50)}║`);
  lines.push(`║ 订单ID:   ${refund.orderId.padEnd(50)}║`);
  lines.push(`║ 用户ID:   ${refund.userId.padEnd(50)}║`);
  lines.push(`║ 状态:     ${statusLabels[refund.status].padEnd(50)}║`);
  lines.push(`║ 退款金额: ${formatAmount(refund.refundAmount).padEnd(50)}║`);
  lines.push(`║ 运费退款: ${formatAmount(refund.shippingFeeRefund).padEnd(50)}║`);
  lines.push(`║ 退款原因: ${refund.refundReason.substring(0, 48).padEnd(50)}║`);
  lines.push(`║ 整单退款: ${(refund.isFullRefund ? '是' : '否').padEnd(50)}║`);
  if (refund.logisticsNumber) {
    lines.push(`║ 物流单号: ${refund.logisticsNumber.padEnd(50)}║`);
  }
  lines.push(`║ 创建时间: ${formatDate(refund.createdAt).padEnd(50)}║`);
  lines.push(`║ 更新时间: ${formatDate(refund.updatedAt).padEnd(50)}║`);
  lines.push('╠══════════════════════════════════════════════════════════════╣');
  lines.push('║                        退款商品                                ║');
  lines.push('╠══════════════════════════════════════════════════════════════╣');
  
  refund.products.forEach((product, index) => {
    lines.push(`║ [${index + 1}] ${product.productName.substring(0, 30).padEnd(32)}║`);
    lines.push(`║     商品ID: ${product.productId.padEnd(45)}║`);
    lines.push(`║     类型: ${(product.productType === 'physical' ? '实物' : '虚拟').padEnd(47)}║`);
    lines.push(`║     数量: ${product.quantity.toString().padEnd(47)}║`);
    lines.push(`║     原支付: ${formatAmount(product.paymentAmount).padEnd(45)}║`);
    lines.push(`║     退款: ${formatAmount(product.refundAmount).padEnd(47)}║`);
    if (product.consumptionProgress !== undefined) {
      lines.push(`║     消费进度: ${product.consumptionProgress}%${''.padEnd(41)}║`);
    }
    if (index < refund.products.length - 1) {
      lines.push('║ ─────────────────────────────────────────────────────────────║');
    }
  });
  
  lines.push('╠══════════════════════════════════════════════════════════════╣');
  lines.push('║                        状态历史                                ║');
  lines.push('╠══════════════════════════════════════════════════════════════╣');
  
  refund.statusHistory.forEach((item, index) => {
    const statusLabel = statusLabels[item.status];
    const timeStr = formatDate(item.time);
    const note = item.note ? ` (${item.note})` : '';
    const line = `${statusLabel} - ${timeStr}${note}`;
    lines.push(`║ ${line.substring(0, 60).padEnd(60)}║`);
  });
  
  lines.push('╚══════════════════════════════════════════════════════════════╝');
  
  return lines.join('\n');
}

export function formatRefundList(refunds: Refund[], total: number, page: number, pageSize: number): string {
  const lines: string[] = [];
  
  lines.push('╔══════════════════════════════════════════════════════════════╗');
  lines.push(`║                    退款列表 (共 ${total} 条, 第 ${page} 页)                     ║`);
  lines.push('╠══════════════════════════════════════════════════════════════╣');
  
  if (refunds.length === 0) {
    lines.push('║                        暂无数据                                ║');
  } else {
    refunds.forEach((refund, index) => {
      lines.push(`║ [${index + 1}] 退款ID: ${refund.refundId.padEnd(43)}║`);
      lines.push(`║     订单: ${refund.orderId.padEnd(47)}║`);
      lines.push(`║     金额: ${formatAmount(refund.refundAmount).padEnd(45)}║`);
      lines.push(`║     状态: ${statusLabels[refund.status].padEnd(47)}║`);
      lines.push(`║     时间: ${formatDate(refund.createdAt).padEnd(45)}║`);
      if (index < refunds.length - 1) {
        lines.push('║ ─────────────────────────────────────────────────────────────║');
      }
    });
  }
  
  lines.push('╚══════════════════════════════════════════════════════════════╝');
  
  return lines.join('\n');
}

export function formatOrder(order: Order): string {
  const lines: string[] = [];
  
  lines.push('╔══════════════════════════════════════════════════════════════╗');
  lines.push('║                        订单详情                                ║');
  lines.push('╠══════════════════════════════════════════════════════════════╣');
  lines.push(`║ 订单ID:   ${order.orderId.padEnd(50)}║`);
  lines.push(`║ 用户ID:   ${order.userId.padEnd(50)}║`);
  lines.push(`║ 下单时间: ${formatDate(order.orderTime).padEnd(50)}║`);
  lines.push(`║ 总金额:   ${formatAmount(order.totalAmount).padEnd(50)}║`);
  lines.push(`║ 运费:     ${formatAmount(order.shippingFee).padEnd(50)}║`);
  lines.push('╠══════════════════════════════════════════════════════════════╣');
  lines.push('║                        商品列表                                ║');
  lines.push('╠══════════════════════════════════════════════════════════════╣');
  
  order.products.forEach((product, index) => {
    lines.push(`║ [${index + 1}] ${product.productName.substring(0, 30).padEnd(32)}║`);
    lines.push(`║     商品ID: ${product.productId.padEnd(45)}║`);
    lines.push(`║     类型: ${(product.productType === 'physical' ? '实物' : '虚拟').padEnd(47)}║`);
    lines.push(`║     单价: ${formatAmount(product.unitPrice).padEnd(47)}║`);
    lines.push(`║     数量: ${product.quantity.toString().padEnd(47)}║`);
    lines.push(`║     支付: ${formatAmount(product.paymentAmount).padEnd(47)}║`);
    if (product.consumptionProgress !== undefined) {
      lines.push(`║     消费进度: ${product.consumptionProgress}%${''.padEnd(41)}║`);
    }
    if (index < order.products.length - 1) {
      lines.push('║ ─────────────────────────────────────────────────────────────║');
    }
  });
  
  lines.push('╚══════════════════════════════════════════════════════════════╝');
  
  return lines.join('\n');
}

export function formatConfig(config: Config): string {
  const lines: string[] = [];
  
  lines.push('╔══════════════════════════════════════════════════════════════╗');
  lines.push('║                        系统配置                                ║');
  lines.push('╠══════════════════════════════════════════════════════════════╣');
  lines.push(`║ 售后天数:          ${config.refundPeriodDays.toString().padEnd(41)}天║`);
  lines.push(`║ 虚拟商品消费阈值:  ${config.virtualProductRefundThreshold.toString().padEnd(41)}%║`);
  lines.push(`║ 数据存储路径:      ${config.dataStoragePath.padEnd(41)}║`);
  lines.push('╚══════════════════════════════════════════════════════════════╝');
  
  return lines.join('\n');
}

export function formatSuccess(message: string): string {
  return `✓ ${message}`;
}

export function formatError(code: string, message: string): string {
  return `✗ [${code}] ${message}`;
}
