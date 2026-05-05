#!/usr/bin/env node

import {
  handleCustomerCommand,
  handleSubscriptionCommand,
  handleUsageCommand,
  handleBillingCommand,
} from './commands';
import { configureClient } from './api-client';
import { printError, BOLD, RESET } from './formatter';

const args = process.argv.slice(2);

function printHelp(): void {
  console.log(`
${BOLD}SaaS 计费引擎 CLI 工具${RESET}

用法: billing <command> [subcommand] [options]

可用命令:
  customer     客户管理 (create, get, status)
  subscription 订阅管理 (create, upgrade, refund)
  usage        用量管理 (record, query)
  billing      账单管理 (generate, list, show, pay, review)

环境变量:
  BILLING_HOST   服务端主机 (默认: localhost)
  BILLING_PORT   服务端端口 (默认: 3000)

示例:
  billing customer create "张三" zhangsan@example.com
  billing subscription create cus-abc123 pro yearly
  billing usage record cus-abc123 1500
  billing billing list cus-abc123

查看详细帮助:
  billing customer
  billing subscription
  billing usage
  billing billing
`);
}

async function main(): Promise<void> {
  const host = process.env.BILLING_HOST;
  const port = process.env.BILLING_PORT ? parseInt(process.env.BILLING_PORT, 10) : undefined;

  if (host || port) {
    configureClient({ host, port });
  }

  if (args.length === 0) {
    printHelp();
    return;
  }

  const command = args[0];
  const subArgs = args.slice(1);

  try {
    switch (command) {
      case 'customer':
      case 'customers':
        await handleCustomerCommand(subArgs);
        break;
      case 'subscription':
      case 'subscriptions':
        await handleSubscriptionCommand(subArgs);
        break;
      case 'usage':
        await handleUsageCommand(subArgs);
        break;
      case 'billing':
      case 'bills':
        await handleBillingCommand(subArgs);
        break;
      case 'help':
      case '--help':
      case '-h':
        printHelp();
        break;
      default:
        console.error(`未知命令: ${command}`);
        console.error('使用 "billing help" 查看可用命令');
        process.exit(1);
    }
  } catch (error) {
    if (error instanceof Error) {
      const errorObj = error as unknown as Record<string, unknown>;
      const code = errorObj.code as string | undefined;
      if (code) {
        printError(`[${code}] ${error.message}`);
      } else {
        printError(error.message);
      }
      const details = errorObj.details as Record<string, unknown> | undefined;
      if (details && Object.keys(details).length > 0) {
        console.error('详细信息:', JSON.stringify(details, null, 2));
      }
    } else {
      printError(String(error));
    }
    process.exit(1);
  }
}

main().catch((error) => {
  printError(`未捕获的错误: ${(error as Error).message}`);
  process.exit(1);
});
