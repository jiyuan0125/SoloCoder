import express from 'express';
import './database';
import accountRoutes from './routes/accounts';
import transferRoutes from './routes/transfers';
import interestRoutes from './routes/interest';
import configRoutes from './routes/config';

const app = express();
const PORT = process.env.PORT || 8102;

app.use(express.json());

app.use('/api/accounts', accountRoutes);
app.use('/api/transfers', transferRoutes);
app.use('/api/interest', interestRoutes);
app.use('/api/config', configRoutes);

app.get('/health', (_req, res) => {
  res.json({ status: 'ok', timestamp: new Date().toISOString() });
});

app.listen(PORT, () => {
  console.log(`资金池管理系统运行在 http://localhost:${PORT}`);
});
