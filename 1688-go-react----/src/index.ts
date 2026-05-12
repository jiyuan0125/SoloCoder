import express, { Application, Request, Response, NextFunction } from 'express';
import { initDatabase } from './database';
import scalesRouter from './routes/scales';
import assessmentsRouter from './routes/assessments';
import alertsRouter from './routes/alerts';
import statisticsRouter from './routes/statistics';

const app: Application = express();
const PORT = process.env.PORT ? parseInt(process.env.PORT, 10) : 8118;

app.use(express.json());

app.use('/api/scales', scalesRouter);
app.use('/api/assessments', assessmentsRouter);
app.use('/api/alerts', alertsRouter);
app.use('/api/statistics', statisticsRouter);

app.get('/health', (req: Request, res: Response) => {
  res.json({ status: 'ok', timestamp: Date.now() });
});

app.use((err: Error, req: Request, res: Response, next: NextFunction) => {
  console.error('Unhandled error:', err);
  res.status(500).json({ error: 'Internal Server Error' });
});

initDatabase().then(() => {
  app.listen(PORT, () => {
    console.log(`Psychological Assessment System listening on port ${PORT}`);
  });
}).catch((err) => {
  console.error('Failed to initialize database:', err);
  process.exit(1);
});
