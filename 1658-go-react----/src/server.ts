import express from 'express';
import usersRouter from './routes/users';
import devicesRouter from './routes/devices';
import ordersRouter from './routes/orders';
import investigationRouter from './routes/investigation';

const app = express();
const PORT = process.env.PORT || 3000;

app.use(express.json());

app.get('/health', (_req, res) => {
  res.json({ status: 'ok', timestamp: Date.now() });
});

app.use('/api/users', usersRouter);
app.use('/api/devices', devicesRouter);
app.use('/api/orders', ordersRouter);
app.use('/api/investigation', investigationRouter);

app.use((_req, res) => {
  res.status(404).json({ error: 'Not found' });
});

app.use((err: any, _req: express.Request, res: express.Response, _next: express.NextFunction) => {
  console.error(err);
  res.status(500).json({ error: 'Internal server error' });
});

app.listen(PORT, () => {
  console.log(`Fraud Detection Platform running on port ${PORT}`);
  console.log(`Health check: http://localhost:${PORT}/health`);
});
