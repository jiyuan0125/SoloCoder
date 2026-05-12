import express, { Express, Request, Response, NextFunction } from 'express';
import { initDatabase, closeDatabase } from './database';
import documentsRouter from './routes/documents';
import searchRouter from './routes/search';
import dictionaryRouter from './routes/dictionary';

const app: Express = express();
const PORT = process.env.PORT ? parseInt(process.env.PORT) : 3000;

app.use(express.json());
app.use(express.urlencoded({ extended: true }));

app.use((req: Request, res: Response, next: NextFunction) => {
  res.setHeader('Content-Type', 'application/json');
  next();
});

app.get('/', (req: Request, res: Response) => {
  res.json({
    name: 'Search Service',
    version: '1.0.0',
    endpoints: {
      documents: '/api/documents',
      search: '/api/search',
      dictionary: '/api/dictionary'
    }
  });
});

app.use('/api/documents', documentsRouter);
app.use('/api/search', searchRouter);
app.use('/api/dictionary', dictionaryRouter);

app.use((err: Error, req: Request, res: Response, next: NextFunction) => {
  console.error('Server error:', err);
  res.status(500).json({
    success: false,
    message: '服务器内部错误'
  });
});

app.use((req: Request, res: Response) => {
  res.status(404).json({
    success: false,
    message: '接口不存在'
  });
});

function gracefulShutdown() {
  console.log('正在关闭服务器...');
  closeDatabase();
  process.exit(0);
}

process.on('SIGTERM', gracefulShutdown);
process.on('SIGINT', gracefulShutdown);

if (require.main === module) {
  try {
    initDatabase();
    console.log('数据库初始化成功');
    
    app.listen(PORT, () => {
      console.log(`搜索服务已启动，监听端口: ${PORT}`);
      console.log('API 文档:');
      console.log('  - 文档管理: http://localhost:' + PORT + '/api/documents');
      console.log('  - 搜索接口: http://localhost:' + PORT + '/api/search');
      console.log('  - 词典管理: http://localhost:' + PORT + '/api/dictionary');
    });
  } catch (error) {
    console.error('启动失败:', error);
    process.exit(1);
  }
}

export default app;
