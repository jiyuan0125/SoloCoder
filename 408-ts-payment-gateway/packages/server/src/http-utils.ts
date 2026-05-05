import { IncomingMessage, ServerResponse } from "http";
import { ApiResponse, ErrorCode } from "@payment-gateway/shared";
import { PaymentGatewayError } from "@payment-gateway/shared";

export function parseBody(req: IncomingMessage): Promise<string> {
  return new Promise((resolve, reject) => {
    let body = "";
    req.on("data", (chunk) => {
      body += chunk.toString();
    });
    req.on("end", () => {
      resolve(body);
    });
    req.on("error", (err) => {
      reject(err);
    });
  });
}

export function parseJsonBody<T>(req: IncomingMessage): Promise<T> {
  return parseBody(req).then((body) => {
    try {
      return JSON.parse(body) as T;
    } catch {
      throw new PaymentGatewayError(ErrorCode.INVALID_PARAMS, "无效的 JSON 格式");
    }
  });
}

export function sendJsonResponse<T>(
  res: ServerResponse,
  statusCode: number,
  response: ApiResponse<T>
): void {
  const jsonStr = JSON.stringify(response);
  res.writeHead(statusCode, {
    "Content-Type": "application/json; charset=utf-8",
    "Content-Length": Buffer.byteLength(jsonStr, "utf8"),
  });
  res.end(jsonStr);
}

export function sendSuccessResponse<T>(
  res: ServerResponse,
  data: T | null = null
): void {
  const response: ApiResponse<T> = {
    code: ErrorCode.SUCCESS,
    message: "成功",
    data: data,
  };
  sendJsonResponse(res, 200, response);
}

export function sendErrorResponse(
  res: ServerResponse,
  error: PaymentGatewayError
): void {
  const response: ApiResponse<null> = {
    code: error.code,
    message: error.message,
    data: null,
  };
  const statusCode = error.code >= 20000 ? 500 : 400;
  sendJsonResponse(res, statusCode, response);
}

export function sendNotFoundResponse(res: ServerResponse): void {
  const response: ApiResponse<null> = {
    code: ErrorCode.INVALID_PARAMS,
    message: "未找到该接口",
    data: null,
  };
  sendJsonResponse(res, 404, response);
}

export function sendMethodNotAllowedResponse(res: ServerResponse, allowedMethods: string[]): void {
  const response: ApiResponse<null> = {
    code: ErrorCode.INVALID_PARAMS,
    message: `方法不允许，允许的方法: ${allowedMethods.join(", ")}`,
    data: null,
  };
  res.writeHead(405, {
    "Content-Type": "application/json; charset=utf-8",
    Allow: allowedMethods.join(", "),
  });
  res.end(JSON.stringify(response));
}

export function parseQueryString(url: string): Record<string, string> {
  const queryIndex = url.indexOf("?");
  if (queryIndex === -1) {
    return {};
  }
  const queryStr = url.substring(queryIndex + 1);
  const params: Record<string, string> = {};
  const pairs = queryStr.split("&");
  for (const pair of pairs) {
    const [key, value] = pair.split("=");
    if (key) {
      params[decodeURIComponent(key)] = value ? decodeURIComponent(value) : "";
    }
  }
  return params;
}

export function getPathname(url: string): string {
  const queryIndex = url.indexOf("?");
  if (queryIndex === -1) {
    return url;
  }
  return url.substring(0, queryIndex);
}
