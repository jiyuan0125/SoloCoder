import express from 'express';
import eventsRouter from './routes/events';
import subscriptionsRouter from './routes/subscriptions';
import { deliveryManager } from './delivery';

const app = express();

app.use(express.json());
app.use(express.urlencoded({ extended: true }));

app.use('/events', eventsRouter);
app.use('/subscriptions', subscriptionsRouter);

app.use((err: Error, _req: express.Request, res: express.Response, _next: express.NextFunction) => {
  console.error('Server error:', err);
  res.status(500).json({ error: 'Internal server error' });
});

export { app, deliveryManager };
