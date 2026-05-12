import express from 'express';
import packagesRouter from './routes/packages';
import tenantsRouter from './routes/tenants';
import alertsRouter from './routes/alerts';
import usageRouter from './routes/usage';

const app = express();

app.use(express.json());

app.get('/health', (_req, res) => {
  res.json({ status: 'ok', timestamp: new Date().toISOString() });
});

app.use('/api/packages', packagesRouter);
app.use('/api/tenants', tenantsRouter);
app.use('/api/alerts', alertsRouter);
app.use('/api/usage', usageRouter);

app.use((_req, res) => {
  res.status(404).json({ message: '接口不存在' });
});

app.use((err: Error, _req: express.Request, res: express.Response, _next: express.NextFunction) => {
  console.error('Internal error:', err);
  res.status(500).json({ message: '服务器内部错误' });
});

export default app;
