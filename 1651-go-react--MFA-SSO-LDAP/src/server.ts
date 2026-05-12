import express from 'express';
import authRoutes from './routes/auth';
import { config } from './config';

const app = express();

app.use(express.json());
app.use(express.urlencoded({ extended: true }));

app.use('/api/auth', authRoutes);

app.get('/health', (req, res) => {
  res.json({ status: 'ok', timestamp: Date.now() });
});

app.use((err: Error, req: express.Request, res: express.Response, next: express.NextFunction) => {
  console.error('Server error:', err);
  res.status(500).json({ error: '服务器内部错误' });
});

app.listen(config.port, () => {
  console.log(`Auth Center server running on http://localhost:${config.port}`);
  console.log('Default apps registered:');
  console.log('  - app1: http://localhost:3001/callback');
  console.log('  - app2: http://localhost:3002/callback');
});
