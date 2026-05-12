import express from 'express';
import { QueueManager } from './queueManager';
import { createRoutes } from './routes';

const PORT = process.env.PORT ? parseInt(process.env.PORT, 10) : 3000;
const TIMEOUT_CHECK_INTERVAL = 5000;

const app = express();
const queueManager = new QueueManager();

app.use(createRoutes(queueManager));

app.use((req: express.Request, res: express.Response) => {
  res.status(404).json({ error: 'Endpoint not found' });
});

const server = app.listen(PORT, () => {
  console.log(`Memory Message Queue server listening on port ${PORT}`);
});

const timeoutInterval = setInterval(() => {
  queueManager.processTimeouts();
}, TIMEOUT_CHECK_INTERVAL);

function shutdown() {
  console.log('Shutting down server...');
  clearInterval(timeoutInterval);
  server.close(() => {
    console.log('Server closed');
    process.exit(0);
  });
}

process.on('SIGTERM', shutdown);
process.on('SIGINT', shutdown);
