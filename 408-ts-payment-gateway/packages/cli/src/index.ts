#!/usr/bin/env node

import { post, get } from "./api-client";
import {
  CreateOrderRequest,
  CreateOrderResponse,
  PayOrderRequest,
  PayOrderResponse,
  OrderQueryResponse,
  CreateRefundRequest,
  CreateRefundResponse,
  RefundQueryResponse,
  CloseOrderRequest,
  CloseOrderResponse,
  CallbackResponse,
  ErrorCode,
} from "@payment-gateway/shared";
import {
  formatCreateOrderResponse,
  formatPayOrderResponse,
  formatOrderQueryResponse,
  formatCreateRefundResponse,
  formatRefundQueryResponse,
  formatCloseOrderResponse,
  formatCallbackResponse,
  formatApiResponse,
  formatHelp,
} from "./formatter";

interface ParsedArgs {
  command: string;
  positional: string[];
  flags: Record<string, boolean>;
}

function parseArgs(): ParsedArgs {
  const args = process.argv.slice(2);
  const positional: string[] = [];
  const flags: Record<string, boolean> = {};

  for (const arg of args) {
    if (arg.startsWith("--")) {
      flags[arg.slice(2)] = true;
    } else {
      positional.push(arg);
    }
  }

  const command = positional[0] ?? "help";

  return {
    command,
    positional: positional.slice(1),
    flags,
  };
}

async function handleCreateOrder(args: ParsedArgs): Promise<void> {
  if (args.positional.length < 2) {
    console.error("用法: payment-gateway create-order <orderNo> <amount>");
    process.exit(1);
  }

  const orderNo = args.positional[0];
  const amount = parseInt(args.positional[1], 10);

  if (isNaN(amount)) {
    console.error("金额必须是整数(单位:分)");
    process.exit(1);
  }

  const request: CreateOrderRequest = { orderNo, amount };
  const response = await post<CreateOrderResponse>("/api/orders", request);

  if (response.code === ErrorCode.SUCCESS && response.data) {
    console.log(formatCreateOrderResponse(response.data));
  } else {
    console.error(formatApiResponse(response));
    process.exit(1);
  }
}

async function handlePayOrder(args: ParsedArgs): Promise<void> {
  if (args.positional.length < 1) {
    console.error("用法: payment-gateway pay-order <orderNo>");
    process.exit(1);
  }

  const orderNo = args.positional[0];
  const request: PayOrderRequest = { orderNo };
  const response = await post<PayOrderResponse>("/api/orders/pay", request);

  if (response.code === ErrorCode.SUCCESS && response.data) {
    console.log(formatPayOrderResponse(response.data));
  } else {
    console.error(formatApiResponse(response));
    process.exit(1);
  }
}

async function handleQueryOrder(args: ParsedArgs): Promise<void> {
  if (args.positional.length < 1) {
    console.error("用法: payment-gateway query-order <orderNo>");
    process.exit(1);
  }

  const orderNo = args.positional[0];
  const response = await get<OrderQueryResponse>(`/api/orders?orderNo=${encodeURIComponent(orderNo)}`);

  if (response.code === ErrorCode.SUCCESS && response.data) {
    console.log(formatOrderQueryResponse(response.data));
  } else {
    console.error(formatApiResponse(response));
    process.exit(1);
  }
}

async function handleCloseOrder(args: ParsedArgs): Promise<void> {
  if (args.positional.length < 1) {
    console.error("用法: payment-gateway close-order <orderNo>");
    process.exit(1);
  }

  const orderNo = args.positional[0];
  const request: CloseOrderRequest = { orderNo };
  const response = await post<CloseOrderResponse>("/api/orders/close", request);

  if (response.code === ErrorCode.SUCCESS && response.data) {
    console.log(formatCloseOrderResponse(response.data));
  } else {
    console.error(formatApiResponse(response));
    process.exit(1);
  }
}

async function handleCreateRefund(args: ParsedArgs): Promise<void> {
  if (args.positional.length < 3) {
    console.error("用法: payment-gateway create-refund <orderNo> <refundNo> <amount> [--final]");
    process.exit(1);
  }

  const orderNo = args.positional[0];
  const refundNo = args.positional[1];
  const amount = parseInt(args.positional[2], 10);
  const isFinalRefund = args.flags.final ?? false;

  if (isNaN(amount)) {
    console.error("金额必须是整数(单位:分)");
    process.exit(1);
  }

  const request: CreateRefundRequest = {
    orderNo,
    refundNo,
    amount,
    isFinalRefund,
  };
  const response = await post<CreateRefundResponse>("/api/refunds", request);

  if (response.code === ErrorCode.SUCCESS && response.data) {
    console.log(formatCreateRefundResponse(response.data));
  } else {
    console.error(formatApiResponse(response));
    process.exit(1);
  }
}

async function handleQueryRefund(args: ParsedArgs): Promise<void> {
  if (args.positional.length < 1) {
    console.error("用法: payment-gateway query-refund <orderNo> [refundNo]");
    process.exit(1);
  }

  const orderNo = args.positional[0];
  const refundNo = args.positional[1];

  let path = `/api/refunds?orderNo=${encodeURIComponent(orderNo)}`;
  if (refundNo) {
    path += `&refundNo=${encodeURIComponent(refundNo)}`;
  }

  const response = await get<RefundQueryResponse>(path);

  if (response.code === ErrorCode.SUCCESS && response.data) {
    console.log(formatRefundQueryResponse(response.data));
  } else {
    console.error(formatApiResponse(response));
    process.exit(1);
  }
}

async function handleCallbackWechat(args: ParsedArgs): Promise<void> {
  if (args.positional.length < 2) {
    console.error("用法: payment-gateway callback-wechat <orderNo> <amount> [--success|--failed]");
    process.exit(1);
  }

  const orderNo = args.positional[0];
  const amount = parseInt(args.positional[1], 10);
  const success = args.flags.failed ? false : true;

  if (isNaN(amount)) {
    console.error("金额必须是整数(单位:分)");
    process.exit(1);
  }

  const request = { orderNo, amount, success };
  const response = await post<CallbackResponse>("/api/test/callback/payment/wechat", request);

  if (response.code === ErrorCode.SUCCESS && response.data) {
    console.log(formatCallbackResponse(response.data));
  } else {
    console.error(formatApiResponse(response));
    process.exit(1);
  }
}

async function handleCallbackAlipay(args: ParsedArgs): Promise<void> {
  if (args.positional.length < 2) {
    console.error("用法: payment-gateway callback-alipay <orderNo> <amount> [--success|--failed]");
    process.exit(1);
  }

  const orderNo = args.positional[0];
  const amount = parseInt(args.positional[1], 10);
  const success = args.flags.failed ? false : true;

  if (isNaN(amount)) {
    console.error("金额必须是整数(单位:分)");
    process.exit(1);
  }

  const request = { orderNo, amount, success };
  const response = await post<CallbackResponse>("/api/test/callback/payment/alipay", request);

  if (response.code === ErrorCode.SUCCESS && response.data) {
    console.log(formatCallbackResponse(response.data));
  } else {
    console.error(formatApiResponse(response));
    process.exit(1);
  }
}

async function handleCallbackRefundWechat(args: ParsedArgs): Promise<void> {
  if (args.positional.length < 3) {
    console.error("用法: payment-gateway callback-refund-wechat <orderNo> <refundNo> <amount> [--success|--failed]");
    process.exit(1);
  }

  const orderNo = args.positional[0];
  const refundNo = args.positional[1];
  const amount = parseInt(args.positional[2], 10);
  const success = args.flags.failed ? false : true;

  if (isNaN(amount)) {
    console.error("金额必须是整数(单位:分)");
    process.exit(1);
  }

  const request = { orderNo, refundNo, amount, success };
  const response = await post<CallbackResponse>("/api/test/callback/refund/wechat", request);

  if (response.code === ErrorCode.SUCCESS && response.data) {
    console.log(formatCallbackResponse(response.data));
  } else {
    console.error(formatApiResponse(response));
    process.exit(1);
  }
}

async function handleCallbackRefundAlipay(args: ParsedArgs): Promise<void> {
  if (args.positional.length < 3) {
    console.error("用法: payment-gateway callback-refund-alipay <orderNo> <refundNo> <amount> [--success|--failed]");
    process.exit(1);
  }

  const orderNo = args.positional[0];
  const refundNo = args.positional[1];
  const amount = parseInt(args.positional[2], 10);
  const success = args.flags.failed ? false : true;

  if (isNaN(amount)) {
    console.error("金额必须是整数(单位:分)");
    process.exit(1);
  }

  const request = { orderNo, refundNo, amount, success };
  const response = await post<CallbackResponse>("/api/test/callback/refund/alipay", request);

  if (response.code === ErrorCode.SUCCESS && response.data) {
    console.log(formatCallbackResponse(response.data));
  } else {
    console.error(formatApiResponse(response));
    process.exit(1);
  }
}

async function main(): Promise<void> {
  const args = parseArgs();

  try {
    switch (args.command) {
      case "create-order":
        await handleCreateOrder(args);
        break;
      case "pay-order":
        await handlePayOrder(args);
        break;
      case "query-order":
        await handleQueryOrder(args);
        break;
      case "close-order":
        await handleCloseOrder(args);
        break;
      case "create-refund":
        await handleCreateRefund(args);
        break;
      case "query-refund":
        await handleQueryRefund(args);
        break;
      case "callback-wechat":
        await handleCallbackWechat(args);
        break;
      case "callback-alipay":
        await handleCallbackAlipay(args);
        break;
      case "callback-refund-wechat":
        await handleCallbackRefundWechat(args);
        break;
      case "callback-refund-alipay":
        await handleCallbackRefundAlipay(args);
        break;
      case "help":
      default:
        console.log(formatHelp());
        break;
    }
  } catch (error) {
    const err = error as Error;
    console.error(`\x1b[31m✗ 错误: ${err.message}\x1b[0m`);
    if (err.message.includes("ECONNREFUSED")) {
      console.error("  请确保服务端已启动: npm run start");
    }
    process.exit(1);
  }
}

main();
