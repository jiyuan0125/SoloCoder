import {
  CATEGORY_LABELS,
  STATUS_LABELS,
  ROLE_LABELS,
} from '@expense-report/shared';

function formatAmount(amountCents: number): string {
  const yuan = Math.floor(amountCents / 100);
  const fen = amountCents % 100;
  return `${yuan}.${fen.toString().padStart(2, '0')} 元`;
}

function formatDate(dateStr: string): string {
  try {
    const date = new Date(dateStr);
    return date.toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    });
  } catch {
    return dateStr;
  }
}

function formatShortDate(dateStr: string): string {
  try {
    const date = new Date(dateStr);
    return date.toLocaleDateString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
    });
  } catch {
    return dateStr;
  }
}

function getCategoryLabel(category: string): string {
  const label = CATEGORY_LABELS[category as keyof typeof CATEGORY_LABELS];
  return label ?? category;
}

function getStatusLabel(status: string): string {
  const label = STATUS_LABELS[status as keyof typeof STATUS_LABELS];
  return label ?? status;
}

function getRoleLabel(role: string): string {
  const label = ROLE_LABELS[role as keyof typeof ROLE_LABELS];
  return label ?? role;
}

function getString(obj: Record<string, unknown>, key: string, defaultValue: string = ''): string {
  const value = obj[key];
  if (typeof value === 'string') return value;
  return defaultValue;
}

function getNumber(obj: Record<string, unknown>, key: string, defaultValue: number = 0): number {
  const value = obj[key];
  if (typeof value === 'number') return value;
  return defaultValue;
}

function getBoolean(obj: Record<string, unknown>, key: string): boolean {
  const value = obj[key];
  return value === true;
}

function formatExpense(expense: Record<string, unknown>): string {
  const lines: string[] = [];
  lines.push('╔════════════════════════════════════════════════════════════╗');
  lines.push('║                        报销单详情                            ║');
  lines.push('╠════════════════════════════════════════════════════════════╣');

  const amount = getNumber(expense, 'amount');
  const status = getString(expense, 'status');
  const category = getString(expense, 'category');

  lines.push(`║ 报销单ID: ${getString(expense, 'id').padEnd(50)}║`);
  lines.push(`║ 员工: ${(getString(expense, 'employeeName') || getString(expense, 'employeeId')).padEnd(54)}║`);
  lines.push(`║ 日期: ${formatShortDate(getString(expense, 'date')).padEnd(54)}║`);
  lines.push(`║ 金额: ${formatAmount(amount).padEnd(54)}║`);
  lines.push(`║ 类别: ${getCategoryLabel(category).padEnd(54)}║`);
  lines.push(`║ 状态: ${getStatusLabel(status).padEnd(54)}║`);

  const needsSupplementary = getBoolean(expense, 'needsSupplementaryInfo');
  if (needsSupplementary) {
    lines.push(`║ ⚠  需要补充说明: 是${' '.repeat(47)}║`);
  }

  const approvalLevel = getString(expense, 'approvalLevel');
  if (approvalLevel) {
    const levelLabel = approvalLevel === 'dept_manager' ? '部门经理已审批' : '财务总监已审批';
    lines.push(`║ 审批进度: ${levelLabel.padEnd(49)}║`);
  }

  lines.push(`║ 事由: ${getString(expense, 'reason').substring(0, 50).padEnd(54)}║`);
  lines.push(`║ 凭证: ${getString(expense, 'voucherFileName').padEnd(54)}║`);

  const supplementaryInfo = getString(expense, 'supplementaryInfo');
  if (supplementaryInfo) {
    const chunks = supplementaryInfo.match(/.{1,50}/g) || [supplementaryInfo];
    lines.push('║────────────────────────────────────────────────────────────║');
    lines.push('║ 补充说明:                                                    ║');
    for (const chunk of chunks) {
      lines.push(`║   ${chunk.padEnd(52)}║`);
    }
  }

  lines.push('╠════════════════════════════════════════════════════════════╣');
  lines.push(`║ 创建时间: ${formatDate(getString(expense, 'createdAt')).padEnd(48)}║`);
  lines.push(`║ 更新时间: ${formatDate(getString(expense, 'updatedAt')).padEnd(48)}║`);
  lines.push('╚════════════════════════════════════════════════════════════╝');

  return lines.join('\n');
}

function formatExpenseList(items: Array<Record<string, unknown>>): string {
  if (items.length === 0) {
    return '暂无数据';
  }

  const lines: string[] = [];
  lines.push('┌─────────────────────────────────────────────────────────────────────────────────────┐');
  lines.push('│ ID                                    │ 日期     │ 金额      │ 类别 │ 状态   │ 员工   │');
  lines.push('├─────────────────────────────────────────────────────────────────────────────────────┤');

  for (const item of items) {
    const id = getString(item, 'id').substring(0, 37).padEnd(37);
    const date = formatShortDate(getString(item, 'date')).padEnd(8);
    const amount = formatAmount(getNumber(item, 'amount')).padEnd(9);
    const category = getCategoryLabel(getString(item, 'category')).padEnd(4);
    const status = getStatusLabel(getString(item, 'status')).padEnd(6);
    const employee = (getString(item, 'employeeName') || getString(item, 'employeeId')).substring(0, 6).padEnd(6);

    const needsSupplementary = getBoolean(item, 'needsSupplementaryInfo');
    const indicator = needsSupplementary ? ' ⚠' : '';

    lines.push(`│ ${id} │ ${date} │ ${amount} │ ${category} │ ${status} │ ${employee} │${indicator}`);
  }

  lines.push('└─────────────────────────────────────────────────────────────────────────────────────┘');
  lines.push('');
  lines.push('⚠ 表示需要补充说明');

  return lines.join('\n');
}

function formatOperationLog(log: Record<string, unknown>): string {
  const lines: string[] = [];
  lines.push('┌────────────────────────────────────────────────────────────────┐');
  lines.push('│                        操作日志                                  │');
  lines.push('├────────────────────────────────────────────────────────────────┤');
  lines.push(`│ 时间: ${formatDate(getString(log, 'createdAt')).padEnd(50)}│`);
  lines.push(`│ 用户: ${(getString(log, 'userName') || getString(log, 'userId')).padEnd(52)}│`);
  lines.push(`│ 操作: ${getString(log, 'action').padEnd(52)}│`);
  lines.push(`│ 类型: ${getString(log, 'targetType').padEnd(52)}│`);
  const targetId = getString(log, 'targetId');
  if (targetId) {
    lines.push(`│ 目标ID: ${targetId.padEnd(50)}│`);
  }
  const details = log['details'];
  if (details) {
    lines.push('│ 详情:');
    try {
      const detailStr = JSON.stringify(details, null, 2);
      const detailLines = detailStr.split('\n');
      for (const dl of detailLines) {
        lines.push(`│   ${dl.substring(0, 50).padEnd(52)}│`);
      }
    } catch {
      lines.push(`│   ${String(details).substring(0, 50).padEnd(52)}│`);
    }
  }
  lines.push('└────────────────────────────────────────────────────────────────┘');
  return lines.join('\n');
}

function formatOperationLogList(items: Array<Record<string, unknown>>): string {
  if (items.length === 0) {
    return '暂无操作日志';
  }

  const lines: string[] = [];
  lines.push('┌──────────────────────────────────────────────────────────────────────────────────────┐');
  lines.push('│ 时间                  │ 用户     │ 操作         │ 类型   │ 目标ID                      │');
  lines.push('├──────────────────────────────────────────────────────────────────────────────────────┤');

  for (const item of items) {
    const time = formatShortDate(getString(item, 'createdAt')).padEnd(20);
    const user = (getString(item, 'userName') || getString(item, 'userId')).substring(0, 8).padEnd(8);
    const action = getString(item, 'action').substring(0, 12).padEnd(12);
    const type = getString(item, 'targetType').padEnd(6);
    const targetId = (getString(item, 'targetId') || '-').substring(0, 26).padEnd(26);

    lines.push(`│ ${time} │ ${user} │ ${action} │ ${type} │ ${targetId} │`);
  }

  lines.push('└──────────────────────────────────────────────────────────────────────────────────────┘');

  return lines.join('\n');
}

function formatAuditLog(log: Record<string, unknown>): string {
  const lines: string[] = [];
  lines.push('┌────────────────────────────────────────────────────────────────┐');
  lines.push('│                        审计日志                                  │');
  lines.push('├────────────────────────────────────────────────────────────────┤');
  lines.push(`│ 时间: ${formatDate(getString(log, 'createdAt')).padEnd(50)}│`);
  lines.push(`│ 用户: ${(getString(log, 'userName') || getString(log, 'userId')).padEnd(52)}│`);
  lines.push(`│ 动作: ${getString(log, 'action').padEnd(52)}│`);
  lines.push(`│ 类型: ${getString(log, 'targetType').padEnd(52)}│`);
  const targetId = getString(log, 'targetId');
  if (targetId) {
    lines.push(`│ 目标ID: ${targetId.padEnd(50)}│`);
  }
  const oldValue = log['oldValue'];
  if (oldValue) {
    lines.push('│ 旧值:');
    try {
      const oldStr = JSON.stringify(oldValue, null, 2);
      const oldLines = oldStr.split('\n');
      for (const ol of oldLines) {
        lines.push(`│   ${ol.substring(0, 50).padEnd(52)}│`);
      }
    } catch {
      lines.push(`│   ${String(oldValue).substring(0, 50).padEnd(52)}│`);
    }
  }
  const newValue = log['newValue'];
  if (newValue) {
    lines.push('│ 新值:');
    try {
      const newStr = JSON.stringify(newValue, null, 2);
      const newLines = newStr.split('\n');
      for (const nl of newLines) {
        lines.push(`│   ${nl.substring(0, 50).padEnd(52)}│`);
      }
    } catch {
      lines.push(`│   ${String(newValue).substring(0, 50).padEnd(52)}│`);
    }
  }
  lines.push('└────────────────────────────────────────────────────────────────┘');
  return lines.join('\n');
}

function formatAuditLogList(items: Array<Record<string, unknown>>): string {
  if (items.length === 0) {
    return '暂无审计日志';
  }

  const lines: string[] = [];
  lines.push('┌─────────────────────────────────────────────────────────────────────────────────────────┐');
  lines.push('│ 时间                  │ 用户     │ 动作               │ 类型   │ 目标ID                      │');
  lines.push('├─────────────────────────────────────────────────────────────────────────────────────────┤');

  for (const item of items) {
    const time = formatShortDate(getString(item, 'createdAt')).padEnd(20);
    const user = (getString(item, 'userName') || getString(item, 'userId')).substring(0, 8).padEnd(8);
    const action = getString(item, 'action').substring(0, 18).padEnd(18);
    const type = getString(item, 'targetType').padEnd(6);
    const targetId = (getString(item, 'targetId') || '-').substring(0, 26).padEnd(26);

    lines.push(`│ ${time} │ ${user} │ ${action} │ ${type} │ ${targetId} │`);
  }

  lines.push('└─────────────────────────────────────────────────────────────────────────────────────────┘');

  return lines.join('\n');
}

function formatError(error: { code: number; message: string; details?: string }): string {
  const lines: string[] = [];
  lines.push('┌────────────────────────────────────────────────────────────┐');
  lines.push('│                        错误信息                              │');
  lines.push('├────────────────────────────────────────────────────────────┤');
  lines.push(`│ 错误码: ${String(error.code).padEnd(48)}│`);
  lines.push(`│ 消息: ${error.message.substring(0, 50).padEnd(50)}│`);
  if (error.details) {
    lines.push('│ 详情:');
    const detailLines = error.details.split('\n');
    for (const dl of detailLines) {
      lines.push(`│   ${dl.substring(0, 48).padEnd(50)}│`);
    }
  }
  lines.push('└────────────────────────────────────────────────────────────┘');
  return lines.join('\n');
}

function formatUser(user: Record<string, unknown>): string {
  const lines: string[] = [];
  lines.push('┌───────────────────────────────────────────┐');
  lines.push('│              用户信息                      │');
  lines.push('├───────────────────────────────────────────┤');
  lines.push(`│ 用户ID: ${getString(user, 'id').padEnd(35)}│`);
  lines.push(`│ 姓名: ${getString(user, 'name').padEnd(37)}│`);
  lines.push(`│ 角色: ${getRoleLabel(getString(user, 'role')).padEnd(37)}│`);
  const department = getString(user, 'department');
  if (department) {
    lines.push(`│ 部门: ${department.padEnd(37)}│`);
  }
  lines.push('└───────────────────────────────────────────┘');
  return lines.join('\n');
}

function formatHelp(): string {
  const lines: string[] = [];
  lines.push('');
  lines.push('╔══════════════════════════════════════════════════════════════════════════════╗');
  lines.push('║                           报销管理系统 CLI 帮助                                ║');
  lines.push('╠══════════════════════════════════════════════════════════════════════════════╣');
  lines.push('║                                                                                  ║');
  lines.push('║ 全局选项:                                                                       ║');
  lines.push('║   --user <userId>    指定操作用户ID (可选，部分命令需要)                      ║');
  lines.push('║   --help, -h         显示帮助信息                                              ║');
  lines.push('║                                                                                  ║');
  lines.push('║ 员工命令:                                                                       ║');
  lines.push('║   create               创建报销单                                               ║');
  lines.push('║     --employeeId <id>     员工ID                                               ║');
  lines.push('║     --employeeName <name> 员工姓名                                             ║');
  lines.push('║     --date <YYYY-MM-DD>  报销日期                                             ║');
  lines.push('║     --amount <yuan>      金额 (支持小数，如 100.50 表示100元50分)            ║');
  lines.push('║     --category <cat>     类别: transport|dining|accommodation|office|other  ║');
  lines.push('║     --reason <text>      事由                                                 ║');
  lines.push('║     --voucher <file>     凭证文件名                                           ║');
  lines.push('║                                                                                  ║');
  lines.push('║   update               更新报销单                                               ║');
  lines.push('║     --id <expenseId>       报销单ID                                           ║');
  lines.push('║     --date <YYYY-MM-DD>     新日期 (可选)                                     ║');
  lines.push('║     --amount <yuan>         新金额 (可选)                                     ║');
  lines.push('║     --category <cat>        新类别 (可选)                                     ║');
  lines.push('║     --reason <text>         新事由 (可选)                                     ║');
  lines.push('║     --voucher <file>        新凭证文件名 (可选)                                ║');
  lines.push('║     --supplementary <text>  补充说明 (可选，用于超出限额的情况)                ║');
  lines.push('║                                                                                  ║');
  lines.push('║   submit               提交报销单                                               ║');
  lines.push('║     --id <expenseId>       报销单ID                                           ║');
  lines.push('║                                                                                  ║');
  lines.push('║   resubmit             重新提交已驳回的报销单                                   ║');
  lines.push('║     --id <expenseId>       报销单ID                                           ║');
  lines.push('║                                                                                  ║');
  lines.push('║ 管理员命令:                                                                     ║');
  lines.push('║   approve              审批通过                                                 ║');
  lines.push('║     --id <expenseId>       报销单ID                                           ║');
  lines.push('║     --comment <text>       审批意见 (可选)                                     ║');
  lines.push('║                                                                                  ║');
  lines.push('║   reject               驳回                                                     ║');
  lines.push('║     --id <expenseId>       报销单ID                                           ║');
  lines.push('║     --comment <text>       驳回原因 (可选)                                     ║');
  lines.push('║                                                                                  ║');
  lines.push('║   batch-approve        批量审批                                                 ║');
  lines.push('║     --ids <id1,id2,...>    报销单ID列表，逗号分隔                              ║');
  lines.push('║     --comment <text>       审批意见 (可选)                                     ║');
  lines.push('║                                                                                  ║');
  lines.push('║   pay                  打款确认                                                 ║');
  lines.push('║     --id <expenseId>       报销单ID                                           ║');
  lines.push('║                                                                                  ║');
  lines.push('║   delete               删除报销单                                               ║');
  lines.push('║     --id <expenseId>       报销单ID                                           ║');
  lines.push('║                                                                                  ║');
  lines.push('║ 查询命令:                                                                       ║');
  lines.push('║   get                  获取单个报销单                                           ║');
  lines.push('║     --id <expenseId>       报销单ID                                           ║');
  lines.push('║                                                                                  ║');
  lines.push('║   list                 列出报销单                                               ║');
  lines.push('║     --start-date <date>     开始日期 (可选)                                    ║');
  lines.push('║     --end-date <date>       结束日期 (可选)                                    ║');
  lines.push('║     --status <status>       状态过滤 (可选): pending|approved|paid|rejected   ║');
  lines.push('║     --employee-id <id>      员工ID过滤 (可选)                                  ║');
  lines.push('║     --page <num>            页码 (可选，默认1)                                 ║');
  lines.push('║     --page-size <num>       每页数量 (可选，默认20)                            ║');
  lines.push('║                                                                                  ║');
  lines.push('║   logs                 查看操作日志                                             ║');
  lines.push('║     --start-date <date>     开始日期 (可选)                                    ║');
  lines.push('║     --end-date <date>       结束日期 (可选)                                    ║');
  lines.push('║     --user-id <id>          用户ID过滤 (可选)                                  ║');
  lines.push('║     --page <num>            页码 (可选)                                        ║');
  lines.push('║                                                                                  ║');
  lines.push('║   audit                查看审计日志                                             ║');
  lines.push('║     --start-date <date>     开始日期 (可选)                                    ║');
  lines.push('║     --end-date <date>       结束日期 (可选)                                    ║');
  lines.push('║     --user-id <id>          用户ID过滤 (可选)                                  ║');
  lines.push('║     --action <type>         动作过滤 (可选)                                    ║');
  lines.push('║     --page <num>            页码 (可选)                                        ║');
  lines.push('║                                                                                  ║');
  lines.push('║   user                 获取用户信息                                             ║');
  lines.push('║     --id <userId>           用户ID                                             ║');
  lines.push('║                                                                                  ║');
  lines.push('╠══════════════════════════════════════════════════════════════════════════════╣');
  lines.push('║ 业务规则说明:                                                                    ║');
  lines.push('║   1. 金额单位: 输入时用元(支持小数)，内部按分存储计算                           ║');
  lines.push('║   2. 费用限额: 餐饮单笔上限800元，住宿每晚上限500元                            ║');
  lines.push('║   3. 超出限额: 需补充说明，否则无法提交                                         ║');
  lines.push('║   4. 大额报销: 超过5000元需二级审批(部门经理→财务总监)                          ║');
  lines.push('║   5. 批量审批: 大额单自动跳过，需单独审批                                       ║');
  lines.push('║   6. 驳回处理: 驳回后流程回到待审批，审批记录清零                                ║');
  lines.push('║                                                                                  ║');
  lines.push('║ 预定义用户:                                                                     ║');
  lines.push('║   emp1    张三        员工                                                      ║');
  lines.push('║   emp2    李四        员工                                                      ║');
  lines.push('║   mgr1    王经理      部门经理                                                  ║');
  lines.push('║   mgr2    刘经理      部门经理                                                  ║');
  lines.push('║   fd1     陈总监      财务总监                                                  ║');
  lines.push('║   admin1  系统管理员  管理员                                                    ║');
  lines.push('║                                                                                  ║');
  lines.push('║ 使用示例:                                                                       ║');
  lines.push('║   # 员工创建报销单                                                               ║');
  lines.push('║   expense-cli create --employeeId emp1 --employeeName 张三 \\                   ║');
  lines.push('║     --date 2026-05-05 --amount 500.00 --category transport \\                  ║');
  lines.push('║     --reason "出差" --voucher "receipt.pdf"                                    ║');
  lines.push('║                                                                                  ║');
  lines.push('║   # 经理审批                                                                     ║');
  lines.push('║   expense-cli --user mgr1 approve --id <expenseId>                             ║');
  lines.push('║                                                                                  ║');
  lines.push('║   # 查看待审批列表                                                               ║');
  lines.push('║   expense-cli list --status pending                                             ║');
  lines.push('╚══════════════════════════════════════════════════════════════════════════════╝');
  lines.push('');

  return lines.join('\n');
}

export {
  formatAmount,
  formatDate,
  formatShortDate,
  formatExpense,
  formatExpenseList,
  formatOperationLog,
  formatOperationLogList,
  formatAuditLog,
  formatAuditLogList,
  formatError,
  formatUser,
  formatHelp,
};
