import type { IncomingMessage, ServerResponse } from 'http';
import {
  handleCreateExpense,
  handleUpdateExpense,
  handleSubmitExpense,
  handleApproveExpense,
  handleRejectExpense,
  handleBatchApprove,
  handlePayExpense,
  handleResubmitExpense,
  handleDeleteExpense,
  handleGetExpense,
  handleListExpenses,
  handleGetOperationLogs,
  handleGetAuditLogs,
  handleGetUser,
  handleNotFound,
  handleMethodNotAllowed,
} from './http';

interface RouteHandler {
  method: string;
  path: string;
  handler: (req: IncomingMessage, res: ServerResponse) => Promise<void>;
}

const routes: RouteHandler[] = [
  { method: 'POST', path: '/api/expenses', handler: handleCreateExpense },
  { method: 'PUT', path: '/api/expenses', handler: handleUpdateExpense },
  { method: 'POST', path: '/api/expenses/submit', handler: handleSubmitExpense },
  { method: 'POST', path: '/api/expenses/approve', handler: handleApproveExpense },
  { method: 'POST', path: '/api/expenses/reject', handler: handleRejectExpense },
  { method: 'POST', path: '/api/expenses/batch-approve', handler: handleBatchApprove },
  { method: 'POST', path: '/api/expenses/pay', handler: handlePayExpense },
  { method: 'POST', path: '/api/expenses/resubmit', handler: handleResubmitExpense },
  { method: 'DELETE', path: '/api/expenses', handler: handleDeleteExpense },
  { method: 'GET', path: '/api/expenses/get', handler: handleGetExpense },
  { method: 'GET', path: '/api/expenses/list', handler: handleListExpenses },
  { method: 'GET', path: '/api/logs/operations', handler: handleGetOperationLogs },
  { method: 'GET', path: '/api/logs/audit', handler: handleGetAuditLogs },
  { method: 'GET', path: '/api/users/get', handler: handleGetUser },
];

function getPathWithoutQuery(url: string): string {
  const queryIndex = url.indexOf('?');
  return queryIndex >= 0 ? url.substring(0, queryIndex) : url;
}

export function routeRequest(req: IncomingMessage, res: ServerResponse): void {
  const method = req.method || 'GET';
  const path = getPathWithoutQuery(req.url || '/');

  const matchingRoute = routes.find(r => r.method === method && r.path === path);

  if (matchingRoute) {
    matchingRoute.handler(req, res).catch((err) => {
      console.error('Handler error:', err);
      res.statusCode = 500;
      res.setHeader('Content-Type', 'application/json; charset=utf-8');
      res.end(JSON.stringify({
        success: false,
        error: { code: 5001, message: '内部服务器错误' },
      }));
    });
  } else {
    const methodExists = routes.some(r => r.path === path);
    if (methodExists) {
      handleMethodNotAllowed(req, res);
    } else {
      handleNotFound(req, res);
    }
  }
}
