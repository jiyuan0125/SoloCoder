import http from 'http';
import { IncomingMessage, ServerResponse } from 'http';
import { handleHealth, handleCalculate, handleSave, handleQuery, handleComparison } from './handlers.js';
import { API_ENDPOINTS } from '@salary/shared';
import { createApiError, ErrorCode } from '@salary/shared';
import { ApiResponse } from '@salary/shared';

type RequestHandler = (req: IncomingMessage, res: ServerResponse) => Promise<void>;

function parseBody(req: IncomingMessage): Promise<unknown> {
  return new Promise((resolve, reject) => {
    let body = '';
    req.on('data', (chunk: Buffer) => {
      body += chunk.toString();
    });
    req.on('end', () => {
      try {
        resolve(body ? JSON.parse(body) : {});
      } catch {
        reject(new Error('Invalid JSON'));
      }
    });
    req.on('error', reject);
  });
}

function sendJson(res: ServerResponse, statusCode: number, data: ApiResponse): void {
  res.setHeader('Content-Type', 'application/json; charset=utf-8');
  res.statusCode = statusCode;
  res.end(JSON.stringify(data));
}

function createErrorResponse(code: ErrorCode, message: string): ApiResponse {
  return {
    success: false,
    error: createApiError(code, message),
  };
}

function createSuccessResponse<T>(data: T): ApiResponse<T> {
  return {
    success: true,
    data,
  };
}

async function handleOptions(req: IncomingMessage, res: ServerResponse): Promise<void> {
  res.setHeader('Access-Control-Allow-Origin', '*');
  res.setHeader('Access-Control-Allow-Methods', 'GET, POST, OPTIONS');
  res.setHeader('Access-Control-Allow-Headers', 'Content-Type');
  res.statusCode = 204;
  res.end();
}

function createRouter(): Map<string, Map<string, RequestHandler>> {
  const router = new Map<string, Map<string, RequestHandler>>();
  
  const getHandlers = new Map<string, RequestHandler>();
  getHandlers.set(API_ENDPOINTS.HEALTH, handleHealth);
  getHandlers.set(API_ENDPOINTS.QUERY, handleQuery);
  getHandlers.set(API_ENDPOINTS.COMPARISON, handleComparison);
  router.set('GET', getHandlers);
  
  const postHandlers = new Map<string, RequestHandler>();
  postHandlers.set(API_ENDPOINTS.CALCULATE, handleCalculate);
  postHandlers.set(API_ENDPOINTS.SAVE, handleSave);
  router.set('POST', postHandlers);
  
  const optionsHandlers = new Map<string, RequestHandler>();
  optionsHandlers.set(API_ENDPOINTS.CALCULATE, handleOptions);
  optionsHandlers.set(API_ENDPOINTS.SAVE, handleOptions);
  optionsHandlers.set(API_ENDPOINTS.QUERY, handleOptions);
  optionsHandlers.set(API_ENDPOINTS.COMPARISON, handleOptions);
  router.set('OPTIONS', optionsHandlers);
  
  return router;
}

const router = createRouter();

export function createServer(): http.Server {
  return http.createServer(async (req: IncomingMessage, res: ServerResponse) => {
    res.setHeader('Access-Control-Allow-Origin', '*');
    
    const method = req.method || 'GET';
    const url = req.url || '/';
    const pathname = url.split('?')[0];
    
    const methodHandlers = router.get(method);
    
    if (!methodHandlers) {
      sendJson(res, 405, createErrorResponse(ErrorCode.INVALID_INPUT, `Method ${method} not allowed`));
      return;
    }
    
    const handler = methodHandlers.get(pathname);
    
    if (!handler) {
      sendJson(res, 404, createErrorResponse(ErrorCode.RECORD_NOT_FOUND, `Endpoint ${pathname} not found`));
      return;
    }
    
    try {
      await handler(req, res);
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Unknown error';
      sendJson(res, 500, createErrorResponse(ErrorCode.INTERNAL_ERROR, message));
    }
  });
}

export { parseBody, sendJson, createErrorResponse, createSuccessResponse };
