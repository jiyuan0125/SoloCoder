export type AlertLevel = 'info' | 'warning' | 'critical' | 'emergency';

export type AlertStatus = 'open' | 'acknowledged' | 'resolved';

export type ChannelType = 'email' | 'sms' | 'webhook';

export interface Alert {
  id: string;
  name: string;
  level: AlertLevel;
  sourceSystem: string;
  description: string;
  metrics: Record<string, any>;
  status: AlertStatus;
  originalLevel: AlertLevel;
  count: number;
  createdAt: number;
  lastOccurrenceAt: number;
  acknowledgedAt: number | null;
  resolvedAt: number | null;
  escalationTime: number;
}

export interface AggregatedAlert {
  id: string;
  alertKey: string;
  name: string;
  level: AlertLevel;
  sourceSystem: string;
  count: number;
  descriptions: string[];
  metricsList: Record<string, any>[];
  windowEndTime: number;
  createdAt: number;
}

export interface ChannelConfig {
  type: ChannelType;
  enabled: boolean;
  rateLimitPerMinute: number;
  emailRecipients?: string[];
  smsNumbers?: string[];
  webhookUrl?: string;
}

export interface NotificationQueue {
  id: string;
  channelType: ChannelType;
  content: string;
  priority: number;
  scheduledTime: number;
  createdAt: number;
  sent: boolean;
  alertId?: string;
}
