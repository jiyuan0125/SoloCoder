import * as http from 'http';
import { handleRequest } from './router';

const PORT = parseInt(process.env.PORT || '3000', 10);
const HOST = process.env.HOST || 'localhost';

const server = http.createServer(async (req, res) => {
  await handleRequest(req, res);
});

server.listen(PORT, HOST, () => {
  console.log(`Commission Tracker API Server running at http://${HOST}:${PORT}`);
  console.log(`API Endpoints:`);
  console.log(`  POST   /api/salespeople          - 创建销售`);
  console.log(`  GET    /api/salespeople          - 销售列表`);
  console.log(`  GET    /api/salespeople/:id      - 获取销售详情`);
  console.log(`  POST   /api/salespeople/:id/resign - 销售离职`);
  console.log(`  GET    /api/salespeople/:id/commission - 获取佣金计算`);
  console.log(`  GET    /api/salespeople/:id/balance - 获取佣金余额`);
  console.log(`  POST   /api/orders               - 创建订单`);
  console.log(`  GET    /api/orders               - 订单列表`);
  console.log(`  GET    /api/orders/:id           - 获取订单详情`);
  console.log(`  POST   /api/orders/:id/refund    - 订单退款`);
  console.log(`  POST   /api/settlements/trigger  - 触发月度结算`);
  console.log(`  GET    /api/settlements          - 结算记录列表`);
  console.log(`  GET    /api/settlements/:id      - 获取结算单`);
  console.log(`  GET    /api/rankings             - 业绩排名`);
});

export { server };
