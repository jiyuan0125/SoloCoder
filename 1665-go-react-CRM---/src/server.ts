import express from 'express';
import { PoolService } from './services/poolService';
import leadsRouter from './routes/leads';
import opportunitiesRouter from './routes/opportunities';
import { db } from './database';

const app = express();
const PORT = process.env.PORT || 3000;

app.use(express.json());
app.use(express.urlencoded({ extended: true }));

app.use((req, res, next) => {
  console.log(`${new Date().toISOString()} - ${req.method} ${req.url}`);
  next();
});

app.use('/api/leads', leadsRouter);
app.use('/api/opportunities', opportunitiesRouter);

app.get('/api/health', (req, res) => {
  res.json({ status: 'ok', timestamp: Date.now() });
});

app.get('/api/sales-persons', (req, res) => {
  const persons = db.prepare('SELECT * FROM sales_persons').all();
  res.json(persons);
});

app.post('/api/pool/process-expired', (req, res) => {
  try {
    const result = PoolService.processExpiredLeads();
    res.json(result);
  } catch (error: any) {
    res.status(500).json({ error: error.message });
  }
});

app.use((err: Error, req: express.Request, res: express.Response, next: express.NextFunction) => {
  console.error('Server error:', err);
  res.status(500).json({ error: '服务器内部错误' });
});

const poolInterval = setInterval(() => {
  try {
    const result = PoolService.processExpiredLeads();
    if (result.moved > 0) {
      console.log(`自动回收了 ${result.moved} 个过期线索到公共池`);
    }
  } catch (error) {
    console.error('自动回收过期线索时出错:', error);
  }
}, 60 * 60 * 1000);

app.listen(PORT, () => {
  console.log(`CRM 后端服务正在运行，端口: ${PORT}`);
  console.log('健康检查: GET /api/health');
  console.log('线索接口: /api/leads');
  console.log('商机接口: /api/opportunities');
  console.log('销售人员列表: GET /api/sales-persons');
  console.log('手动处理过期线索: POST /api/pool/process-expired');
});

const gracefulShutdown = () => {
  console.log('正在关闭服务器...');
  clearInterval(poolInterval);
  db.close();
  process.exit(0);
};

process.on('SIGTERM', gracefulShutdown);
process.on('SIGINT', gracefulShutdown);
