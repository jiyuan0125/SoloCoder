import { Event, Subscription, Delivery, DeliveryStatus } from '../types';

const SEVEN_DAYS_MS = 7 * 24 * 60 * 60 * 1000;
const THIRTY_DAYS_MS = 30 * 24 * 60 * 60 * 1000;

class Store {
  private events: Map<string, Event> = new Map();
  private subscriptions: Map<string, Subscription> = new Map();
  private deliveries: Map<string, Delivery> = new Map();
  private subscriptionsByTopic: Map<string, Set<string>> = new Map();
  private deliveriesByEventId: Map<string, Set<string>> = new Map();

  addEvent(event: Event): boolean {
    if (this.events.has(event.id)) {
      return false;
    }
    this.events.set(event.id, event);
    return true;
  }

  getEvent(id: string): Event | undefined {
    return this.events.get(id);
  }

  addSubscription(subscription: Subscription): void {
    this.subscriptions.set(subscription.id, subscription);
    if (!this.subscriptionsByTopic.has(subscription.topic)) {
      this.subscriptionsByTopic.set(subscription.topic, new Set());
    }
    this.subscriptionsByTopic.get(subscription.topic)!.add(subscription.id);
  }

  getSubscription(id: string): Subscription | undefined {
    return this.subscriptions.get(id);
  }

  getSubscriptionsByTopic(topic: string): Subscription[] {
    const ids = this.subscriptionsByTopic.get(topic);
    if (!ids) return [];
    return Array.from(ids)
      .map(id => this.subscriptions.get(id))
      .filter((s): s is Subscription => s !== undefined);
  }

  updateSubscription(id: string, updates: Partial<Subscription>): Subscription | undefined {
    const existing = this.subscriptions.get(id);
    if (!existing) return undefined;
    const updated = { ...existing, ...updates };
    this.subscriptions.set(id, updated);
    return updated;
  }

  deleteSubscription(id: string): boolean {
    const subscription = this.subscriptions.get(id);
    if (!subscription) return false;
    this.subscriptions.delete(id);
    const topicSubs = this.subscriptionsByTopic.get(subscription.topic);
    if (topicSubs) {
      topicSubs.delete(id);
    }
    return true;
  }

  addDelivery(delivery: Delivery): void {
    this.deliveries.set(delivery.id, delivery);
    if (!this.deliveriesByEventId.has(delivery.eventId)) {
      this.deliveriesByEventId.set(delivery.eventId, new Set());
    }
    this.deliveriesByEventId.get(delivery.eventId)!.add(delivery.id);
  }

  getDelivery(id: string): Delivery | undefined {
    return this.deliveries.get(id);
  }

  getDeliveriesByEventId(eventId: string): Delivery[] {
    const ids = this.deliveriesByEventId.get(eventId);
    if (!ids) return [];
    return Array.from(ids)
      .map(id => this.deliveries.get(id))
      .filter((d): d is Delivery => d !== undefined);
  }

  updateDelivery(id: string, updates: Partial<Delivery>): Delivery | undefined {
    const existing = this.deliveries.get(id);
    if (!existing) return undefined;
    const updated = { ...existing, ...updates };
    this.deliveries.set(id, updated);
    return updated;
  }

  cleanupExpired(): void {
    const now = Date.now();
    
    for (const [deliveryId, delivery] of this.deliveries.entries()) {
      if (delivery.expiresAt <= now) {
        this.deliveries.delete(deliveryId);
        const eventDeliveries = this.deliveriesByEventId.get(delivery.eventId);
        if (eventDeliveries) {
          eventDeliveries.delete(deliveryId);
          if (eventDeliveries.size === 0) {
            this.deliveriesByEventId.delete(delivery.eventId);
          }
        }
      }
    }

    for (const [eventId, event] of this.events.entries()) {
      const deliveries = this.getDeliveriesByEventId(eventId);
      const allDelivered = deliveries.every(d => d.status === 'delivered');
      const allFailed = deliveries.every(d => d.status === 'failed');
      
      let shouldRemove = false;
      if (deliveries.length === 0) {
        if (now - event.createdAt > SEVEN_DAYS_MS) {
          shouldRemove = true;
        }
      } else if (allDelivered) {
        const lastCreated = Math.max(...deliveries.map(d => d.createdAt));
        if (now - lastCreated > SEVEN_DAYS_MS) {
          shouldRemove = true;
        }
      } else if (allFailed) {
        const lastCreated = Math.max(...deliveries.map(d => d.createdAt));
        if (now - lastCreated > THIRTY_DAYS_MS) {
          shouldRemove = true;
        }
      }

      if (shouldRemove) {
        this.events.delete(eventId);
      }
    }
  }
}

export const store = new Store();
export { SEVEN_DAYS_MS, THIRTY_DAYS_MS };
