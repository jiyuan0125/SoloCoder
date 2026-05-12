import express, { Request, Response, NextFunction } from 'express';
import fs from 'fs';
import path from 'path';
import { initDatabase } from './db';
import salespeopleRoutes from './routes/salespeople';
import opportunitiesRoutes from './routes/opportunities';
import funnelRoutes from './routes/funnel';

const dataDir = path.join(process.cwd(), 'data');
if (!fs.existsSync(dataDir)) {
  fs.mkdirSync(dataDir, { recursive: true });
}

initDatabase();

const app = express();
const PORT = process.env.PORT ? parseInt(process.env.PORT, 10) : 3000;

app.use(express.json());

app.use((_req: Request, res: Response, next: NextFunction) => {
  res.setHeader('Content-Type', 'application/json');
  next();
});

app.use('/api/salespeople', salespeopleRoutes);
app.use('/api/opportunities', opportunitiesRoutes);
app.use('/api/funnel', funnelRoutes);

app.get('/health', (_req: Request, res: Response) => {
  res.json({ status: 'ok' });
});

app.use((err: Error, _req: Request, res: Response, _next: NextFunction) => {
  console.error('Unhandled error:', err);
  res.status(500).json({ error: '服务器内部错误' });
});

app.use((_req: Request, res: Response) => {
  res.status(404).json({ error: '接口不存在' });
});

if (require.main === module) {
  app.listen(PORT, () => {
    console.log(`Server is running on port ${PORT}`);
  });
}

export default app;
