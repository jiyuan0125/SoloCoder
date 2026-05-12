export interface Webhook {
  id: string;
  event_type: string;
  callback_url: string;
  secret: string;
  is_active: number;
  created_at: string;
  updated_at: string;
}

export type DeliveryStatus = 'pending' | 'processing' | 'success' | 'failed' | 'final_failed';

export interface Delivery {
  id: string;
  webhook_id: string;
  event_type: string;
  event_data: string;
  request_url: string;
  request_body: string;
  response_status: number | null;
  response_body: string;
  status: DeliveryStatus;
  retry_count: number;
  next_retry_at: string | null;
  last_delivered_at: string | null;
  created_at: string;
}

export interface EventPayload {
  event: string;
  data: any;
  timestamp: number;
  signature: string;
}

export const RETRY_INTERVALS = [
  1 * 60 * 1000,
  5 * 60 * 1000,
  30 * 60 * 1000,
  2 * 60 * 60 * 1000,
  24 * 60 * 60 * 1000,
];

export const MAX_RETRIES = 5;
export const TIMESTAMP_WINDOW = 5 * 60 * 1000;
export const REQUEST_TIMEOUT = 10 * 1000;

export const SUPPORTED_EVENT_TYPES = [
  'order.created',
  'order.updated',
  'order.cancelled',
  'user.created',
  'user.updated',
  'payment.success',
  'payment.failed',
];
