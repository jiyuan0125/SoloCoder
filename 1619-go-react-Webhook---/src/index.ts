import express, { Request, Response } from 'express';
import webhooksRouter from './routes/webhooks';
import { startDeliveryLoop, emitEvent } from './deliverer';
import { SUPPORTED_EVENT_TYPES } from './types';

const app = express();
const PORT = process.env.PORT ? parseInt(process.env.PORT) : 3000;

app.use(express.json());

app.get('/health', (req: Request, res: Response) => {
  res.json({ status: 'ok', timestamp: Date.now() });
});

app.post('/emit-event', (req: Request, res: Response) => {
  const { event_type, data } = req.body;

  if (!event_type || typeof event_type !== 'string') {
    return res.status(400).json({ error: 'event_type is required and must be a string' });
  }

  if (!SUPPORTED_EVENT_TYPES.includes(event_type)) {
    return res.status(400).json({ error: 'Invalid event_type' });
  }

  emitEvent(event_type, data || {});
  res.status(202).json({ message: 'Event emitted' });
});

app.use('/webhooks', webhooksRouter);

app.use((err: any, req: Request, res: Response, next: any) => {
  console.error('Error:', err);
  res.status(500).json({ error: 'Internal server error' });
});

app.listen(PORT, () => {
  console.log(`Webhook service listening on port ${PORT}`);
  startDeliveryLoop();
});

process.on('SIGTERM', () => {
  process.exit(0);
});

process.on('SIGINT', () => {
  process.exit(0);
});
