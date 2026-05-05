import http from 'http';
import type { IncomingMessage } from 'http';
import type { ApiResponse, ApiErrorResponse } from '@billing/shared';

const DEFAULT_HOST = 'localhost';
const DEFAULT_PORT = 3000;

interface RequestOptions {
  method: 'GET' | 'POST' | 'PUT' | 'DELETE';
  path: string;
  body?: object;
  host?: string;
  port?: number;
}

interface ApiClientConfig {
  host: string;
  port: number;
}

interface ApiClientConfigOptional {
  host?: string | undefined;
  port?: number | undefined;
}

let config: ApiClientConfig = {
  host: DEFAULT_HOST,
  port: DEFAULT_PORT,
};

export function configureClient(newConfig: ApiClientConfigOptional): void {
  const merged: Partial<ApiClientConfig> = {};
  if (newConfig.host !== undefined) {
    merged.host = newConfig.host;
  }
  if (newConfig.port !== undefined) {
    merged.port = newConfig.port;
  }
  config = { ...config, ...merged };
}

async function makeRequest<T>(options: RequestOptions): Promise<T> {
  return new Promise((resolve, reject) => {
    const bodyString = options.body ? JSON.stringify(options.body) : undefined;

    const reqOptions = {
      hostname: options.host || config.host,
      port: options.port || config.port,
      path: options.path,
      method: options.method,
      headers: bodyString
        ? {
            'Content-Type': 'application/json',
            'Content-Length': Buffer.byteLength(bodyString),
          }
        : {},
    };

    const req = http.request(reqOptions, (res: IncomingMessage) => {
      const chunks: Buffer[] = [];

      res.on('data', (chunk: Buffer) => {
        chunks.push(chunk);
      });

      res.on('end', () => {
        let responseBody = '';
        try {
          responseBody = Buffer.concat(chunks).toString();
          const response = JSON.parse(responseBody) as ApiResponse<T>;

          if (response.success) {
            resolve(response.data);
          } else {
            const errorResponse = response as ApiErrorResponse;
            const error = new Error(errorResponse.error.message);
            (error as unknown as Record<string, unknown>).code = errorResponse.error.code;
            if (errorResponse.error.details !== undefined) {
              (error as unknown as Record<string, unknown>).details = errorResponse.error.details;
            }
            reject(error);
          }
        } catch (parseError) {
          reject(
            new Error(`解析响应失败: ${(parseError as Error).message}\n响应内容: ${responseBody}`)
          );
        }
      });

      res.on('error', (err) => {
        reject(new Error(`响应错误: ${err.message}`));
      });
    });

    req.on('error', (err) => {
      reject(
        new Error(
          `连接失败: ${err.message}\n请确保服务端已启动 (${config.host}:${config.port})`
        )
      );
    });

    if (bodyString) {
      req.write(bodyString);
    }

    req.end();
  });
}

export const apiClient = {
  get<T>(path: string, options?: Omit<RequestOptions, 'method' | 'path'>): Promise<T> {
    return makeRequest<T>({ ...options, method: 'GET', path });
  },

  post<T>(path: string, body?: object, options?: Omit<RequestOptions, 'method' | 'path' | 'body'>): Promise<T> {
    if (body !== undefined) {
      return makeRequest<T>({ ...options, method: 'POST', path, body });
    }
    return makeRequest<T>({ ...options, method: 'POST', path });
  },
};
