import {
  IncomingMessage,
  ServerResponse,
} from 'http';
import {
  CreateOrderRequest,
  Order,
  OrderStatus,
  SalesContribution,
  successResponse,
  errorResponse,
  ErrorCodes,
  getErrorMessage,
} from '@commission-tracker/shared';
import { store } from '../store';
import { validateSalesContributions } from '../services/teamAllocation';
import { handleRefund } from '../services/settlementService';
import { parseRequestBody, sendJsonResponse } from '../utils';

function generateOrderId(): string {
  const timestamp = Date.now().toString(36);
  const random = Math.random().toString(36).substring(2, 6);
  return `ORD-${timestamp}-${random}`;
}

export async function handleCreateOrder(
  req: IncomingMessage,
  res: ServerResponse
): Promise<void> {
  try {
    const body = await parseRequestBody(req);
    const request: CreateOrderRequest = JSON.parse(body);

    if (!request.amount || request.amount <= 0) {
      sendJsonResponse(
        res,
        400,
        errorResponse(ErrorCodes.INVALID_AMOUNT, '订单金额必须大于0')
      );
      return;
    }

    if (!request.date) {
      sendJsonResponse(
        res,
        400,
        errorResponse(ErrorCodes.INVALID_DATE, '缺少订单日期')
      );
      return;
    }

    const orderDate = new Date(request.date);
    if (isNaN(orderDate.getTime())) {
      sendJsonResponse(
        res,
        400,
        errorResponse(ErrorCodes.INVALID_DATE, '日期格式无效')
      );
      return;
    }

    if (!request.salesContributions || request.salesContributions.length === 0) {
      sendJsonResponse(
        res,
        400,
        errorResponse(ErrorCodes.INVALID_SALES_CONTRIBUTION, '订单必须关联至少一个销售')
      );
      return;
    }

    for (const contrib of request.salesContributions) {
      const salesperson = store.getSalesperson(contrib.salespersonId);
      if (!salesperson) {
        sendJsonResponse(
          res,
          404,
          errorResponse(ErrorCodes.SALESPERSON_NOT_FOUND, `销售 ${contrib.salespersonId} 不存在`)
        );
        return;
      }
      if (salesperson.status === 'resigned') {
        sendJsonResponse(
          res,
          400,
          errorResponse(ErrorCodes.SALESPERSON_ALREADY_RESIGNED, `销售 ${salesperson.name} 已离职，无法创建新订单`)
        );
        return;
      }
    }

    try {
      validateSalesContributions(request.salesContributions);
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : '销售贡献分配无效';
      sendJsonResponse(
        res,
        400,
        errorResponse(ErrorCodes.INVALID_SALES_CONTRIBUTION, message)
      );
      return;
    }

    const order: Order = {
      id: generateOrderId(),
      amount: request.amount,
      date: request.date,
      status: 'active',
      locked: false,
      salesContributions: request.salesContributions,
    };

    store.saveOrder(order);

    sendJsonResponse(res, 201, successResponse(order));
  } catch (error) {
    sendJsonResponse(
      res,
      500,
      errorResponse(ErrorCodes.INTERNAL_ERROR, getErrorMessage(ErrorCodes.INTERNAL_ERROR))
    );
  }
}

export async function handleGetOrder(
  req: IncomingMessage,
  res: ServerResponse
): Promise<void> {
  try {
    const url = new URL(req.url || '', `http://${req.headers.host}`);
    const id = url.pathname.split('/').pop();

    if (!id) {
      sendJsonResponse(
        res,
        400,
        errorResponse(ErrorCodes.INVALID_INPUT, '缺少订单ID')
      );
      return;
    }

    const order = store.getOrder(id);

    if (!order) {
      sendJsonResponse(
        res,
        404,
        errorResponse(ErrorCodes.ORDER_NOT_FOUND, getErrorMessage(ErrorCodes.ORDER_NOT_FOUND))
      );
      return;
    }

    sendJsonResponse(res, 200, successResponse(order));
  } catch (error) {
    sendJsonResponse(
      res,
      500,
      errorResponse(ErrorCodes.INTERNAL_ERROR, getErrorMessage(ErrorCodes.INTERNAL_ERROR))
    );
  }
}

export async function handleListOrders(
  req: IncomingMessage,
  res: ServerResponse
): Promise<void> {
  try {
    const url = new URL(req.url || '', `http://${req.headers.host}`);
    const salespersonId = url.searchParams.get('salespersonId');
    const fromDate = url.searchParams.get('fromDate');
    const toDate = url.searchParams.get('toDate');

    let orders = store.getOrders();

    if (salespersonId) {
      orders = orders.filter(o =>
        o.salesContributions.some((c: SalesContribution) => c.salespersonId === salespersonId)
      );
    }

    if (fromDate) {
      orders = orders.filter(o => o.date >= fromDate);
    }

    if (toDate) {
      orders = orders.filter(o => o.date <= toDate);
    }

    sendJsonResponse(res, 200, successResponse(orders));
  } catch (error) {
    sendJsonResponse(
      res,
      500,
      errorResponse(ErrorCodes.INTERNAL_ERROR, getErrorMessage(ErrorCodes.INTERNAL_ERROR))
    );
  }
}

export async function handleRefundOrder(
  req: IncomingMessage,
  res: ServerResponse
): Promise<void> {
  try {
    const url = new URL(req.url || '', `http://${req.headers.host}`);
    const pathParts = url.pathname.split('/');
    const id = pathParts[pathParts.length - 2];

    if (!id) {
      sendJsonResponse(
        res,
        400,
        errorResponse(ErrorCodes.INVALID_INPUT, '缺少订单ID')
      );
      return;
    }

    const order = store.getOrder(id);

    if (!order) {
      sendJsonResponse(
        res,
        404,
        errorResponse(ErrorCodes.ORDER_NOT_FOUND, getErrorMessage(ErrorCodes.ORDER_NOT_FOUND))
      );
      return;
    }

    if (order.status === 'refunded') {
      sendJsonResponse(
        res,
        400,
        errorResponse(ErrorCodes.ORDER_ALREADY_REFUNDED, getErrorMessage(ErrorCodes.ORDER_ALREADY_REFUNDED))
      );
      return;
    }

    order.status = 'refunded';
    store.saveOrder(order);

    handleRefund(order);

    sendJsonResponse(res, 200, successResponse(order));
  } catch (error) {
    sendJsonResponse(
      res,
      500,
      errorResponse(ErrorCodes.INTERNAL_ERROR, getErrorMessage(ErrorCodes.INTERNAL_ERROR))
    );
  }
}
