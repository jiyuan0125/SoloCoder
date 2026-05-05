import http from "node:http";
import https from "node:https";
import { URL } from "node:url";
import { ApiResponse, CreateInvoiceRequest, InvoiceFilter, UpdateInvoiceRequest } from "@tax-invoice/shared";
import { config } from "./config.js";

interface RequestOptions {
  method: "GET" | "POST" | "PUT" | "DELETE";
  path: string;
  body?: unknown;
}

function request<T>(options: RequestOptions): Promise<ApiResponse<T>> {
  return new Promise((resolve, reject) => {
    const url = new URL(options.path, config.baseUrl);
    const httpModule = url.protocol === "https:" ? https : http;

    const requestBody = options.body ? JSON.stringify(options.body) : undefined;

    const req = httpModule.request(
      {
        hostname: url.hostname,
        port: url.port || (url.protocol === "https:" ? 443 : 80),
        path: url.pathname + url.search,
        method: options.method,
        headers: {
          "Content-Type": "application/json",
          ...(requestBody ? { "Content-Length": Buffer.byteLength(requestBody) } : {}),
        },
        timeout: config.timeout,
      },
      (res) => {
        let responseBody = "";

        res.on("data", (chunk) => {
          responseBody += chunk;
        });

        res.on("end", () => {
          try {
            const response: ApiResponse<T> = JSON.parse(responseBody);
            resolve(response);
          } catch (error) {
            reject(new Error(`Failed to parse response: ${responseBody}`));
          }
        });
      }
    );

    req.on("error", (error) => {
      reject(error);
    });

    req.on("timeout", () => {
      req.destroy();
      reject(new Error("Request timeout"));
    });

    if (requestBody) {
      req.write(requestBody);
    }

    req.end();
  });
}

export async function createInvoice(data: CreateInvoiceRequest) {
  return request({
    method: "POST",
    path: "/api/invoices",
    body: data,
  });
}

export async function getInvoice(id: string) {
  return request({
    method: "GET",
    path: `/api/invoices/${id}`,
  });
}

export async function queryInvoices(filter: InvoiceFilter) {
  const params = new URLSearchParams();
  if (filter.taxAmountMin !== undefined) params.append("taxAmountMin", filter.taxAmountMin.toString());
  if (filter.taxAmountMax !== undefined) params.append("taxAmountMax", filter.taxAmountMax.toString());
  if (filter.invoiceType) params.append("invoiceType", filter.invoiceType);
  if (filter.status) params.append("status", filter.status);
  if (filter.startDate) params.append("startDate", filter.startDate);
  if (filter.endDate) params.append("endDate", filter.endDate);

  const queryString = params.toString();
  const path = queryString ? `/api/invoices?${queryString}` : "/api/invoices";

  return request({
    method: "GET",
    path,
  });
}

export async function updateInvoice(data: UpdateInvoiceRequest) {
  return request({
    method: "PUT",
    path: `/api/invoices/${data.id}`,
    body: data,
  });
}

export async function voidInvoice(id: string) {
  return request({
    method: "POST",
    path: `/api/invoices/${id}/void`,
  });
}

export async function redInvoice(originalId: string, data: {
  invoiceNumber: string;
  buyer: { name: string; taxNumber: string };
  seller: { name: string; taxNumber: string };
  issuedAt: string;
}) {
  return request({
    method: "POST",
    path: `/api/invoices/${originalId}/red`,
    body: data,
  });
}

export async function batchImport(invoices: CreateInvoiceRequest[]) {
  return request({
    method: "POST",
    path: "/api/invoices/batch-import",
    body: invoices,
  });
}

export async function getMonthlySummary(year?: number, month?: number) {
  const params = new URLSearchParams();
  if (year !== undefined) params.append("year", year.toString());
  if (month !== undefined) params.append("month", month.toString());

  const queryString = params.toString();
  const path = queryString ? `/api/invoices/monthly-summary?${queryString}` : "/api/invoices/monthly-summary";

  return request({
    method: "GET",
    path,
  });
}
