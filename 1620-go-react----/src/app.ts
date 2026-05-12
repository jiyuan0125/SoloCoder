import express, { Request, Response, NextFunction } from 'express';
import tasksRouter from './routes/tasks';
import { initDatabase } from './database';

const app = express();

initDatabase();

app.use(express.json());

app.get('/health', (_req: Request, res: Response) => {
  res.json({ status: 'ok' });
});

app.use('/sync/tasks', tasksRouter);

app.use((err: any, _req: Request, res: Response, _next: NextFunction) => {
  console.error(err);
  res.status(500).json({
    success: false,
    error: 'Internal server error',
  });
});

export default app;
