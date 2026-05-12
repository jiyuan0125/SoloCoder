import express, { Express, Request, Response } from 'express';
import { handleError } from './middleware/errorHandler';
import { elderRoutes } from './routes/elders';
import { careRoutes } from './routes/care';
import { notificationRoutes } from './routes/notifications';
import { familyRoutes } from './routes/family';
import { getDb } from './database';

const app: Express = express();
const PORT = process.env.PORT ? parseInt(process.env.PORT, 10) : 8114;

app.use(express.json());
app.use(express.urlencoded({ extended: true }));

app.get('/health', (_req: Request, res: Response) => {
  res.json({ 
    success: true, 
    message: '养老院管理系统 API 运行正常',
    timestamp: new Date().toISOString()
  });
});

app.use('/api/elders', elderRoutes);
app.use('/api/care', careRoutes);
app.use('/api', notificationRoutes);
app.use('/api/family', familyRoutes);

app.use(handleError);

function startServer(): void {
  getDb();
  
  app.listen(PORT, () => {
    console.log(`养老院管理系统 API 服务运行在端口 ${PORT}`);
    console.log(`健康检查: http://localhost:${PORT}/health`);
  });
}

if (require.main === module) {
  startServer();
}

export { app, startServer };
