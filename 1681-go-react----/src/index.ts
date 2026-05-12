import express from 'express';
import cors from 'cors';
import './db';
import { usersRouter } from './routes/users';
import { requestsRouter } from './routes/requests';
import { statsRouter } from './routes/stats';

const app = express();
const PORT = process.env.PORT || 8111;

app.use(cors());
app.use(express.json());

app.get('/health', (_req, res) => {
  res.json({ status: 'ok', timestamp: Date.now() });
});

app.use('/api/users', usersRouter);
app.use('/api/requests', requestsRouter);
app.use('/api/stats', statsRouter);

app.listen(PORT, () => {
  console.log(`邻里互助平台服务运行在端口 ${PORT}`);
  console.log(`健康检查: http://localhost:${PORT}/health`);
  console.log(`API 端点:`);
  console.log(`  - POST /api/users - 创建用户`);
  console.log(`  - GET  /api/users - 获取所有用户`);
  console.log(`  - GET  /api/users/:id - 获取用户详情`);
  console.log(`  - POST /api/requests - 发布需求`);
  console.log(`  - GET  /api/requests - 获取所有需求`);
  console.log(`  - GET  /api/requests/:id - 获取需求详情`);
  console.log(`  - POST /api/requests/:id/accept - 接单`);
  console.log(`  - POST /api/requests/:id/cancel - 取消订单`);
  console.log(`  - POST /api/requests/:id/complete - 完成需求`);
  console.log(`  - POST /api/requests/:id/rate - 评价`);
  console.log(`  - POST /api/requests/:id/match - 智能匹配`);
  console.log(`  - GET  /api/stats/monthly - 月度统计`);
  console.log(`  - GET  /api/stats/building-activity - 楼栋活跃度`);
  console.log(`  - GET  /api/stats/warnings - 预警列表`);
});
