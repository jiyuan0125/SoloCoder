import express from 'express';
import { initDatabase } from './database';
import invoiceRoutes from './routes/invoiceRoutes';
import contractRoutes from './routes/contractRoutes';

const app = express();
const PORT = process.env.PORT ? parseInt(process.env.PORT, 10) : 3000;

initDatabase();

app.use(express.json());

app.get('/health', (req, res) => {
  res.json({ status: 'ok', timestamp: new Date().toISOString() });
});

app.use('/api/invoices', invoiceRoutes);
app.use('/api/contracts', contractRoutes);

app.use((req, res) => {
  res.status(404).json({ error: '路径不存在' });
});

app.use((err: Error, req: express.Request, res: express.Response, next: express.NextFunction) => {
  console.error(err.stack);
  res.status(500).json({ error: '服务器内部错误' });
});

app.listen(PORT, () => {
  console.log(`发票管理系统服务器已启动：http://localhost:${PORT}`);
  console.log(`API 端点：`);
  console.log(`  GET  /health`);
  console.log(`  POST /api/invoices - 创建发票`);
  console.log(`  POST /api/invoices/red - 红冲发票`);
  console.log(`  POST /api/invoices/verify - 查验发票`);
  console.log(`  GET  /api/invoices - 查看所有发票`);
  console.log(`  GET  /api/invoices/:id - 按ID查看发票`);
  console.log(`  GET  /api/invoices/code/:code/number/:number - 按代码和号码查看发票`);
  console.log(`  POST /api/contracts - 创建合同`);
  console.log(`  GET  /api/contracts - 查看所有合同`);
});
