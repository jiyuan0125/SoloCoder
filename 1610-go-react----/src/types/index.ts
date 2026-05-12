export type ChannelType = 'email' | 'sms' | 'inapp' | 'dingtalk';

export interface Template {
  id: string;
  name: string;
  type: ChannelType;
  content: string;
  variables: string[];
  createdAt: number;
  updatedAt: number;
}

export interface ChannelInstance {
  name: string;
  config: Record<string, any>;
}

export interface Channel {
  id: string;
  name: string;
  type: ChannelType;
  primary: ChannelInstance;
  backup?: ChannelInstance;
  activeInstance: 'primary' | 'backup';
  rateLimits: {
    perMinute: number;
    perHour: number;
    perDay: number;
  };
  status: 'active' | 'inactive' | 'alerting';
  createdAt: number;
  updatedAt: number;
}

export type NotificationStatus = 'pending' | 'sent' | 'failed' | 'delayed' | 'cancelled';

export interface Notification {
  id: string;
  channelId: string;
  templateId: string;
  userId: string;
  variables: Record<string, any>;
  status: NotificationStatus;
  scheduledAt: number;
  retries: number;
  createdAt: number;
  updatedAt: number;
}

export interface AlertRecord {
  channelId: string;
  startedAt: number;
  endedAt?: number;
  failureRate: number;
}
