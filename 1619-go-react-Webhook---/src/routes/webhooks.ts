import { Router, Request, Response } from 'express';
import {
  createWebhook,
  getWebhookById,
  listWebhooks,
  updateWebhook,
  deleteWebhook,
  hasPendingDeliveries,
  getDeliveriesByWebhookId,
  getDeliveryById,
} from '../database';
import {
  generateSecret,
  isValidEventTypes,
  isValidCallbackUrl,
} from '../utils';

const router = Router();

function webhookResponse(webhook: any) {
  return {
    id: webhook.id,
    event_type: webhook.event_type,
    callback_url: webhook.callback_url,
    secret: webhook.secret,
    is_active: webhook.is_active === 1,
    created_at: webhook.created_at,
    updated_at: webhook.updated_at,
  };
}

router.post('/', (req: Request, res: Response) => {
  const { event_type, callback_url } = req.body;

  if (!event_type || typeof event_type !== 'string') {
    return res.status(400).json({ error: 'event_type is required and must be a string' });
  }

  if (!callback_url || typeof callback_url !== 'string') {
    return res.status(400).json({ error: 'callback_url is required and must be a string' });
  }

  if (!isValidEventTypes(event_type)) {
    return res.status(400).json({ error: 'Invalid event_type' });
  }

  if (!isValidCallbackUrl(callback_url)) {
    return res.status(400).json({ error: 'callback_url must be HTTPS, or localhost/127.0.0.1' });
  }

  const secret = generateSecret();
  const webhook = createWebhook(event_type, callback_url, secret);

  res.status(201).json(webhookResponse(webhook));
});

router.get('/', (req: Request, res: Response) => {
  const webhooks = listWebhooks();
  res.json(webhooks.map(webhookResponse));
});

router.get('/:id', (req: Request, res: Response) => {
  const webhook = getWebhookById(req.params.id);
  if (!webhook) {
    return res.status(404).json({ error: 'Webhook not found' });
  }
  res.json(webhookResponse(webhook));
});

router.put('/:id', (req: Request, res: Response) => {
  const { callback_url, is_active } = req.body;

  const webhook = getWebhookById(req.params.id);
  if (!webhook) {
    return res.status(404).json({ error: 'Webhook not found' });
  }

  if (callback_url !== undefined && typeof callback_url !== 'string') {
    return res.status(400).json({ error: 'callback_url must be a string' });
  }

  if (is_active !== undefined && typeof is_active !== 'boolean') {
    return res.status(400).json({ error: 'is_active must be a boolean' });
  }

  if (callback_url !== undefined && !isValidCallbackUrl(callback_url)) {
    return res.status(400).json({ error: 'callback_url must be HTTPS, or localhost/127.0.0.1' });
  }

  const updated = updateWebhook(
    req.params.id,
    callback_url,
    is_active
  );

  if (!updated) {
    return res.status(404).json({ error: 'Webhook not found' });
  }

  res.json(webhookResponse(updated));
});

router.delete('/:id', (req: Request, res: Response) => {
  const webhook = getWebhookById(req.params.id);
  if (!webhook) {
    return res.status(404).json({ error: 'Webhook not found' });
  }

  if (hasPendingDeliveries(req.params.id)) {
    return res.status(409).json({ error: 'Webhook has pending deliveries' });
  }

  deleteWebhook(req.params.id);
  res.status(204).send();
});

router.get('/:id/deliveries', (req: Request, res: Response) => {
  const webhook = getWebhookById(req.params.id);
  if (!webhook) {
    return res.status(404).json({ error: 'Webhook not found' });
  }

  const deliveries = getDeliveriesByWebhookId(req.params.id);
  res.json(deliveries.map(d => ({
    id: d.id,
    webhook_id: d.webhook_id,
    event_type: d.event_type,
    request_url: d.request_url,
    request_body: d.request_body,
    response_status: d.response_status,
    response_body: d.response_body,
    status: d.status,
    retry_count: d.retry_count,
    next_retry_at: d.next_retry_at,
    last_delivered_at: d.last_delivered_at,
    created_at: d.created_at,
  })));
});

router.post('/:id/deliveries/:deliveryId/retry', (req: Request, res: Response) => {
  const webhook = getWebhookById(req.params.id);
  if (!webhook) {
    return res.status(404).json({ error: 'Webhook not found' });
  }

  const delivery = getDeliveryById(req.params.deliveryId);
  if (!delivery) {
    return res.status(404).json({ error: 'Delivery not found' });
  }

  if (delivery.webhook_id !== req.params.id) {
    return res.status(404).json({ error: 'Delivery not found for this webhook' });
  }

  const { triggerManualRetry } = require('../deliverer');
  triggerManualRetry(delivery.id);

  res.status(202).json({ message: 'Retry scheduled' });
});

export default router;
