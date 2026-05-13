import express, { Request, Response, NextFunction } from 'express';
import { initDatabase } from './db';
import rulesRouter from './routes/rules';
import riskRouter from './routes/risk';
import reviewsRouter from './routes/reviews';

const app = express();
const PORT = process.env.PORT || 8202;

initDatabase();

app.use(express.json());

app.use((req: Request, res: Response, next: NextFunction) => {
  console.log(`${new Date().toISOString()} - ${req.method} ${req.path}`);
  next();
});

app.use('/api/rules', rulesRouter);
app.use('/api/risk', riskRouter);
app.use('/api/reviews', reviewsRouter);

app.get('/health', (req: Request, res: Response) => {
  res.json({ status: 'ok', timestamp: Date.now() });
});

app.use((err: Error, req: Request, res: Response, next: NextFunction) => {
  console.error('Unhandled error:', err);
  res.status(500).json({ error: 'Internal server error' });
});

app.listen(PORT, () => {
  console.log(`Risk engine server running on port ${PORT}`);
});
