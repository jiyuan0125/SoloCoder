import { IncomingMessage, ServerResponse } from "http";
import {
  CreateOrderRequest,
  CreateOrderResponse,
  PayOrderRequest,
  PayOrderResponse,
  OrderQueryRequest,
  OrderQueryResponse,
  CreateRefundRequest,
  CreateRefundResponse,
  RefundQueryRequest,
  RefundQueryResponse,
  CloseOrderRequest,
  CloseOrderResponse,
  CallbackRequest,
  CallbackResponse,
  PaymentChannel,
  WechatCallbackData,
  AlipayCallbackData,
  ChannelRefundCallbackData,
} from "@payment-gateway/shared";
import { PaymentGatewayError, ErrorCode } from "@payment-gateway/shared";
import {
  parseJsonBody,
  sendSuccessResponse,
  sendErrorResponse,
  sendNotFoundResponse,
  sendMethodNotAllowedResponse,
  getPathname,
  parseQueryString,
} from "./http-utils";
import { createOrder, queryOrder, closeOrder } from "./services/order-service";
import { payOrder } from "./services/payment-service";
import { createRefund, queryRefund } from "./services/refund-service";
import { processPaymentCallback, processRefundCallback } from "./services/callback-service";
import {
  simulateWechatPaymentCallback,
  simulateAlipayPaymentCallback,
  simulateWechatRefundCallback,
  simulateAlipayRefundCallback,
} from "./channel-mock";

type RouteHandler = (req: IncomingMessage, res: ServerResponse) => void | Promise<void>;

interface Route {
  method: string;
  path: string;
  handler: RouteHandler;
}

const routes: Route[] = [
  { method: "POST", path: "/api/orders", handler: handleCreateOrder },
  { method: "POST", path: "/api/orders/pay", handler: handlePayOrder },
  { method: "GET", path: "/api/orders", handler: handleQueryOrder },
  { method: "POST", path: "/api/orders/close", handler: handleCloseOrder },
  { method: "POST", path: "/api/refunds", handler: handleCreateRefund },
  { method: "GET", path: "/api/refunds", handler: handleQueryRefund },
  { method: "POST", path: "/api/callbacks/payment", handler: handlePaymentCallback },
  { method: "POST", path: "/api/callbacks/refund", handler: handleRefundCallback },
  { method: "POST", path: "/api/test/callback/payment/wechat", handler: handleTestWechatPaymentCallback },
  { method: "POST", path: "/api/test/callback/payment/alipay", handler: handleTestAlipayPaymentCallback },
  { method: "POST", path: "/api/test/callback/refund/wechat", handler: handleTestWechatRefundCallback },
  { method: "POST", path: "/api/test/callback/refund/alipay", handler: handleTestAlipayRefundCallback },
];

export async function handleRequest(req: IncomingMessage, res: ServerResponse): Promise<void> {
  const method = req.method ?? "GET";
  const url = req.url ?? "/";
  const pathname = getPathname(url);

  try {
    const matchedRoute = routes.find((r) => r.path === pathname && r.method === method);

    if (matchedRoute) {
      await matchedRoute.handler(req, res);
      return;
    }

    const matchedPath = routes.find((r) => r.path === pathname);
    if (matchedPath) {
      const allowedMethods = routes.filter((r) => r.path === pathname).map((r) => r.method);
      sendMethodNotAllowedResponse(res, allowedMethods);
      return;
    }

    sendNotFoundResponse(res);
  } catch (error) {
    if (error instanceof PaymentGatewayError) {
      sendErrorResponse(res, error);
    } else {
      const err = new PaymentGatewayError(ErrorCode.INTERNAL_ERROR, String(error));
      sendErrorResponse(res, err);
    }
  }
}

async function handleCreateOrder(req: IncomingMessage, res: ServerResponse): Promise<void> {
  const body = await parseJsonBody<CreateOrderRequest>(req);
  const result = createOrder(body);
  sendSuccessResponse<CreateOrderResponse>(res, result);
}

async function handlePayOrder(req: IncomingMessage, res: ServerResponse): Promise<void> {
  const body = await parseJsonBody<PayOrderRequest>(req);
  const result = payOrder(body);
  sendSuccessResponse<PayOrderResponse>(res, result);
}

async function handleQueryOrder(req: IncomingMessage, res: ServerResponse): Promise<void> {
  const url = req.url ?? "";
  const query = parseQueryString(url);
  const orderNo = query.orderNo;

  if (!orderNo) {
    throw new PaymentGatewayError(ErrorCode.INVALID_PARAMS, "订单号不能为空");
  }

  const request: OrderQueryRequest = { orderNo };
  const result = queryOrder(request);
  sendSuccessResponse<OrderQueryResponse>(res, result);
}

async function handleCloseOrder(req: IncomingMessage, res: ServerResponse): Promise<void> {
  const body = await parseJsonBody<CloseOrderRequest>(req);
  const result = closeOrder(body);
  sendSuccessResponse<CloseOrderResponse>(res, result);
}

async function handleCreateRefund(req: IncomingMessage, res: ServerResponse): Promise<void> {
  const body = await parseJsonBody<CreateRefundRequest>(req);
  const result = createRefund(body);
  sendSuccessResponse<CreateRefundResponse>(res, result);
}

async function handleQueryRefund(req: IncomingMessage, res: ServerResponse): Promise<void> {
  const url = req.url ?? "";
  const query = parseQueryString(url);
  const orderNo = query.orderNo;
  const refundNo = query.refundNo ?? undefined;

  if (!orderNo) {
    throw new PaymentGatewayError(ErrorCode.INVALID_PARAMS, "订单号不能为空");
  }

  const request: RefundQueryRequest = { orderNo, refundNo };
  const result = queryRefund(request);
  sendSuccessResponse<RefundQueryResponse>(res, result);
}

async function handlePaymentCallback(req: IncomingMessage, res: ServerResponse): Promise<void> {
  const body = await parseJsonBody<CallbackRequest>(req);
  const channel = body.channel;
  const data = body.data;

  let callbackData: WechatCallbackData | AlipayCallbackData;
  try {
    callbackData = JSON.parse(data) as WechatCallbackData | AlipayCallbackData;
  } catch {
    throw new PaymentGatewayError(ErrorCode.INVALID_PARAMS, "回调数据格式错误");
  }

  processPaymentCallback(channel, callbackData);

  const response: CallbackResponse = {
    success: true,
    message: "回调处理成功",
  };
  sendSuccessResponse<CallbackResponse>(res, response);
}

async function handleRefundCallback(req: IncomingMessage, res: ServerResponse): Promise<void> {
  const body = await parseJsonBody<{ channel: PaymentChannel; data: string }>(req);
  const channel = body.channel;
  const data = body.data;

  let callbackData: ChannelRefundCallbackData;
  try {
    callbackData = JSON.parse(data) as ChannelRefundCallbackData;
  } catch {
    throw new PaymentGatewayError(ErrorCode.INVALID_PARAMS, "回调数据格式错误");
  }

  processRefundCallback(channel, callbackData);

  const response: CallbackResponse = {
    success: true,
    message: "回调处理成功",
  };
  sendSuccessResponse<CallbackResponse>(res, response);
}

async function handleTestWechatPaymentCallback(req: IncomingMessage, res: ServerResponse): Promise<void> {
  const body = await parseJsonBody<{ orderNo: string; amount: number; success?: boolean }>(req);
  const callbackData = simulateWechatPaymentCallback(body.orderNo, body.amount, body.success ?? true);

  const callbackRequest: CallbackRequest = {
    channel: PaymentChannel.WECHAT,
    data: JSON.stringify(callbackData),
  };

  try {
    processPaymentCallback(PaymentChannel.WECHAT, callbackData);
    const response: CallbackResponse = {
      success: true,
      message: "微信支付回调模拟成功",
    };
    sendSuccessResponse<CallbackResponse>(res, response);
  } catch (error) {
    if (error instanceof PaymentGatewayError && error.code === ErrorCode.CALLBACK_DUPLICATE) {
      const response: CallbackResponse = {
        success: true,
        message: "回调已处理(幂等)",
      };
      sendSuccessResponse<CallbackResponse>(res, response);
    } else {
      throw error;
    }
  }
}

async function handleTestAlipayPaymentCallback(req: IncomingMessage, res: ServerResponse): Promise<void> {
  const body = await parseJsonBody<{ orderNo: string; amount: number; success?: boolean }>(req);
  const callbackData = simulateAlipayPaymentCallback(body.orderNo, body.amount, body.success ?? true);

  try {
    processPaymentCallback(PaymentChannel.ALIPAY, callbackData);
    const response: CallbackResponse = {
      success: true,
      message: "支付宝支付回调模拟成功",
    };
    sendSuccessResponse<CallbackResponse>(res, response);
  } catch (error) {
    if (error instanceof PaymentGatewayError && error.code === ErrorCode.CALLBACK_DUPLICATE) {
      const response: CallbackResponse = {
        success: true,
        message: "回调已处理(幂等)",
      };
      sendSuccessResponse<CallbackResponse>(res, response);
    } else {
      throw error;
    }
  }
}

async function handleTestWechatRefundCallback(req: IncomingMessage, res: ServerResponse): Promise<void> {
  const body = await parseJsonBody<{ orderNo: string; refundNo: string; amount: number; success?: boolean }>(req);
  const callbackData = simulateWechatRefundCallback(body.orderNo, body.refundNo, body.amount, body.success ?? true);

  try {
    processRefundCallback(PaymentChannel.WECHAT, callbackData);
    const response: CallbackResponse = {
      success: true,
      message: "微信退款回调模拟成功",
    };
    sendSuccessResponse<CallbackResponse>(res, response);
  } catch (error) {
    if (error instanceof PaymentGatewayError && error.code === ErrorCode.CALLBACK_DUPLICATE) {
      const response: CallbackResponse = {
        success: true,
        message: "回调已处理(幂等)",
      };
      sendSuccessResponse<CallbackResponse>(res, response);
    } else {
      throw error;
    }
  }
}

async function handleTestAlipayRefundCallback(req: IncomingMessage, res: ServerResponse): Promise<void> {
  const body = await parseJsonBody<{ orderNo: string; refundNo: string; amount: number; success?: boolean }>(req);
  const callbackData = simulateAlipayRefundCallback(body.orderNo, body.refundNo, body.amount, body.success ?? true);

  try {
    processRefundCallback(PaymentChannel.ALIPAY, callbackData);
    const response: CallbackResponse = {
      success: true,
      message: "支付宝退款回调模拟成功",
    };
    sendSuccessResponse<CallbackResponse>(res, response);
  } catch (error) {
    if (error instanceof PaymentGatewayError && error.code === ErrorCode.CALLBACK_DUPLICATE) {
      const response: CallbackResponse = {
        success: true,
        message: "回调已处理(幂等)",
      };
      sendSuccessResponse<CallbackResponse>(res, response);
    } else {
      throw error;
    }
  }
}
