import http from 'http';
import type { IncomingMessage, ServerResponse } from 'http';
import { sendError, getPathSegments } from './utils';
import {
  handleCustomerRoutes,
  handleSubscriptionRoutes,
  handleUsageRoutes,
  handleBillingRoutes,
} from './handlers';
import { BillingError, ErrorCodes } from '@billing/shared';

const PORT = 3000;
const HOST = 'localhost';

async function requestHandler(req: IncomingMessage, res: ServerResponse): Promise<void> {
  res.setHeader('Content-Type', 'application/json');

  const segments = getPathSegments(req);

  if (segments.length === 0) {
    res.statusCode = 200;
    res.end(
      JSON.stringify({
        success: true,
        data: {
          service: 'billing-engine',
          version: '1.0.0',
          endpoints: {
            customers: '/customers',
            subscriptions: '/subscriptions',
            usage: '/usage',
            bills: '/bills',
          },
        },
      })
    );
    return;
  }

  try {
    const resource = segments[0];

    switch (resource) {
      case 'customers':
        await handleCustomerRoutes(req, res, segments.slice(1));
        break;
      case 'subscriptions':
        await handleSubscriptionRoutes(req, res, segments.slice(1));
        break;
      case 'usage':
        await handleUsageRoutes(req, res, segments.slice(1));
        break;
      case 'bills':
        await handleBillingRoutes(req, res, segments.slice(1));
        break;
      default:
        sendError(res, new BillingError(ErrorCodes.BAD_REQUEST, '无效的资源路径'), 404);
    }
  } catch (error) {
    sendError(res, error);
  }
}

function createServer(): http.Server {
  return http.createServer((req: IncomingMessage, res: ServerResponse) => {
    requestHandler(req, res).catch((error: unknown) => {
      sendError(res, error);
    });
  });
}

const server = createServer();

server.on('error', (error: Error) => {
  console.error('服务器错误:', error);
  process.exit(1);
});

if (require.main === module) {
  server.listen(PORT, HOST, () => {
    console.log(`计费引擎服务运行在 http://${HOST}:${PORT}`);
    console.log(`可用端点:`);
    console.log(`  POST /customers          - 创建客户`);
    console.log(`  GET  /customers/:id      - 获取客户信息`);
    console.log(`  GET  /customers/:id/status - 检查客户状态`);
    console.log(`  POST /subscriptions      - 创建订阅`);
    console.log(`  POST /subscriptions/upgrade - 升级订阅`);
    console.log(`  POST /subscriptions/refund - 申请退款`);
    console.log(`  POST /usage              - 记录用量`);
    console.log(`  GET  /usage/:customerId  - 查询用量`);
    console.log(`  POST /bills              - 生成账单`);
    console.log(`  GET  /bills?customerId=  - 查询账单列表`);
    console.log(`  GET  /bills/:id          - 查询账单详情`);
    console.log(`  POST /bills/pay          - 支付账单`);
    console.log(`  POST /bills/review       - 申请复核`);
  });
}

export { server, createServer, requestHandler };
