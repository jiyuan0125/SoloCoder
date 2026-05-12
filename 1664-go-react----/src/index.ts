import express from 'express';
import { initDatabase } from './db';
import router from './routes';
import { startBroadcastEngine } from './services/broadcastEngine';

const app = express();
const PORT = process.env.PORT ? parseInt(process.env.PORT) : 3000;

app.use(express.json());

app.use('/api', router);

app.use((_req: express.Request, res: express.Response) => {
  res.status(404).json({ error: 'Not found' });
});

app.use((err: any, _req: express.Request, res: express.Response, _next: express.NextFunction) => {
  console.error('Unhandled error:', err);
  const status = err.status || 500;
  res.status(status).json({ error: err.message || 'Internal server error' });
});

initDatabase();
startBroadcastEngine();

app.listen(PORT, () => {
  console.log(`Server running on http://localhost:${PORT}`);
});
