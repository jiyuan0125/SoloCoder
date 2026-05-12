import { Router, Request, Response } from 'express';
import { store } from '../store';
import { Subscription, CreateSubscriptionRequest } from '../types';

const router = Router();

function generateId(): string {
  return `${Date.now()}-${Math.random().toString(36).slice(2, 10)}`;
}

function isValidUrl(url: string): boolean {
  try {
    const parsed = new URL(url);
    return parsed.protocol === 'http:' || parsed.protocol === 'https:';
  } catch {
    return false;
  }
}

router.post('/', (req: Request, res: Response) => {
  const body = req.body as Partial<CreateSubscriptionRequest>;

  if (!body.topic || !body.callbackUrl) {
    return res.status(400).json({ error: 'Topic and callbackUrl are required' });
  }

  if (!isValidUrl(body.callbackUrl)) {
    return res.status(400).json({ error: 'Invalid callbackUrl format' });
  }

  const subscription: Subscription = {
    id: generateId(),
    topic: body.topic,
    callbackUrl: body.callbackUrl,
    isPaused: false,
    createdAt: Date.now(),
  };

  store.addSubscription(subscription);

  return res.status(201).json({
    id: subscription.id,
    topic: subscription.topic,
    callbackUrl: subscription.callbackUrl,
    isPaused: subscription.isPaused,
  });
});

router.put('/:id/pause', (req: Request, res: Response) => {
  const { id } = req.params;

  const subscription = store.getSubscription(id);
  if (!subscription) {
    return res.status(404).json({ error: 'Subscription not found' });
  }

  const updated = store.updateSubscription(id, { isPaused: true });
  if (!updated) {
    return res.status(500).json({ error: 'Failed to pause subscription' });
  }

  return res.status(200).json({
    id: updated.id,
    topic: updated.topic,
    isPaused: updated.isPaused,
  });
});

router.put('/:id/resume', (req: Request, res: Response) => {
  const { id } = req.params;

  const subscription = store.getSubscription(id);
  if (!subscription) {
    return res.status(404).json({ error: 'Subscription not found' });
  }

  const updated = store.updateSubscription(id, { isPaused: false });
  if (!updated) {
    return res.status(500).json({ error: 'Failed to resume subscription' });
  }

  return res.status(200).json({
    id: updated.id,
    topic: updated.topic,
    isPaused: updated.isPaused,
  });
});

router.delete('/:id', (req: Request, res: Response) => {
  const { id } = req.params;

  const deleted = store.deleteSubscription(id);
  if (!deleted) {
    return res.status(404).json({ error: 'Subscription not found' });
  }

  return res.status(204).send();
});

export default router;
