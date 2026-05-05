import * as http from 'http';

const DEFAULT_HOST = 'localhost';
const DEFAULT_PORT = 3000;

export interface HttpClientConfig {
  host?: string;
  port?: number;
}

export class HttpClient {
  private host: string;
  private port: number;

  constructor(config?: HttpClientConfig) {
    this.host = config?.host ?? DEFAULT_HOST;
    this.port = config?.port ?? DEFAULT_PORT;
  }

  public async get<T>(path: string): Promise<T> {
    return this.request<T>('GET', path);
  }

  public async post<T>(path: string, body: unknown): Promise<T> {
    return this.request<T>('POST', path, body);
  }

  private request<T>(method: string, path: string, body?: unknown): Promise<T> {
    return new Promise((resolve, reject) => {
      const postData = body ? JSON.stringify(body) : '';

      const options: http.RequestOptions = {
        hostname: this.host,
        port: this.port,
        path,
        method,
        headers: {
          'Content-Type': 'application/json',
          'Content-Length': Buffer.byteLength(postData),
        },
      };

      const req = http.request(options, (res) => {
        let data = '';

        res.on('data', (chunk) => {
          data += chunk;
        });

        res.on('end', () => {
          try {
            const parsed = JSON.parse(data) as T;
            resolve(parsed);
          } catch (error) {
            reject(new Error(`Failed to parse response: ${data}`));
          }
        });
      });

      req.on('error', (error) => {
        reject(error);
      });

      if (postData) {
        req.write(postData);
      }
      req.end();
    });
  }
}

export const httpClient = new HttpClient();
