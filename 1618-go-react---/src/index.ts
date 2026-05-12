import express, { Request, Response, NextFunction } from 'express';
import { initDatabase } from './database';
import developersRouter from './routes/developers';
import appsRouter from './routes/apps';
import statsRouter from './routes/stats';

const app = express();
const PORT = process.env.PORT || 9108;

app.use(express.json());

app.use((req: Request, res: Response, next: NextFunction) => {
  res.setHeader('Content-Type', 'application/json');
  next();
});

app.use('/developers', developersRouter);
app.use('/developers/:devId/apps', appsRouter);
app.use('/developers/:devId/apps/:appId/stats', statsRouter);

app.get('/health', (req: Request, res: Response) => {
  res.json({ status: 'ok', timestamp: new Date().toISOString() });
});

app.use((err: any, req: Request, res: Response, next: NextFunction) => {
  console.error('Unhandled error:', err);
  res.status(500).json({ error: 'Internal server error' });
});

app.use((req: Request, res: Response) => {
  res.status(404).json({ error: 'Route not found' });
});

async function startServer() {
  try {
    await initDatabase();
    console.log('Database initialized successfully');
    
    app.listen(PORT, () => {
      console.log(`Server running on port ${PORT}`);
      console.log(`Health check: http://localhost:${PORT}/health`);
    });
  } catch (err) {
    console.error('Failed to start server:', err);
    process.exit(1);
  }
}

startServer();

export default app;
