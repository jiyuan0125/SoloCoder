import http from 'http';
import { IncomingMessage, ServerResponse } from 'http';
import { ApiResponse, RefundStatus, ErrorCodes, OrderProduct, ProductType } from '@refund/shared';
import * as service from './service.js';

interface ParsedRequest {
  method: string;
  pathname: string;
  query: Record<string, string>;
  body: unknown;
}

type RequestHandler = (req: IncomingMessage, res: ServerResponse, parsed: ParsedRequest) => Promise<void>;

async function parseRequest(req: IncomingMessage): Promise<ParsedRequest> {
  const url = new URL(req.url ?? '', `http://${req.headers.host}`);
  
  let body: unknown = null;
  if (req.method === 'POST' || req.method === 'PUT' || req.method === 'PATCH') {
    const buffers: Buffer[] = [];
    for await (const chunk of req) {
      buffers.push(chunk as Buffer);
    }
    const rawBody = Buffer.concat(buffers).toString();
    if (rawBody) {
      try {
        body = JSON.parse(rawBody);
      } catch {
        body = null;
      }
    }
  }

  const query: Record<string, string> = {};
  url.searchParams.forEach((value, key) => {
    query[key] = value;
  });

  return {
    method: req.method ?? 'GET',
    pathname: url.pathname,
    query,
    body
  };
}

function sendJsonResponse(res: ServerResponse, statusCode: number, data: ApiResponse): void {
  res.statusCode = statusCode;
  res.setHeader('Content-Type', 'application/json');
  res.end(JSON.stringify(data));
}

function sendSuccess<T>(res: ServerResponse, data: T): void {
  sendJsonResponse(res, 200, { success: true, data });
}

function sendError(res: ServerResponse, code: typeof ErrorCodes[keyof typeof ErrorCodes], message: string, statusCode: number = 400): void {
  sendJsonResponse(res, statusCode, {
    success: false,
    error: { code, message }
  });
}

function validateBody(body: unknown, requiredFields: string[]): { valid: boolean; missing: string[] } {
  if (!body || typeof body !== 'object') {
    return { valid: false, missing: requiredFields };
  }
  const obj = body as Record<string, unknown>;
  const missing = requiredFields.filter(field => obj[field] === undefined);
  return { valid: missing.length === 0, missing };
}

const createRefundHandler: RequestHandler = async (_req, res, parsed) => {
  const body = parsed.body as Record<string, unknown>;
  
  const validation = validateBody(body, ['orderId', 'userId', 'refundReason', 'productIds']);
  if (!validation.valid) {
    sendError(res, 'INVALID_PARAMETER', `Missing required fields: ${validation.missing.join(', ')}`);
    return;
  }

  const result = service.createRefund(
    body.orderId as string,
    body.userId as string,
    body.refundReason as string,
    body.productIds as string[]
  );

  if ('code' in result) {
    sendError(res, result.code, result.message);
  } else {
    sendSuccess(res, result);
  }
};

const submitLogisticsHandler: RequestHandler = async (_req, res, parsed) => {
  const body = parsed.body as Record<string, unknown>;
  
  const validation = validateBody(body, ['refundId', 'logisticsNumber']);
  if (!validation.valid) {
    sendError(res, 'INVALID_PARAMETER', `Missing required fields: ${validation.missing.join(', ')}`);
    return;
  }

  const result = service.submitLogistics(
    body.refundId as string,
    body.logisticsNumber as string
  );

  if ('code' in result) {
    sendError(res, result.code, result.message);
  } else {
    sendSuccess(res, result);
  }
};

const warehouseConfirmHandler: RequestHandler = async (_req, res, parsed) => {
  const body = parsed.body as Record<string, unknown>;
  
  const validation = validateBody(body, ['refundId', 'received']);
  if (!validation.valid) {
    sendError(res, 'INVALID_PARAMETER', `Missing required fields: ${validation.missing.join(', ')}`);
    return;
  }

  const result = service.warehouseConfirm(
    body.refundId as string,
    body.received as boolean
  );

  if ('code' in result) {
    sendError(res, result.code, result.message);
  } else {
    sendSuccess(res, result);
  }
};

const getRefundHandler: RequestHandler = async (_req, res, parsed) => {
  const refundId = parsed.pathname.split('/').pop();
  
  if (!refundId) {
    sendError(res, 'INVALID_PARAMETER', 'Refund ID is required');
    return;
  }

  const result = service.getRefundById(refundId);

  if ('code' in result) {
    sendError(res, result.code, result.message, 404);
  } else {
    sendSuccess(res, result);
  }
};

const queryRefundsHandler: RequestHandler = async (_req, res, parsed) => {
  const query = parsed.query;

  const options: service.QueryOptions = {};
  
  if (query.page) options.page = parseInt(query.page, 10);
  if (query.pageSize) options.pageSize = parseInt(query.pageSize, 10);
  if (query.status) options.status = query.status as RefundStatus;
  if (query.startTime) options.startTime = new Date(query.startTime);
  if (query.endTime) options.endTime = new Date(query.endTime);
  if (query.orderId) options.orderId = query.orderId;

  const result = service.queryRefunds(options);
  sendSuccess(res, result);
};

const updateConfigHandler: RequestHandler = async (_req, res, parsed) => {
  const body = parsed.body as Record<string, unknown>;

  const updates: { refundPeriodDays?: number; virtualProductRefundThreshold?: number } = {};
  
  if (body.refundPeriodDays !== undefined) {
    updates.refundPeriodDays = body.refundPeriodDays as number;
  }
  if (body.virtualProductRefundThreshold !== undefined) {
    updates.virtualProductRefundThreshold = body.virtualProductRefundThreshold as number;
  }

  const result = service.updateServiceConfig(updates);
  sendSuccess(res, result);
};

const getConfigHandler: RequestHandler = async (_req, res, _parsed) => {
  const result = service.getServiceConfig();
  sendSuccess(res, result);
};

const addMockOrderHandler: RequestHandler = async (_req, res, parsed) => {
  const body = parsed.body as Record<string, unknown>;
  
  const validation = validateBody(body, ['order']);
  if (!validation.valid) {
    sendError(res, 'INVALID_PARAMETER', 'Missing order data');
    return;
  }

  const order = body.order as Record<string, unknown>;
  const result = service.addMockOrder({
    orderId: order.orderId as string,
    userId: order.userId as string,
    totalAmount: order.totalAmount as number,
    shippingFee: order.shippingFee as number,
    products: order.products as unknown as OrderProduct[],
    orderTime: order.orderTime as string | undefined
  });

  sendSuccess(res, result);
};

const getOrdersHandler: RequestHandler = async (_req, res, _parsed) => {
  const result = service.getAllOrders();
  sendSuccess(res, result);
};

const routes: Map<string, Map<string, RequestHandler>> = new Map();

function addRoute(method: string, path: string, handler: RequestHandler): void {
  if (!routes.has(method)) {
    routes.set(method, new Map());
  }
  routes.get(method)!.set(path, handler);
}

addRoute('POST', '/api/refunds', createRefundHandler);
addRoute('GET', '/api/refunds', queryRefundsHandler);
addRoute('GET', '/api/refunds/:id', getRefundHandler);
addRoute('POST', '/api/refunds/logistics', submitLogisticsHandler);
addRoute('POST', '/api/refunds/warehouse-confirm', warehouseConfirmHandler);
addRoute('PUT', '/api/config', updateConfigHandler);
addRoute('GET', '/api/config', getConfigHandler);
addRoute('POST', '/api/orders', addMockOrderHandler);
addRoute('GET', '/api/orders', getOrdersHandler);

function matchRoute(method: string, pathname: string): { handler: RequestHandler; params: Record<string, string> } | null {
  const methodRoutes = routes.get(method);
  if (!methodRoutes) return null;

  const pathParts = pathname.split('/');

  for (const [pattern, handler] of methodRoutes) {
    const patternParts = pattern.split('/');
    if (patternParts.length !== pathParts.length) continue;

    const params: Record<string, string> = {};
    let match = true;

    for (let i = 0; i < patternParts.length; i++) {
      if (patternParts[i].startsWith(':')) {
        params[patternParts[i].slice(1)] = pathParts[i];
      } else if (patternParts[i] !== pathParts[i]) {
        match = false;
        break;
      }
    }

    if (match) {
      return { handler, params };
    }
  }

  return null;
}

const notFoundHandler: RequestHandler = async (_req, res, _parsed) => {
  sendError(res, 'NOT_FOUND', 'Route not found', 404);
};

export function createServer(): http.Server {
  return http.createServer(async (req, res) => {
    try {
      const parsed = await parseRequest(req);
      const match = matchRoute(parsed.method, parsed.pathname);

      if (match) {
        await match.handler(req, res, parsed);
      } else {
        await notFoundHandler(req, res, parsed);
      }
    } catch (error) {
      const message = error instanceof Error ? error.message : 'Internal server error';
      console.error('Server error:', error);
      sendError(res, 'INTERNAL_ERROR', message, 500);
    }
  });
}
