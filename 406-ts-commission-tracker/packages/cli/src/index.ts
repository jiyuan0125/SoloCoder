#!/usr/bin/env node

import {
  handleCreateSalesperson,
  handleGetSalesperson,
  handleListSalespeople,
  handleResignSalesperson,
  handleGetBalance,
  handleGetCommission,
  handleCreateOrder,
  handleGetOrder,
  handleListOrders,
  handleRefundOrder,
  handleTriggerSettlement,
  handleListSettlements,
  handleGetSettlement,
  handleGetRankings,
} from './commands';

function printHelp(): void {
  console.log(`
销售佣金追踪系统 CLI

用法: commission <命令> [子命令] [选项]

命令:
  salesperson  销售管理
    create <姓名> <入职日期>     创建新销售
    get <销售ID>                  获取销售详情
    list                          列出所有销售
    resign <销售ID>               销售离职
    balance <销售ID>              查看佣金余额
    commission <销售ID> [月份] [年份]  查看佣金计算

  order        订单管理
    create <金额> <日期> <销售贡献...>  创建订单
      销售贡献格式: <销售ID>[:primary]
      示例: SP-001:primary SP-002
    get <订单ID>                   获取订单详情
    list [销售ID] [开始日期] [结束日期]  列出订单
    refund <订单ID>                订单退款

  settlement   结算管理
    trigger [月份] [年份]          触发月度结算
    list [销售ID] [年份] [月份]    列出结算记录
    get <结算单ID>                 获取结算单详情

  ranking      业绩排名
    monthly [月份] [年份]          月度排名
    quarterly [季度] [年份]        季度排名
    yearly [年份]                  年度排名

示例:
  commission salesperson create "张三" "2026-01-01"
  commission salesperson list
  commission order create 1000000 "2026-05-01" SP-xxx:primary
  commission settlement trigger 4 2026
  commission ranking monthly 5 2026
`);
}

async function main(): Promise<void> {
  const args = process.argv.slice(2);

  if (args.length === 0 || args[0] === 'help' || args[0] === '--help' || args[0] === '-h') {
    printHelp();
    return;
  }

  const command = args[0];
  const subCommand = args[1];
  const restArgs = args.slice(2);

  try {
    switch (command) {
      case 'salesperson':
      case 'sp':
        switch (subCommand) {
          case 'create':
            await handleCreateSalesperson(restArgs);
            break;
          case 'get':
            await handleGetSalesperson(restArgs);
            break;
          case 'list':
            await handleListSalespeople();
            break;
          case 'resign':
            await handleResignSalesperson(restArgs);
            break;
          case 'balance':
            await handleGetBalance(restArgs);
            break;
          case 'commission':
            await handleGetCommission(restArgs);
            break;
          default:
            console.log('salesperson 子命令: create, get, list, resign, balance, commission');
        }
        break;

      case 'order':
      case 'o':
        switch (subCommand) {
          case 'create':
            await handleCreateOrder(restArgs);
            break;
          case 'get':
            await handleGetOrder(restArgs);
            break;
          case 'list':
            await handleListOrders(restArgs);
            break;
          case 'refund':
            await handleRefundOrder(restArgs);
            break;
          default:
            console.log('order 子命令: create, get, list, refund');
        }
        break;

      case 'settlement':
      case 's':
        switch (subCommand) {
          case 'trigger':
            await handleTriggerSettlement(restArgs);
            break;
          case 'list':
            await handleListSettlements(restArgs);
            break;
          case 'get':
            await handleGetSettlement(restArgs);
            break;
          default:
            console.log('settlement 子命令: trigger, list, get');
        }
        break;

      case 'ranking':
      case 'rank':
      case 'r':
        await handleGetRankings(args.slice(1));
        break;

      default:
        console.log(`未知命令: ${command}`);
        printHelp();
    }
  } catch (error: unknown) {
    const message = error instanceof Error ? error.message : String(error);
    console.error('执行出错:', message);
    process.exit(1);
  }
}

main().catch((error: unknown) => {
  const message = error instanceof Error ? error.message : String(error);
  console.error('执行出错:', message);
  process.exit(1);
});
