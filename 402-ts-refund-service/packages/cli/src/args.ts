export type Command =
  | { type: 'help' }
  | { type: 'create-refund'; orderId: string; userId: string; reason: string; productIds: string[] }
  | { type: 'get-refund'; refundId: string }
  | { type: 'list-refunds'; page?: number; pageSize?: number; status?: string; startTime?: string; endTime?: string; orderId?: string }
  | { type: 'submit-logistics'; refundId: string; logisticsNumber: string }
  | { type: 'warehouse-confirm'; refundId: string; received: boolean }
  | { type: 'get-config' }
  | { type: 'update-config'; refundPeriodDays?: number; virtualThreshold?: number }
  | { type: 'add-order'; orderJson: string }
  | { type: 'list-orders' };

function extractValue(args: string[], flag: string): string | null {
  const index = args.findIndex(arg => arg === flag);
  if (index !== -1 && index + 1 < args.length) {
    return args[index + 1];
  }
  return null;
}

function extractBoolean(args: string[], flag: string): boolean | null {
  const value = extractValue(args, flag);
  if (value === null) return null;
  return value.toLowerCase() === 'true' || value === '1';
}

function extractNumber(args: string[], flag: string): number | null {
  const value = extractValue(args, flag);
  if (value === null) return null;
  const num = parseInt(value, 10);
  return isNaN(num) ? null : num;
}

function extractProductIds(args: string[], flag: string): string[] {
  const value = extractValue(args, flag);
  if (value === null) return [];
  return value.split(',').map(s => s.trim()).filter(s => s.length > 0);
}

export function parseArgs(args: string[]): Command {
  if (args.length === 0 || args.includes('--help') || args.includes('-h')) {
    return { type: 'help' };
  }

  const command = args[0];

  switch (command) {
    case 'create':
    case 'create-refund': {
      const orderId = extractValue(args, '--order-id') || extractValue(args, '-o');
      const userId = extractValue(args, '--user-id') || extractValue(args, '-u');
      const reason = extractValue(args, '--reason') || extractValue(args, '-r');
      const productIds = extractProductIds(args, '--products') || extractProductIds(args, '-p');

      if (!orderId || !userId || !reason || productIds.length === 0) {
        return { type: 'help' };
      }

      return {
        type: 'create-refund',
        orderId,
        userId,
        reason,
        productIds
      };
    }

    case 'get':
    case 'get-refund': {
      const refundId = extractValue(args, '--id') || extractValue(args, '-i') || args[1];
      if (!refundId) {
        return { type: 'help' };
      }
      return { type: 'get-refund', refundId };
    }

    case 'list':
    case 'list-refunds': {
      return {
        type: 'list-refunds',
        page: extractNumber(args, '--page') ?? undefined,
        pageSize: extractNumber(args, '--page-size') ?? undefined,
        status: extractValue(args, '--status') ?? undefined,
        startTime: extractValue(args, '--start-time') ?? undefined,
        endTime: extractValue(args, '--end-time') ?? undefined,
        orderId: extractValue(args, '--order-id') ?? undefined
      };
    }

    case 'logistics':
    case 'submit-logistics': {
      const refundId = extractValue(args, '--refund-id') || extractValue(args, '-r') || args[1];
      const logisticsNumber = extractValue(args, '--tracking') || extractValue(args, '-t') || args[2];

      if (!refundId || !logisticsNumber) {
        return { type: 'help' };
      }

      return {
        type: 'submit-logistics',
        refundId,
        logisticsNumber
      };
    }

    case 'confirm':
    case 'warehouse-confirm': {
      const refundId = extractValue(args, '--refund-id') || extractValue(args, '-r') || args[1];
      const received = extractBoolean(args, '--received') ?? true;

      if (!refundId) {
        return { type: 'help' };
      }

      return {
        type: 'warehouse-confirm',
        refundId,
        received
      };
    }

    case 'config':
    case 'get-config': {
      return { type: 'get-config' };
    }

    case 'update-config': {
      const refundPeriodDays = extractNumber(args, '--period') ?? undefined;
      const virtualThreshold = extractNumber(args, '--virtual-threshold') ?? undefined;

      if (refundPeriodDays === undefined && virtualThreshold === undefined) {
        return { type: 'help' };
      }

      return {
        type: 'update-config',
        refundPeriodDays,
        virtualThreshold
      };
    }

    case 'add-order': {
      const orderJson = args[1];
      if (!orderJson) {
        return { type: 'help' };
      }
      return { type: 'add-order', orderJson };
    }

    case 'list-orders': {
      return { type: 'list-orders' };
    }

    default:
      return { type: 'help' };
  }
}

export function getHelpText(): string {
  return `
退款服务 CLI 工具

用法: refund-cli <command> [options]

命令:
  create-refund    创建退款申请
    选项:
      -o, --order-id <id>       订单ID
      -u, --user-id <id>        用户ID
      -r, --reason <text>       退款原因
      -p, --products <ids>      商品ID列表, 逗号分隔

  get-refund       获取退款详情
    选项:
      -i, --id <id>             退款ID

  list-refunds     列出退款列表
    选项:
      --page <num>              页码 (默认: 1)
      --page-size <num>         每页数量 (默认: 10)
      --status <status>         按状态筛选
      --start-time <time>       开始时间
      --end-time <time>         结束时间
      --order-id <id>           按订单ID筛选

  submit-logistics 提交物流单号
    选项:
      -r, --refund-id <id>      退款ID
      -t, --tracking <num>      物流单号

  warehouse-confirm 仓库确认收货
    选项:
      -r, --refund-id <id>      退款ID
      --received <bool>         是否收到 (true/false, 默认: true)

  get-config       获取系统配置

  update-config    更新系统配置
    选项:
      --period <days>           售后天数
      --virtual-threshold <pct> 虚拟商品消费阈值 (百分比)

  add-order        添加测试订单 (JSON格式)

  list-orders      列出所有订单

  help, -h, --help 显示此帮助信息

示例:
  refund-cli create-refund -o ORDER-001 -u USER-001 -r "商品有缺陷" -p "PROD-001,PROD-002"
  refund-cli get-refund -i REFUND-123
  refund-cli list-refunds --status pending
  refund-cli submit-logistics -r REFUND-123 -t SF1234567890
  refund-cli warehouse-confirm -r REFUND-123 --received true
  refund-cli get-config
  refund-cli update-config --period 45 --virtual-threshold 70
`;
}
