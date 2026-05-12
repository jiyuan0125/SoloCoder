import express, { Request, Response, NextFunction } from 'express';
import plansRouter from './routes/plans';
import subscriptionsRouter from './routes/subscriptions';
import usagesRouter from './routes/usages';
import billsRouter from './routes/bills';
import { startScheduler } from './scheduler';

const app = express();
const PORT = process.env.PORT || 3000;

app.use(express.json());

app.use('/api/plans', plansRouter);
app.use('/api/subscriptions', subscriptionsRouter);
app.use('/api/usages', usagesRouter);
app.use('/api/bills', billsRouter);

app.get('/health', (req: Request, res: Response) => {
  res.json({ status: 'ok' });
});

app.use((err: Error, req: Request, res: Response, next: NextFunction) => {
  console.error(err.stack);
  res.status(500).json({ error: '服务器内部错误' });
});

app.use('*', (req: Request, res: Response) => {
  res.status(404).json({ error: '接口不存在' });
});

app.listen(PORT, () => {
  console.log(`服务器已启动，监听端口 ${PORT}`);
  startScheduler();
});

export default app;
