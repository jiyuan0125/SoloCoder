import { Router, Request, Response } from 'express';
import { store } from '../store';
import { deliveryManager } from '../delivery';
import { Event, CreateEventRequest } from '../types';

const router = Router();

function generateId(): string {
  return `${Date.now()}-${Math.random().toString(36).slice(2, 10)}`;
}

router.post('/', (req: Request, res: Response) => {
  const body = req.body as Partial<CreateEventRequest>;

  if (!body.id || !body.topic) {
    return res.status(400).json({ error: 'Event id and topic are required' });
  }

  const existing = store.getEvent(body.id);
  if (existing) {
    return res.status(208).json({ message: 'Event already exists', id: body.id });
  }

  const event: Event = {
    id: body.id,
    topic: body.topic,
    eventType: body.eventType || '',
    payload: body.payload || {},
    publishedAt: body.publishedAt || new Date().toISOString(),
    createdAt: Date.now(),
  };

  const added = store.addEvent(event);
  if (!added) {
    return res.status(208).json({ message: 'Event already exists', id: body.id });
  }

  deliveryManager.createDeliveriesForEvent(event);

  return res.status(201).json({ id: event.id, topic: event.topic });
});

router.get('/:eventId/deliveries', (req: Request, res: Response) => {
  const { eventId } = req.params;

  const event = store.getEvent(eventId);
  if (!event) {
    return res.status(404).json({ error: 'Event not found or expired' });
  }

  const deliveries = store.getDeliveriesByEventId(eventId);

  const response = deliveries.map(d => ({
    id: d.id,
    subscriptionId: d.subscriptionId,
    status: d.status,
    retryCount: d.retryCount,
  }));

  return res.status(200).json({ eventId, deliveries: response });
});

export default router;
