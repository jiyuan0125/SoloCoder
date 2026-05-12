export interface Event {
  id: string;
  topic: string;
  eventType: string;
  payload: Record<string, unknown>;
  publishedAt: string;
  createdAt: number;
}

export interface Subscription {
  id: string;
  topic: string;
  callbackUrl: string;
  isPaused: boolean;
  createdAt: number;
}

export type DeliveryStatus = 'pending' | 'delivering' | 'delivered' | 'failed';

export interface Delivery {
  id: string;
  eventId: string;
  subscriptionId: string;
  status: DeliveryStatus;
  retryCount: number;
  nextRetryAt?: number;
  lastAttemptAt?: number;
  createdAt: number;
  expiresAt: number;
}

export interface CreateEventRequest {
  id: string;
  topic: string;
  eventType?: string;
  payload?: Record<string, unknown>;
  publishedAt?: string;
}

export interface CreateSubscriptionRequest {
  topic: string;
  callbackUrl: string;
}
