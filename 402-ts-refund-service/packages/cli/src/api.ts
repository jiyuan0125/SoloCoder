import http from 'http';
import { ApiResponse } from '@refund/shared';

const SERVER_HOST = 'localhost';
const SERVER_PORT = 3000;

interface RequestOptions {
  method: string;
  path: string;
  body?: unknown;
}

export async function request<T>(options: RequestOptions): Promise<ApiResponse<T>> {
  return new Promise((resolve, reject) => {
    const bodyString = options.body ? JSON.stringify(options.body) : undefined;
    
    const reqOptions: http.RequestOptions = {
      host: SERVER_HOST,
      port: SERVER_PORT,
      path: options.path,
      method: options.method,
      headers: {
        'Content-Type': 'application/json',
        ...(bodyString ? { 'Content-Length': Buffer.byteLength(bodyString) } : {})
      }
    };

    const req = http.request(reqOptions, (res) => {
      const buffers: Buffer[] = [];
      
      res.on('data', (chunk) => {
        buffers.push(chunk as Buffer);
      });
      
      res.on('end', () => {
        try {
          const rawData = Buffer.concat(buffers).toString();
          if (!rawData) {
            reject(new Error('Empty response from server'));
            return;
          }
          const data = JSON.parse(rawData) as ApiResponse<T>;
          resolve(data);
        } catch (error) {
          reject(new Error(`Failed to parse response: ${error instanceof Error ? error.message : 'Unknown error'}`));
        }
      });
    });

    req.on('error', (error) => {
      reject(new Error(`Connection error: ${error.message}`));
    });

    if (bodyString) {
      req.write(bodyString);
    }
    
    req.end();
  });
}
