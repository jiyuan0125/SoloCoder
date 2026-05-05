import {
  CreateOrderResponse,
  PayOrderResponse,
  OrderQueryResponse,
  CreateRefundResponse,
  RefundQueryResponse,
  CloseOrderResponse,
  CallbackResponse,
  ApiResponse,
  OrderStatus,
  PaymentChannel,
} from "@payment-gateway/shared";

const STATUS_COLORS: Record<string, string> = {
  [OrderStatus.PENDING]: "\x1b[33m",
  [OrderStatus.PROCESSING]: "\x1b[34m",
  [OrderStatus.SUCCESS]: "\x1b[32m",
  [OrderStatus.FAILED]: "\x1b[31m",
  [OrderStatus.PENDING_CONFIRM]: "\x1b[35m",
  [OrderStatus.CLOSED]: "\x1b[90m",
  [OrderStatus.REFUNDED]: "\x1b[36m",
  [OrderStatus.PARTIAL_REFUNDED]: "\x1b[36m",
};

const CHANNEL_NAMES: Record<PaymentChannel, string> = {
  [PaymentChannel.WECHAT]: "微信支付",
  [PaymentChannel.ALIPAY]: "支付宝",
};

const RESET_COLOR = "\x1b[0m";

function colorizeStatus(status: string): string {
  const color = STATUS_COLORS[status] ?? "";
  return `${color}${status}${RESET_COLOR}`;
}

function formatTimestamp(timestamp: number | null): string {
  if (timestamp === null) {
    return "N/A";
  }
  return new Date(timestamp).toISOString();
}

function formatAmount(amount: number): string {
  const yuan = Math.floor(amount / 100);
  const fen = amount % 100;
  return `${yuan}.${fen.toString().padStart(2, "0")} 元 (${amount} 分)`;
}

export function formatApiResponse<T>(response: ApiResponse<T>): string {
  if (response.code === 0) {
    return `\x1b[32m✓ 成功\x1b[0m`;
  }
  return `\x1b[31m✗ 错误 (${response.code}): ${response.message}\x1b[0m`;
}

export function formatCreateOrderResponse(response: CreateOrderResponse): string {
  const lines = [
    "",
    "┌─────────────────────────────────────────┐",
    "│           订单创建成功                   │",
    "├─────────────────────────────────────────┤",
    `│ 订单号: ${response.orderNo}`,
    `│ 金额: ${formatAmount(response.amount)}`,
    `│ 渠道: ${CHANNEL_NAMES[response.channel]}`,
    `│ 状态: ${colorizeStatus(response.status)}`,
    `│ 支付链接: ${response.paymentUrl}`,
    `│ 创建时间: ${formatTimestamp(response.createdAt)}`,
    "└─────────────────────────────────────────┘",
    "",
  ];
  return lines.join("\n");
}

export function formatPayOrderResponse(response: PayOrderResponse): string {
  const lines = [
    "",
    "┌─────────────────────────────────────────┐",
    "│           支付发起成功                   │",
    "├─────────────────────────────────────────┤",
    `│ 订单号: ${response.orderNo}`,
    `│ 金额: ${formatAmount(response.amount)}`,
    `│ 渠道: ${CHANNEL_NAMES[response.channel]}`,
    `│ 状态: ${colorizeStatus(response.status)}`,
    `│ 支付ID: ${response.paymentId}`,
    `│ 支付链接: ${response.paymentUrl}`,
    "└─────────────────────────────────────────┘",
    "",
  ];
  return lines.join("\n");
}

export function formatOrderQueryResponse(response: OrderQueryResponse): string {
  const lines = [
    "",
    "┌─────────────────────────────────────────┐",
    "│           订单详情                       │",
    "├─────────────────────────────────────────┤",
    `│ 订单号: ${response.orderNo}`,
    `│ 金额: ${formatAmount(response.amount)}`,
    `│ 渠道: ${CHANNEL_NAMES[response.channel]}`,
    `│ 状态: ${colorizeStatus(response.status)}`,
    `│ 交易流水号: ${response.transactionNo ?? "N/A"}`,
    `│ 支付时间: ${formatTimestamp(response.paidAt)}`,
    `│ 创建时间: ${formatTimestamp(response.createdAt)}`,
    `│ 已退款金额: ${formatAmount(response.totalRefundedAmount)}`,
    `│ 可退款金额: ${formatAmount(response.refundableAmount)}`,
  ];

  if (response.refunds.length > 0) {
    lines.push("├─────────────────────────────────────────┤");
    lines.push("│ 退款记录:");
    for (const refund of response.refunds) {
      lines.push(`│   - 退款单号: ${refund.refundNo}`);
      lines.push(`│     金额: ${formatAmount(refund.amount)}`);
      lines.push(`│     状态: ${colorizeStatus(refund.status)}`);
      lines.push(`│     完成时间: ${formatTimestamp(refund.completedAt)}`);
    }
  }

  lines.push("└─────────────────────────────────────────┘");
  lines.push("");

  return lines.join("\n");
}

export function formatCreateRefundResponse(response: CreateRefundResponse): string {
  const lines = [
    "",
    "┌─────────────────────────────────────────┐",
    "│           退款申请成功                   │",
    "├─────────────────────────────────────────┤",
    `│ 订单号: ${response.orderNo}`,
    `│ 退款单号: ${response.refundNo}`,
    `│ 金额: ${formatAmount(response.amount)}`,
    `│ 渠道: ${CHANNEL_NAMES[response.channel]}`,
    `│ 状态: ${colorizeStatus(response.status)}`,
    `│ 申请时间: ${formatTimestamp(response.createdAt)}`,
    "└─────────────────────────────────────────┘",
    "",
  ];
  return lines.join("\n");
}

export function formatRefundQueryResponse(response: RefundQueryResponse): string {
  const lines = [
    "",
    "┌─────────────────────────────────────────┐",
    "│           退款详情                       │",
    "├─────────────────────────────────────────┤",
    `│ 订单号: ${response.orderNo}`,
    `│ 订单金额: ${formatAmount(response.totalAmount)}`,
    `│ 累计退款: ${formatAmount(response.totalRefundedAmount)}`,
    `│ 可退款金额: ${formatAmount(response.refundableAmount)}`,
  ];

  if (response.refunds.length > 0) {
    lines.push("├─────────────────────────────────────────┤");
    lines.push("│ 退款记录:");
    for (const refund of response.refunds) {
      lines.push(`│   ┌─────────────────────────────────┐`);
      lines.push(`│   │ 退款单号: ${refund.refundNo}`);
      lines.push(`│   │ 金额: ${formatAmount(refund.amount)}`);
      lines.push(`│   │ 状态: ${colorizeStatus(refund.status)}`);
      lines.push(`│   │ 渠道退款ID: ${refund.channelRefundId ?? "N/A"}`);
      lines.push(`│   │ 是否最终退款: ${refund.isFinalRefund ? "是" : "否"}`);
      lines.push(`│   │ 申请时间: ${formatTimestamp(refund.createdAt)}`);
      lines.push(`│   │ 完成时间: ${formatTimestamp(refund.completedAt)}`);
      lines.push(`│   └─────────────────────────────────┘`);
    }
  } else {
    lines.push("│ 退款记录: 无");
  }

  lines.push("└─────────────────────────────────────────┘");
  lines.push("");

  return lines.join("\n");
}

export function formatCloseOrderResponse(response: CloseOrderResponse): string {
  const lines = [
    "",
    "┌─────────────────────────────────────────┐",
    "│           订单关闭成功                   │",
    "├─────────────────────────────────────────┤",
    `│ 订单号: ${response.orderNo}`,
    `│ 状态: ${colorizeStatus(response.status)}`,
    `│ 关闭时间: ${formatTimestamp(response.closedAt)}`,
    "└─────────────────────────────────────────┘",
    "",
  ];
  return lines.join("\n");
}

export function formatCallbackResponse(response: CallbackResponse): string {
  if (response.success) {
    return `\n\x1b[32m✓ 回调处理成功: ${response.message}\x1b[0m\n`;
  }
  return `\n\x1b[31m✗ 回调处理失败: ${response.message}\x1b[0m\n`;
}

export function formatHelp(): string {
  const lines = [
    "",
    "支付网关 CLI 工具",
    "",
    "用法: payment-gateway <命令> [选项]",
    "",
    "命令:",
    "  create-order <orderNo> <amount>          创建订单",
    "  pay-order <orderNo>                       发起支付",
    "  query-order <orderNo>                     查询订单",
    "  close-order <orderNo>                     关闭订单",
    "  create-refund <orderNo> <refundNo> <amount> [--final]",
    "                                            创建退款",
    "  query-refund <orderNo> [refundNo]         查询退款",
    "  callback-wechat <orderNo> <amount> [--success|--failed]",
    "                                            模拟微信支付回调",
    "  callback-alipay <orderNo> <amount> [--success|--failed]",
    "                                            模拟支付宝支付回调",
    "  callback-refund-wechat <orderNo> <refundNo> <amount> [--success|--failed]",
    "                                            模拟微信退款回调",
    "  callback-refund-alipay <orderNo> <refundNo> <amount> [--success|--failed]",
    "                                            模拟支付宝退款回调",
    "  help                                      显示帮助信息",
    "",
    "说明:",
    "  - 金额以分为单位(整数)",
    "  - 100分以下自动走微信支付",
    "  - 100分及以上自动走支付宝",
    "  - 订单关闭后不能重新支付，需创建新订单",
    "",
  ];
  return lines.join("\n");
}
