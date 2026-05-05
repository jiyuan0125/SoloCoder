import http from 'http';
import https from 'https';
import { URL } from 'url';
import { ApiResponse } from '@salary/shared';

export async function httpGet<T>(url: string): Promise<ApiResponse<T>> {
  const urlObj = new URL(url);
  const isHttps = urlObj.protocol === 'https:';
  const client = isHttps ? https : http;
  
  return new Promise((resolve, reject) => {
    const req = client.get(
      {
        hostname: urlObj.hostname,
        port: urlObj.port || (isHttps ? 443 : 80),
        path: urlObj.pathname + urlObj.search,
        headers: {
          'Content-Type': 'application/json',
        },
      },
      (res) => {
        let body = '';
        res.on('data', (chunk: Buffer) => {
          body += chunk.toString();
        });
        res.on('end', () => {
          try {
            const response = JSON.parse(body) as ApiResponse<T>;
            resolve(response);
          } catch {
            reject(new Error('Invalid JSON response from server'));
          }
        });
      }
    );
    
    req.on('error', (err) => {
      reject(err);
    });
    
    req.end();
  });
}

export async function httpPost<T>(url: string, data: unknown): Promise<ApiResponse<T>> {
  const urlObj = new URL(url);
  const isHttps = urlObj.protocol === 'https:';
  const client = isHttps ? https : http;
  const postData = JSON.stringify(data);
  
  return new Promise((resolve, reject) => {
    const req = client.request(
      {
        hostname: urlObj.hostname,
        port: urlObj.port || (isHttps ? 443 : 80),
        path: urlObj.pathname,
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Content-Length': Buffer.byteLength(postData),
        },
      },
      (res) => {
        let body = '';
        res.on('data', (chunk: Buffer) => {
          body += chunk.toString();
        });
        res.on('end', () => {
          try {
            const response = JSON.parse(body) as ApiResponse<T>;
            resolve(response);
          } catch {
            reject(new Error('Invalid JSON response from server'));
          }
        });
      }
    );
    
    req.on('error', (err) => {
      reject(err);
    });
    
    req.write(postData);
    req.end();
  });
}
