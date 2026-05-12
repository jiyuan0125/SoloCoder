import axios from 'axios';
import { Event, Subscription, Delivery } from '../types';
import { store, SEVEN_DAYS_MS, THIRTY_DAYS_MS } from '../store';

const RETRY_INTERVALS_MS = [30 * 1000, 60 * 1000, 5 * 60 * 1000];

function generateId(): string {
  return `${Date.now()}-${Math.random().toString(36).slice(2, 10)}`;
}

class DeliveryManager {
  private running = false;
  private timer: NodeJS.Timeout | null = null;
  private cleanupTimer: NodeJS.Timeout | null = null;

  start(): void {
    if (this.running) return;
    this.running = true;
    this.timer = setInterval(() => this.processDeliveries(), 1000);
    this.cleanupTimer = setInterval(() => store.cleanupExpired(), 60 * 1000);
  }

  stop(): void {
    this.running = false;
    if (this.timer) {
      clearInterval(this.timer);
      this.timer = null;
    }
    if (this.cleanupTimer) {
      clearInterval(this.cleanupTimer);
      this.cleanupTimer = null;
    }
  }

  createDeliveriesForEvent(event: Event): void {
    const subscriptions = store.getSubscriptionsByTopic(event.topic);
    const now = Date.now();

    for (const sub of subscriptions) {
      if (sub.isPaused) {
        continue;
      }

      const delivery: Delivery = {
        id: generateId(),
        eventId: event.id,
        subscriptionId: sub.id,
        status: 'pending',
        retryCount: 0,
        createdAt: now,
        expiresAt: now + SEVEN_DAYS_MS,
      };
      store.addDelivery(delivery);
    }
  }

  private async processDeliveries(): Promise<void> {
    if (!this.running) return;

    const now = Date.now();
    const allDeliveries = this.getAllDeliveries();

    for (const delivery of allDeliveries) {
      if (delivery.status === 'delivered' || delivery.status === 'failed') {
        continue;
      }

      if (delivery.status === 'pending') {
        await this.attemptDelivery(delivery);
        continue;
      }

      if (delivery.status === 'delivering') {
        if (delivery.nextRetryAt && now >= delivery.nextRetryAt) {
          await this.attemptDelivery(delivery);
        }
      }
    }
  }

  private getAllDeliveries(): Delivery[] {
    const result: Delivery[] = [];
    const storeAny = store as unknown as { deliveries: Map<string, Delivery> };
    for (const delivery of storeAny.deliveries.values()) {
      result.push(delivery);
    }
    return result;
  }

  private async attemptDelivery(delivery: Delivery): Promise<void> {
    const event = store.getEvent(delivery.eventId);
    const subscription = store.getSubscription(delivery.subscriptionId);

    if (!event || !subscription) {
      store.updateDelivery(delivery.id, { status: 'failed' });
      return;
    }

    if (subscription.isPaused) {
      return;
    }

    store.updateDelivery(delivery.id, { status: 'delivering', lastAttemptAt: Date.now() });

    try {
      const response = await axios.post(
        subscription.callbackUrl,
        {
          eventId: event.id,
          topic: event.topic,
          eventType: event.eventType,
          payload: event.payload,
          publishedAt: event.publishedAt,
        },
        {
          timeout: 10000,
          validateStatus: (status) => status >= 200 && status < 300,
        }
      );

      if (response.status >= 200 && response.status < 300) {
        store.updateDelivery(delivery.id, {
          status: 'delivered',
          expiresAt: Date.now() + SEVEN_DAYS_MS,
        });
      }
    } catch (error) {
      const newRetryCount = delivery.retryCount + 1;

      if (newRetryCount >= 3) {
        store.updateDelivery(delivery.id, {
          status: 'failed',
          retryCount: newRetryCount,
          expiresAt: Date.now() + THIRTY_DAYS_MS,
        });
      } else {
        const nextInterval = RETRY_INTERVALS_MS[newRetryCount - 1];
        store.updateDelivery(delivery.id, {
          status: 'delivering',
          retryCount: newRetryCount,
          nextRetryAt: Date.now() + nextInterval,
        });
      }
    }
  }
}

export const deliveryManager = new DeliveryManager();
