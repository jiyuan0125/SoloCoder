import * as http from "http";
import { ApiResponse } from "@payment-gateway/shared";

const SERVER_HOST = "localhost";
const SERVER_PORT = 8080;

interface HttpRequestOptions {
  method: string;
  path: string;
  headers?: Record<string, string>;
  body?: string;
}

async function makeRequest<T>(options: HttpRequestOptions): Promise<ApiResponse<T>> {
  return new Promise((resolve, reject) => {
    const headers: Record<string, string> = {
      ...options.headers,
    };

    if (options.body) {
      headers["Content-Type"] = "application/json";
      headers["Content-Length"] = Buffer.byteLength(options.body, "utf8").toString();
    }

    const req = http.request(
      {
        hostname: SERVER_HOST,
        port: SERVER_PORT,
        path: options.path,
        method: options.method,
        headers: headers,
      },
      (res) => {
        let responseBody = "";
        res.on("data", (chunk) => {
          responseBody += chunk;
        });
        res.on("end", () => {
          try {
            const parsed = JSON.parse(responseBody) as ApiResponse<T>;
            resolve(parsed);
          } catch {
            reject(new Error(`Invalid JSON response: ${responseBody}`));
          }
        });
      }
    );

    req.on("error", (err) => {
      reject(err);
    });

    if (options.body) {
      req.write(options.body);
    }

    req.end();
  });
}

export async function post<T>(path: string, body?: unknown): Promise<ApiResponse<T>> {
  const bodyStr = body !== undefined ? JSON.stringify(body) : undefined;
  return makeRequest<T>({
    method: "POST",
    path: path,
    body: bodyStr,
  });
}

export async function get<T>(path: string): Promise<ApiResponse<T>> {
  return makeRequest<T>({
    method: "GET",
    path: path,
  });
}
