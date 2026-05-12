import axios from 'axios';
import {
  getWebhooksByEventType,
  createDelivery,
  getDeliveriesReadyToSend,
  getDeliveryById,
  getWebhookById,
  updateDeliveryStatus,
  getPendingDeliveries,
} from './database';
import { computeSignature, getNextRetryInterval, addMinutes } from './utils';
import { REQUEST_TIMEOUT, DeliveryStatus } from './types';

const processingWebhooks = new Set<string>();

export async function emitEvent(eventType: string, data: any): Promise<void> {
  const webhooks = getWebhooksByEventType(eventType);

  for (const webhook of webhooks) {
    const timestamp = Date.now();
    const payload = {
      event: eventType,
      data,
      timestamp,
      signature: '',
    };
    const payloadString = JSON.stringify(payload);
    const signature = computeSignature(timestamp, payloadString, webhook.secret);

    const finalPayload = {
      event: eventType,
      data,
      timestamp,
      signature,
    };
    const finalPayloadString = JSON.stringify(finalPayload);

    createDelivery(
      webhook.id,
      eventType,
      JSON.stringify(data),
      webhook.callback_url,
      finalPayloadString
    );
  }

  scheduleDeliveryLoop();
}

async function sendDelivery(deliveryId: string): Promise<void> {
  const delivery = getDeliveryById(deliveryId);
  if (!delivery) return;

  const webhook = getWebhookById(delivery.webhook_id);
  if (!webhook || webhook.is_active !== 1) return;

  try {
    const response = await axios.post(delivery.request_url, JSON.parse(delivery.request_body), {
      timeout: REQUEST_TIMEOUT,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    updateDeliveryStatus(
      delivery.id,
      'success',
      response.status,
      JSON.stringify(response.data),
      delivery.retry_count,
      null
    );
  } catch (error: any) {
    const responseStatus = error.response?.status || null;
    const responseBody = error.response?.data
      ? typeof error.response.data === 'string'
        ? error.response.data
        : JSON.stringify(error.response.data)
      : error.message;

    const newRetryCount = delivery.retry_count + 1;
    const nextInterval = getNextRetryInterval(newRetryCount);

    let newStatus: DeliveryStatus = 'failed';
    let nextRetryAt: string | null = null;

    if (nextInterval === null) {
      newStatus = 'final_failed';
    } else {
      nextRetryAt = addMinutes(delivery.last_delivered_at, nextInterval);
    }

    updateDeliveryStatus(
      delivery.id,
      newStatus,
      responseStatus,
      responseBody,
      newRetryCount,
      nextRetryAt
    );
  }
}

async function processWebhookQueue(webhookId: string): Promise<void> {
  if (processingWebhooks.has(webhookId)) return;
  processingWebhooks.add(webhookId);

  try {
    while (true) {
      const pending = getPendingDeliveries(webhookId);
      if (pending.length === 0) break;

      const delivery = pending[0];
      const now = new Date().toISOString();

      if (delivery.next_retry_at && delivery.next_retry_at > now) {
        break;
      }

      if (delivery.status === 'pending' || delivery.status === 'failed') {
        updateDeliveryStatus(delivery.id, 'processing');
        await sendDelivery(delivery.id);
      } else {
        break;
      }
    }
  } finally {
    processingWebhooks.delete(webhookId);
  }
}

async function deliveryLoop(): Promise<void> {
  const ready = getDeliveriesReadyToSend();
  const webhookIds = new Set(ready.map(d => d.webhook_id));

  for (const webhookId of webhookIds) {
    processWebhookQueue(webhookId).catch(() => {});
  }
}

let loopTimer: NodeJS.Timeout | null = null;

export function startDeliveryLoop(): void {
  if (loopTimer) return;
  loopTimer = setInterval(() => {
    deliveryLoop().catch(() => {});
  }, 1000);
  deliveryLoop().catch(() => {});
}

export function stopDeliveryLoop(): void {
  if (loopTimer) {
    clearInterval(loopTimer);
    loopTimer = null;
  }
}

function scheduleDeliveryLoop(): void {
  setImmediate(() => {
    deliveryLoop().catch(() => {});
  });
}

export function triggerManualRetry(deliveryId: string): void {
  const delivery = getDeliveryById(deliveryId);
  if (!delivery) return;

  const webhook = getWebhookById(delivery.webhook_id);
  if (!webhook || webhook.is_active !== 1) return;

  updateDeliveryStatus(
    delivery.id,
    'pending',
    undefined,
    undefined,
    0,
    null
  );

  scheduleDeliveryLoop();
}
