export type MessageStatus = 'available' | 'processing' | 'acknowledged' | 'dead-letter';

export interface Message {
  id: string;
  status: MessageStatus;
  content: any;
  createdAt: number;
  updatedAt: number;
  retryCount: number;
  consumerId?: string;
  consumedAt?: number;
}

export interface Queue {
  name: string;
  maxSize: number;
  retentionTime: number;
  messages: Map<string, Message>;
  availableQueue: string[];
  deadLetterQueue: string[];
  createdAt: number;
}

export interface QueueInfo {
  name: string;
  maxSize: number;
  retentionTime: number;
  currentSize: number;
  deadLetterSize: number;
}

export interface CreateQueueRequest {
  maxSize?: number;
  retentionTime?: number;
}

export interface ConsumeMessagesRequest {
  consumerId?: string;
  batchSize?: number;
}

export interface SendMessageRequest {
  content: any;
}
