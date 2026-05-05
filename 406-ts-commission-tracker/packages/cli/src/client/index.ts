import * as http from 'http';
import { ApiResponse } from '@commission-tracker/shared';

const DEFAULT_HOST = process.env.COMMISSION_API_HOST || 'localhost';
const DEFAULT_PORT = parseInt(process.env.COMMISSION_API_PORT || '3000', 10);

export interface ApiClientConfig {
  host: string;
  port: number;
}

export class ApiClient {
  private config: ApiClientConfig;

  constructor(config?: Partial<ApiClientConfig>) {
    this.config = {
      host: config?.host || DEFAULT_HOST,
      port: config?.port || DEFAULT_PORT,
    };
  }

  private async request<T>(
    method: string,
    path: string,
    body?: Record<string, unknown>
  ): Promise<ApiResponse<T>> {
    return new Promise((resolve, reject) => {
      const postData = body ? JSON.stringify(body) : undefined;
      const headers: Record<string, string> = {
        'Content-Type': 'application/json',
      };
      
      if (postData) {
        headers['Content-Length'] = String(Buffer.byteLength(postData));
      }

      const options: http.RequestOptions = {
        hostname: this.config.host,
        port: this.config.port,
        path,
        method,
        headers,
      };

      const req = http.request(options, (res) => {
        let data = '';
        res.on('data', (chunk: Buffer) => {
          data += chunk.toString();
        });
        res.on('end', () => {
          try {
            const result = JSON.parse(data) as ApiResponse<T>;
            resolve(result);
          } catch {
            reject(new Error(`Failed to parse response: ${data}`));
          }
        });
      });

      req.on('error', reject);

      if (postData) {
        req.write(postData);
      }
      req.end();
    });
  }

  async get<T>(path: string): Promise<ApiResponse<T>> {
    return this.request<T>('GET', path);
  }

  async post<T>(path: string, body?: Record<string, unknown>): Promise<ApiResponse<T>> {
    return this.request<T>('POST', path, body);
  }
}

export const apiClient = new ApiClient();
