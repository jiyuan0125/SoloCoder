import express, { Request, Response } from 'express';
import { initDatabase } from './database';
import projectsRouter from './routes/projects';
import auditRouter from './routes/audit';

const app = express();
const PORT = process.env.PORT || 3000;

app.use(express.json());

app.get('/health', (req: Request, res: Response) => {
  res.json({ status: 'ok', timestamp: new Date().toISOString() });
});

app.use('/api/projects', projectsRouter);
app.use('/api/audit-logs', auditRouter);

app.use((err: Error, req: Request, res: Response, next: express.NextFunction) => {
  console.error(err.stack);
  res.status(500).json({ error: '服务器内部错误' });
});

initDatabase();

app.listen(PORT, () => {
  console.log(`公益基金会项目管理系统服务已启动，监听端口 ${PORT}`);
});
