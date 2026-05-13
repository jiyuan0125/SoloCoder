import express, { Request, Response, NextFunction } from 'express';
import mothersRouter from './routes/mothers';
import babiesRouter from './routes/babies';
import schedulingRouter from './routes/scheduling';

const app = express();
const PORT = process.env.PORT || 8300;

app.use(express.json());

app.get('/health', (_req: Request, res: Response): void => {
  res.json({ status: 'ok', timestamp: new Date().toISOString() });
});

app.use('/api/mothers', mothersRouter);
app.use('/api/babies', babiesRouter);
app.use('/api/scheduling', schedulingRouter);

app.use((_req: Request, res: Response): void => {
  res.status(404).json({ error: 'Not Found' });
});

app.use((err: Error, _req: Request, res: Response, _next: NextFunction): void => {
  console.error(err.stack);
  res.status(500).json({ error: 'Internal Server Error' });
});

app.listen(PORT, (): void => {
  console.log(`Maternity Center Management System running on port ${PORT}`);
  console.log(`Health check: http://localhost:${PORT}/health`);
});
