import { IncomingMessage, ServerResponse } from 'http';
import {
  handleCreateSalesperson,
  handleGetSalesperson,
  handleListSalespeople,
  handleResignSalesperson,
  handleGetSalespersonBalance,
} from '../handlers/salesperson';
import {
  handleCreateOrder,
  handleGetOrder,
  handleListOrders,
  handleRefundOrder,
} from '../handlers/order';
import {
  handleGetSalespersonCommission,
} from '../handlers/commission';
import {
  handleTriggerMonthlySettlement,
  handleGetSettlements,
  handleGetSettlement,
} from '../handlers/settlement';
import {
  handleGetRankings,
} from '../handlers/ranking';
import { sendJsonResponse } from '../utils';
import { errorResponse, ErrorCodes } from '@commission-tracker/shared';

type RouteHandler = (req: IncomingMessage, res: ServerResponse) => Promise<void>;

interface Route {
  method: string;
  pattern: RegExp;
  handler: RouteHandler;
}

const routes: Route[] = [
  { method: 'POST', pattern: /^\/api\/salespeople$/, handler: handleCreateSalesperson },
  { method: 'GET', pattern: /^\/api\/salespeople$/, handler: handleListSalespeople },
  { method: 'GET', pattern: /^\/api\/salespeople\/[^\/]+$/, handler: handleGetSalesperson },
  { method: 'POST', pattern: /^\/api\/salespeople\/[^\/]+\/resign$/, handler: handleResignSalesperson },
  { method: 'GET', pattern: /^\/api\/salespeople\/[^\/]+\/commission$/, handler: handleGetSalespersonCommission },
  { method: 'GET', pattern: /^\/api\/salespeople\/[^\/]+\/balance$/, handler: handleGetSalespersonBalance },
  
  { method: 'POST', pattern: /^\/api\/orders$/, handler: handleCreateOrder },
  { method: 'GET', pattern: /^\/api\/orders$/, handler: handleListOrders },
  { method: 'GET', pattern: /^\/api\/orders\/[^\/]+$/, handler: handleGetOrder },
  { method: 'POST', pattern: /^\/api\/orders\/[^\/]+\/refund$/, handler: handleRefundOrder },
  
  { method: 'POST', pattern: /^\/api\/settlements\/trigger$/, handler: handleTriggerMonthlySettlement },
  { method: 'GET', pattern: /^\/api\/settlements$/, handler: handleGetSettlements },
  { method: 'GET', pattern: /^\/api\/settlements\/[^\/]+$/, handler: handleGetSettlement },
  
  { method: 'GET', pattern: /^\/api\/rankings$/, handler: handleGetRankings },
];

function matchRoute(method: string, url: string): RouteHandler | null {
  const path = url.split('?')[0];
  
  for (const route of routes) {
    if (route.method === method && route.pattern.test(path)) {
      return route.handler;
    }
  }
  
  return null;
}

export async function handleRequest(req: IncomingMessage, res: ServerResponse): Promise<void> {
  const method = req.method || 'GET';
  const url = req.url || '/';

  if (method === 'OPTIONS') {
    res.statusCode = 200;
    res.setHeader('Access-Control-Allow-Origin', '*');
    res.setHeader('Access-Control-Allow-Methods', 'GET, POST, OPTIONS');
    res.setHeader('Access-Control-Allow-Headers', 'Content-Type');
    res.end();
    return;
  }

  res.setHeader('Access-Control-Allow-Origin', '*');
  res.setHeader('Content-Type', 'application/json; charset=utf-8');

  const handler = matchRoute(method, url);

  if (handler) {
    try {
      await handler(req, res);
    } catch (error) {
      sendJsonResponse(
        res,
        500,
        errorResponse(ErrorCodes.INTERNAL_ERROR, '服务器内部错误')
      );
    }
  } else {
    sendJsonResponse(
      res,
      404,
      errorResponse('NOT_FOUND', '路由不存在')
    );
  }
}
