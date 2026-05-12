import express, { Express, Request, Response } from 'express';
import cors from 'cors';
import tracesRoutes from './routes/traces';
import { store } from './storage/InMemoryStore';

const app: Express = express();
const PORT = parseInt(process.env.PORT || '3000', 10);

app.use(cors());
app.use(express.json());

app.get('/health', (_req: Request, res: Response) => {
  res.json({ status: 'ok', timestamp: Date.now() });
});

app.use('/', tracesRoutes);

const CLEANUP_INTERVAL = 60 * 60 * 1000;

let cleanupTimer: NodeJS.Timeout | null = null;

const runCleanup = () => {
  const now = Date.now();
  const removedCount = store.cleanupExpiredData(now);
  if (removedCount > 0) {
    console.log(`[Cleanup] Removed ${removedCount} expired spans at ${new Date(now).toISOString()}`);
  }
};

const startCleanupJob = () => {
  if (cleanupTimer) {
    clearInterval(cleanupTimer);
  }
  cleanupTimer = setInterval(runCleanup, CLEANUP_INTERVAL);
  console.log(`[Cleanup] Scheduled to run every ${CLEANUP_INTERVAL / 1000} seconds`);
};

const server = app.listen(PORT, () => {
  console.log(`Distributed Tracing System is running on port ${PORT}`);
  startCleanupJob();
});

process.on('SIGTERM', () => {
  console.log('Received SIGTERM, shutting down gracefully...');
  if (cleanupTimer) {
    clearInterval(cleanupTimer);
  }
  server.close(() => {
    console.log('Server closed');
    process.exit(0);
  });
});

process.on('SIGINT', () => {
  console.log('Received SIGINT, shutting down gracefully...');
  if (cleanupTimer) {
    clearInterval(cleanupTimer);
  }
  server.close(() => {
    console.log('Server closed');
    process.exit(0);
  });
});

export default app;
