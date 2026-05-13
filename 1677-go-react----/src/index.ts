import express from 'express';
import './database';
import projectsRouter from './routes/projects';
import donationsRouter from './routes/donations';
import vouchersRouter from './routes/vouchers';
import warningsRouter from './routes/warnings';
import { checkAndGenerateProjectWarnings } from './services/warning-service';

const app = express();
const PORT = process.env.PORT ? parseInt(process.env.PORT) : 8206;

app.use(express.json());

app.get('/health', (_req, res) => {
  res.json({ status: 'ok', timestamp: new Date().toISOString() });
});

app.use('/api/projects', projectsRouter);
app.use('/api/donations', donationsRouter);
app.use('/api/vouchers', vouchersRouter);
app.use('/api/warnings', warningsRouter);

app.use((_req, res) => {
  res.status(404).json({ error: 'Not Found' });
});

app.use((err: Error, _req: express.Request, res: express.Response, _next: express.NextFunction) => {
  console.error('Error:', err);
  res.status(500).json({ error: 'Internal Server Error' });
});

app.listen(PORT, () => {
  console.log(`Server is running on port ${PORT}`);
  checkAndGenerateProjectWarnings();
  
  setInterval(() => {
    checkAndGenerateProjectWarnings();
  }, 60 * 60 * 1000);
});
