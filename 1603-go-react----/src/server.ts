import express from 'express';
import fs from 'fs';
import path from 'path';
import { initDb } from './db/database';
import metricsRouter from './routes/metrics';
import dashboardsRouter from './routes/dashboards';
import chartCardsRouter from './routes/chartCards';

const dataDir = path.join(process.cwd(), 'data');
if (!fs.existsSync(dataDir)) {
  fs.mkdirSync(dataDir, { recursive: true });
}

initDb();

const app = express();
const PORT = process.env.PORT ? parseInt(process.env.PORT, 10) : 9310;

app.use(express.json());

app.get('/health', (req, res) => {
  res.json({ status: 'ok', timestamp: new Date().toISOString() });
});

app.use('/api/metrics', metricsRouter);
app.use('/api/dashboards', dashboardsRouter);
app.use('/api/chart-cards', chartCardsRouter);

app.listen(PORT, () => {
  console.log(`数据可视化平台服务已启动`);
  console.log(`监听端口: ${PORT}`);
  console.log(`健康检查: http://localhost:${PORT}/health`);
  console.log('');
  console.log('API 端点:');
  console.log('  指标:     http://localhost:${PORT}/api/metrics');
  console.log('  仪表盘:   http://localhost:${PORT}/api/dashboards');
  console.log('  图表卡片: http://localhost:${PORT}/api/chart-cards');
});
