import * as http from 'http';
import * as url from 'url';
import { ApiResponse, createErrorResponse, ErrorCode } from '@budget-planner/shared';
import { handleGetDepartments, handlePostDepartments } from './handlers/departmentHandlers';
import { 
  handleGetBudgets, 
  handlePostBudgets, 
  handlePostBudgetsAdjust, 
  handlePostBudgetsDraft, 
  handlePostBudgetsConfirm 
} from './handlers/budgetHandlers';
import { handlePostReimbursements } from './handlers/reimbursementHandlers';
import { handlePostCarryovers } from './handlers/carryoverHandlers';
import { handleGetStatisticsUsage, handleGetStatisticsTrend } from './handlers/statisticsHandlers';

export interface RouteResult {
  statusCode: number;
  body: ApiResponse;
}

export interface ParsedRequest {
  method: string;
  pathname: string;
  query: Record<string, string | undefined>;
  body: Record<string, unknown>;
}

async function parseBody(req: http.IncomingMessage): Promise<Record<string, unknown>> {
  return new Promise((resolve, reject) => {
    if (req.method === 'GET' || req.method === 'DELETE') {
      resolve({});
      return;
    }

    let data = '';
    
    req.on('data', (chunk: Buffer) => {
      data += chunk;
    });

    req.on('end', () => {
      try {
        if (data.trim() === '') {
          resolve({});
          return;
        }
        const parsed = JSON.parse(data);
        resolve(parsed);
      } catch (error) {
        reject(error);
      }
    });

    req.on('error', (error: Error) => {
      reject(error);
    });
  });
}

function parseRequest(req: http.IncomingMessage, body: Record<string, unknown>): ParsedRequest {
  const method = req.method || 'GET';
  const parsedUrl = url.parse(req.url || '', true);
  const pathname = parsedUrl.pathname || '/';
  
  return {
    method,
    pathname,
    query: parsedUrl.query as Record<string, string | undefined>,
    body,
  };
}

export async function handleRequest(req: http.IncomingMessage): Promise<RouteResult> {
  let parsedRequest: ParsedRequest;
  
  try {
    const body = await parseBody(req);
    parsedRequest = parseRequest(req, body);
  } catch (error) {
    return {
      statusCode: 400,
      body: createErrorResponse(
        ErrorCode.INVALID_REQUEST,
        'Invalid request body',
        { error: String(error) }
      ),
    };
  }

  const { method, pathname } = parsedRequest;

  if (method === 'GET' && pathname === '/departments') {
    return handleGetDepartments(parsedRequest);
  }

  if (method === 'POST' && pathname === '/departments') {
    return handlePostDepartments(parsedRequest);
  }

  if (method === 'GET' && pathname === '/budgets') {
    return handleGetBudgets(parsedRequest);
  }

  if (method === 'POST' && pathname === '/budgets') {
    return handlePostBudgets(parsedRequest);
  }

  if (method === 'POST' && pathname === '/budgets/adjust') {
    return handlePostBudgetsAdjust(parsedRequest);
  }

  if (method === 'POST' && pathname === '/budgets/draft') {
    return handlePostBudgetsDraft(parsedRequest);
  }

  if (method === 'POST' && pathname === '/budgets/confirm') {
    return handlePostBudgetsConfirm(parsedRequest);
  }

  if (method === 'POST' && pathname === '/reimbursements') {
    return handlePostReimbursements(parsedRequest);
  }

  if (method === 'POST' && pathname === '/carryovers') {
    return handlePostCarryovers(parsedRequest);
  }

  if (method === 'GET' && pathname === '/statistics/usage') {
    return handleGetStatisticsUsage(parsedRequest);
  }

  if (method === 'GET' && pathname === '/statistics/trend') {
    return handleGetStatisticsTrend(parsedRequest);
  }

  return {
    statusCode: 404,
    body: createErrorResponse(
      ErrorCode.INVALID_REQUEST,
      `Route not found: ${method} ${pathname}`
    ),
  };
}
