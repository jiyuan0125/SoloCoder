import { IncomingMessage, ServerResponse } from 'http';
import { ApiResponse } from '@commission-tracker/shared';

export function parseRequestBody(req: IncomingMessage): Promise<string> {
  return new Promise((resolve, reject) => {
    let body = '';
    req.on('data', (chunk: Buffer) => {
      body += chunk.toString();
    });
    req.on('end', () => {
      resolve(body);
    });
    req.on('error', reject);
  });
}

export function sendJsonResponse(res: ServerResponse, statusCode: number, data: ApiResponse): void {
  res.statusCode = statusCode;
  res.setHeader('Content-Type', 'application/json; charset=utf-8');
  res.end(JSON.stringify(data));
}

export function parseQueryParams(url: string): Record<string, string> {
  const params: Record<string, string> = {};
  const queryIndex = url.indexOf('?');
  if (queryIndex === -1) return params;

  const queryString = url.substring(queryIndex + 1);
  const pairs = queryString.split('&');

  for (const pair of pairs) {
    const [key, value] = pair.split('=');
    if (key) {
      params[decodeURIComponent(key)] = value ? decodeURIComponent(value) : '';
    }
  }

  return params;
}

export function extractPathSegment(pathname: string, index: number): string | undefined {
  const parts = pathname.split('/');
  return parts[index];
}

export function extractSalespersonId(pathname: string): string | undefined {
  return extractPathSegment(pathname, 3);
}

export function extractOrderId(pathname: string): string | undefined {
  return extractPathSegment(pathname, 3);
}

export function extractSettlementId(pathname: string): string | undefined {
  return extractPathSegment(pathname, 3);
}
