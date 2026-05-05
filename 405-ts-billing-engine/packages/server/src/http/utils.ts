import type { IncomingMessage, ServerResponse } from 'http';
import { BillingError, ErrorCodes } from '@billing/shared';
import type { ApiResponse, ApiErrorResponse } from '@billing/shared';

export async function parseRequestBody<T>(req: IncomingMessage): Promise<T> {
  return new Promise((resolve, reject) => {
    const chunks: Buffer[] = [];

    req.on('data', (chunk: Buffer) => {
      chunks.push(chunk);
    });

    req.on('end', () => {
      try {
        const body = Buffer.concat(chunks).toString();
        if (!body) {
          resolve({} as T);
          return;
        }
        resolve(JSON.parse(body) as T);
      } catch (error) {
        reject(new BillingError(ErrorCodes.BAD_REQUEST, '无效的JSON请求体'));
      }
    });

    req.on('error', (error) => {
      reject(error);
    });
  });
}

export function sendResponse<T>(
  res: ServerResponse,
  data: T,
  statusCode: number = 200
): void {
  const response: ApiResponse<T> = {
    success: true,
    data,
  };

  res.statusCode = statusCode;
  res.setHeader('Content-Type', 'application/json');
  res.end(JSON.stringify(response));
}

export function sendError(
  res: ServerResponse,
  error: unknown,
  statusCode: number = 500
): void {
  let errorResponse: ApiErrorResponse;

  if (error instanceof BillingError) {
    const errorObj: { code: string; message: string; details?: Record<string, unknown> } = {
      code: error.code,
      message: error.message,
    };
    if (error.details !== undefined) {
      errorObj.details = error.details;
    }
    errorResponse = {
      success: false,
      error: errorObj,
    };
    if (
      error.code === ErrorCodes.CUSTOMER_NOT_FOUND ||
      error.code === ErrorCodes.SUBSCRIPTION_NOT_FOUND ||
      error.code === ErrorCodes.BILL_NOT_FOUND
    ) {
      statusCode = 404;
    } else if (
      error.code === ErrorCodes.CUSTOMER_FROZEN ||
      error.code === ErrorCodes.BAD_REQUEST ||
      error.code === ErrorCodes.INVALID_PLAN_TYPE ||
      error.code === ErrorCodes.INVALID_USAGE_AMOUNT ||
      error.code === ErrorCodes.INVALID_PAYMENT_AMOUNT ||
      error.code === ErrorCodes.CUSTOMER_ALREADY_EXISTS ||
      error.code === ErrorCodes.SUBSCRIPTION_ALREADY_ACTIVE ||
      error.code === ErrorCodes.USAGE_OVER_QUOTA_LIMITED ||
      error.code === ErrorCodes.BILL_ALREADY_PAID ||
      error.code === ErrorCodes.BILL_REVIEW_WINDOW_EXPIRED ||
      error.code === ErrorCodes.BILL_ALREADY_UNDER_REVIEW ||
      error.code === ErrorCodes.REFUND_NOT_ELIGIBLE
    ) {
      statusCode = 400;
    }
  } else if (error instanceof Error) {
    errorResponse = {
      success: false,
      error: {
        code: ErrorCodes.INTERNAL_ERROR,
        message: error.message,
      },
    };
  } else {
    errorResponse = {
      success: false,
      error: {
        code: ErrorCodes.INTERNAL_ERROR,
        message: '发生未知错误',
      },
    };
  }

  res.statusCode = statusCode;
  res.setHeader('Content-Type', 'application/json');
  res.end(JSON.stringify(errorResponse));
}

export function parseQueryParams(req: IncomingMessage): Record<string, string> {
  const url = req.url || '';
  const queryIndex = url.indexOf('?');
  if (queryIndex === -1) {
    return {};
  }

  const queryString = url.slice(queryIndex + 1);
  const params: Record<string, string> = {};

  for (const pair of queryString.split('&')) {
    const [key, value] = pair.split('=', 2);
    if (key && value !== undefined) {
      params[decodeURIComponent(key)] = decodeURIComponent(value);
    }
  }

  return params;
}

export function getPathSegments(req: IncomingMessage): string[] {
  const url = req.url || '';
  const path = url.split('?')[0] || '';
  return path.split('/').filter(Boolean);
}
