import http from "node:http";
import {
  ApiResponse,
  CreateInvoiceRequest,
  InvoiceFilter,
  UpdateInvoiceRequest,
  BusinessError,
  ErrorCode,
} from "@tax-invoice/shared";
import { invoiceService } from "./service.js";
import { asyncWrapper } from "./utils.js";

type HttpMethod = "GET" | "POST" | "PUT" | "DELETE" | "PATCH";

interface RouteHandler {
  (req: http.IncomingMessage, res: http.ServerResponse): Promise<void>;
}

interface Route {
  method: HttpMethod;
  path: string | RegExp;
  handler: RouteHandler;
}

function parseJsonBody(req: http.IncomingMessage): Promise<unknown> {
  return new Promise((resolve, reject) => {
    let body = "";

    req.on("data", (chunk: Buffer) => {
      body += chunk.toString();
    });

    req.on("end", () => {
      try {
        if (body) {
          resolve(JSON.parse(body));
        } else {
          resolve({});
        }
      } catch (error) {
        reject(error);
      }
    });

    req.on("error", (error) => {
      reject(error);
    });
  });
}

function sendJson(res: http.ServerResponse, statusCode: number, data: ApiResponse): void {
  res.setHeader("Content-Type", "application/json; charset=utf-8");
  res.statusCode = statusCode;
  res.end(JSON.stringify(data));
}

function sendSuccess<T>(res: http.ServerResponse, data: T, statusCode = 200): void {
  sendJson(res, statusCode, {
    success: true,
    data,
  });
}

function sendError(res: http.ServerResponse, statusCode: number, code: ErrorCode, message: string, details?: unknown): void {
  sendJson(res, statusCode, {
    success: false,
    error: {
      code,
      message,
      details,
    },
  });
}

function handleError(res: http.ServerResponse, error: unknown): void {
  console.error("Request error:", error);

  if (error instanceof BusinessError) {
    let statusCode = 400;
    switch (error.code) {
      case ErrorCode.INVOICE_NOT_FOUND:
        statusCode = 404;
        break;
      case ErrorCode.INTERNAL_ERROR:
        statusCode = 500;
        break;
    }
    sendError(res, statusCode, error.code, error.message, error.details);
    return;
  }

  if (error instanceof SyntaxError) {
    sendError(res, 400, ErrorCode.MISSING_REQUIRED_FIELD, "JSON解析错误");
    return;
  }

  sendError(res, 500, ErrorCode.INTERNAL_ERROR, "内部服务器错误");
}

async function createInvoiceHandler(req: http.IncomingMessage, res: http.ServerResponse): Promise<void> {
  const body = (await parseJsonBody(req)) as CreateInvoiceRequest;
  const invoice = invoiceService.createInvoice(body);
  sendSuccess(res, invoice, 201);
}

async function getInvoiceHandler(req: http.IncomingMessage, res: http.ServerResponse): Promise<void> {
  const url = new URL(req.url || "", `http://${req.headers.host}`);
  const pathParts = url.pathname.split("/").filter(Boolean);
  const id = pathParts[2];

  if (!id) {
    sendError(res, 400, ErrorCode.MISSING_REQUIRED_FIELD, "缺少发票ID");
    return;
  }

  const invoice = invoiceService.getInvoiceById(id);
  sendSuccess(res, invoice);
}

async function queryInvoicesHandler(req: http.IncomingMessage, res: http.ServerResponse): Promise<void> {
  const url = new URL(req.url || "", `http://${req.headers.host}`);
  const filter: InvoiceFilter = {};

  const taxAmountMin = url.searchParams.get("taxAmountMin");
  const taxAmountMax = url.searchParams.get("taxAmountMax");
  const invoiceType = url.searchParams.get("invoiceType");
  const status = url.searchParams.get("status");
  const startDate = url.searchParams.get("startDate");
  const endDate = url.searchParams.get("endDate");

  if (taxAmountMin) {
    filter.taxAmountMin = parseInt(taxAmountMin, 10);
  }
  if (taxAmountMax) {
    filter.taxAmountMax = parseInt(taxAmountMax, 10);
  }
  if (invoiceType) {
    filter.invoiceType = invoiceType as InvoiceFilter["invoiceType"];
  }
  if (status) {
    filter.status = status as InvoiceFilter["status"];
  }
  if (startDate) {
    filter.startDate = startDate;
  }
  if (endDate) {
    filter.endDate = endDate;
  }

  const invoices = invoiceService.queryInvoices(filter);
  sendSuccess(res, { items: invoices, total: invoices.length });
}

async function updateInvoiceHandler(req: http.IncomingMessage, res: http.ServerResponse): Promise<void> {
  const url = new URL(req.url || "", `http://${req.headers.host}`);
  const pathParts = url.pathname.split("/").filter(Boolean);
  const id = pathParts[2];

  if (!id) {
    sendError(res, 400, ErrorCode.MISSING_REQUIRED_FIELD, "缺少发票ID");
    return;
  }

  const body = (await parseJsonBody(req)) as UpdateInvoiceRequest;
  const invoice = invoiceService.updateInvoice({ ...body, id });
  sendSuccess(res, invoice);
}

async function voidInvoiceHandler(req: http.IncomingMessage, res: http.ServerResponse): Promise<void> {
  const url = new URL(req.url || "", `http://${req.headers.host}`);
  const pathParts = url.pathname.split("/").filter(Boolean);
  const id = pathParts[2];

  if (!id) {
    sendError(res, 400, ErrorCode.MISSING_REQUIRED_FIELD, "缺少发票ID");
    return;
  }

  const invoice = invoiceService.voidInvoice(id);
  sendSuccess(res, invoice);
}

async function redInvoiceHandler(req: http.IncomingMessage, res: http.ServerResponse): Promise<void> {
  const url = new URL(req.url || "", `http://${req.headers.host}`);
  const pathParts = url.pathname.split("/").filter(Boolean);
  const originalId = pathParts[2];

  if (!originalId) {
    sendError(res, 400, ErrorCode.MISSING_REQUIRED_FIELD, "缺少原发票ID");
    return;
  }

  const body = (await parseJsonBody(req)) as {
    invoiceNumber: string;
    buyer: { name: string; taxNumber: string };
    seller: { name: string; taxNumber: string };
    issuedAt: string;
  };

  const result = invoiceService.redInvoice(originalId, body);
  sendSuccess(res, result);
}

async function batchImportHandler(req: http.IncomingMessage, res: http.ServerResponse): Promise<void> {
  const body = (await parseJsonBody(req)) as CreateInvoiceRequest[];
  const result = invoiceService.batchImport(body);
  sendSuccess(res, result);
}

async function monthlySummaryHandler(req: http.IncomingMessage, res: http.ServerResponse): Promise<void> {
  const url = new URL(req.url || "", `http://${req.headers.host}`);
  const yearParam = url.searchParams.get("year");
  const monthParam = url.searchParams.get("month");

  const now = new Date();
  const year = yearParam ? parseInt(yearParam, 10) : now.getFullYear();
  const month = monthParam ? parseInt(monthParam, 10) : now.getMonth() + 1;

  const summary = invoiceService.getMonthlySummary(year, month);
  sendSuccess(res, summary);
}

const routes: Route[] = [
  { method: "POST", path: "/api/invoices", handler: createInvoiceHandler },
  { method: "GET", path: /^\/api\/invoices\/[^\/]+$/, handler: getInvoiceHandler },
  { method: "GET", path: "/api/invoices", handler: queryInvoicesHandler },
  { method: "PUT", path: /^\/api\/invoices\/[^\/]+$/, handler: updateInvoiceHandler },
  { method: "POST", path: /^\/api\/invoices\/[^\/]+\/void$/, handler: voidInvoiceHandler },
  { method: "POST", path: /^\/api\/invoices\/[^\/]+\/red$/, handler: redInvoiceHandler },
  { method: "POST", path: "/api/invoices/batch-import", handler: batchImportHandler },
  { method: "GET", path: "/api/invoices/monthly-summary", handler: monthlySummaryHandler },
];

function matchRoute(req: http.IncomingMessage): Route | null {
  const method = req.method as HttpMethod;
  const url = new URL(req.url || "", `http://${req.headers.host}`);
  const path = url.pathname;

  for (const route of routes) {
    if (route.method !== method) {
      continue;
    }

    if (typeof route.path === "string") {
      if (route.path === path) {
        return route;
      }
    } else if (route.path.test(path)) {
      return route;
    }
  }

  return null;
}

export async function handleRequest(req: http.IncomingMessage, res: http.ServerResponse): Promise<void> {
  res.setHeader("Access-Control-Allow-Origin", "*");
  res.setHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS");
  res.setHeader("Access-Control-Allow-Headers", "Content-Type");

  if (req.method === "OPTIONS") {
    res.statusCode = 200;
    res.end();
    return;
  }

  const route = matchRoute(req);

  if (!route) {
    sendError(res, 404, ErrorCode.INVOICE_NOT_FOUND, "路由不存在");
    return;
  }

  try {
    await asyncWrapper(
      async () => route.handler(req, res),
      (error) => {
        handleError(res, error);
        return undefined;
      }
    );
  } catch (error) {
    handleError(res, error);
  }
}
