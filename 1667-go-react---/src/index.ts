import express, { Application } from 'express';
import { initDatabase } from './database';
import templatesRouter from './routes/templates';
import contractsRouter from './routes/contracts';

const app: Application = express();
const PORT = process.env.PORT || 3000;

app.use(express.json());

initDatabase();

app.get('/health', (req, res) => {
  res.json({ status: 'ok', timestamp: new Date().toISOString() });
});

app.use('/api/templates', templatesRouter);
app.use('/api/contracts', contractsRouter);

app.use((err: Error, req: express.Request, res: express.Response, next: express.NextFunction) => {
  console.error(err.stack);
  res.status(500).json({ error: '服务器内部错误' });
});

if (require.main === module) {
  app.listen(PORT, () => {
    console.log(`合同管理平台服务运行在 http://localhost:${PORT}`);
  });
}

export default app;
