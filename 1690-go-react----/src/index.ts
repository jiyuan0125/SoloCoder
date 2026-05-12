import express, { Request, Response } from 'express';
import recruitmentRouter from './routes/recruitment';
import trainingRouter from './routes/training';
import incentiveRouter from './routes/incentive';
import { getAllNotifications } from './services/notification';

const app = express();
const PORT = process.env.PORT || 3000;

app.use(express.json());

app.use('/api/recruitment', recruitmentRouter);
app.use('/api/training', trainingRouter);
app.use('/api/incentive', incentiveRouter);

app.get('/api/health', (req: Request, res: Response) => {
  res.json({ status: 'ok', timestamp: new Date().toISOString() });
});

app.get('/api/notifications', (req: Request, res: Response) => {
  res.json(getAllNotifications());
});

app.listen(PORT, () => {
  console.log(`志愿者管理平台服务已启动: http://localhost:${PORT}`);
  console.log(`API 端点:`);
  console.log(`  - GET  /api/health - 健康检查`);
  console.log(`  - 招募管理: /api/recruitment/*`);
  console.log(`  - 培训管理: /api/training/*`);
  console.log(`  - 激励管理: /api/incentive/*`);
});

export default app;
