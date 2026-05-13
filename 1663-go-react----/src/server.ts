import express from 'express';
import { initializeDatabase } from './database';
import routes from './routes';

const app = express();
const PORT = process.env.PORT || 8203;

// 中间件
app.use(express.json());

// 初始化数据库
initializeDatabase();

// 路由
app.use('/api', routes);

// 健康检查
app.get('/health', (req, res) => {
  res.json({ status: 'OK', message: '联盟营销管理系统运行正常' });
});

// 错误处理中间件
app.use((err: Error, req: express.Request, res: express.Response, next: express.NextFunction) => {
  console.error(err.stack);
  res.status(500).json({ error: '服务器内部错误' });
});

app.listen(PORT, () => {
  console.log(`联盟营销管理系统服务器正在运行在 http://localhost:${PORT}`);
  console.log(`健康检查: http://localhost:${PORT}/health`);
});
