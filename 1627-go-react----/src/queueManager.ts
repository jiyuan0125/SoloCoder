import { v4 as uuidv4 } from 'uuid';
import { Queue, Message, MessageStatus, QueueInfo } from './types';

const MAX_BATCH_SIZE = 100;
const DEFAULT_MAX_SIZE = 1000;
const DEFAULT_RETENTION_TIME = 24 * 60 * 60 * 1000;
const ACK_TIMEOUT = 30 * 1000;
const MAX_RETRY = 3;

export class QueueAlreadyExistsError extends Error {
  constructor(queueName: string) {
    super(`Queue "${queueName}" already exists`);
    this.name = 'QueueAlreadyExistsError';
  }
}

export class QueueNotFoundError extends Error {
  constructor(queueName: string) {
    super(`Queue "${queueName}" not found`);
    this.name = 'QueueNotFoundError';
  }
}

export class MessageNotFoundError extends Error {
  constructor(messageId: string) {
    super(`Message "${messageId}" not found`);
    this.name = 'MessageNotFoundError';
  }
}

export class QueueFullError extends Error {
  public readonly currentSize: number;
  public readonly maxSize: number;

  constructor(queueName: string, currentSize: number, maxSize: number) {
    super(`Queue "${queueName}" is full: ${currentSize}/${maxSize}`);
    this.name = 'QueueFullError';
    this.currentSize = currentSize;
    this.maxSize = maxSize;
  }
}

export class InvalidStateTransitionError extends Error {
  constructor(from: MessageStatus, to: MessageStatus) {
    super(`Invalid state transition from "${from}" to "${to}"`);
    this.name = 'InvalidStateTransitionError';
  }
}

export class NotOwnerError extends Error {
  constructor(messageId: string, consumerId: string) {
    super(`Consumer "${consumerId}" is not the owner of message "${messageId}"`);
    this.name = 'NotOwnerError';
  }
}

export class QueueManager {
  private queues: Map<string, Queue> = new Map();

  public createQueue(
    name: string,
    maxSize: number = DEFAULT_MAX_SIZE,
    retentionTime: number = DEFAULT_RETENTION_TIME
  ): Queue {
    if (this.queues.has(name)) {
      throw new QueueAlreadyExistsError(name);
    }

    const queue: Queue = {
      name,
      maxSize,
      retentionTime,
      messages: new Map(),
      availableQueue: [],
      deadLetterQueue: [],
      createdAt: Date.now()
    };

    this.queues.set(name, queue);
    return queue;
  }

  public deleteQueue(name: string): void {
    this.queues.delete(name);
  }

  public getQueue(name: string): Queue {
    const queue = this.queues.get(name);
    if (!queue) {
      throw new QueueNotFoundError(name);
    }
    return queue;
  }

  public getQueueInfo(name: string): QueueInfo {
    const queue = this.getQueue(name);
    return {
      name: queue.name,
      maxSize: queue.maxSize,
      retentionTime: queue.retentionTime,
      currentSize: queue.availableQueue.length + this.getProcessingCount(queue),
      deadLetterSize: queue.deadLetterQueue.length
    };
  }

  public getAllQueuesInfo(): QueueInfo[] {
    return Array.from(this.queues.keys()).map(name => this.getQueueInfo(name));
  }

  public sendMessage(queueName: string, content: any): Message {
    const queue = this.getQueue(queueName);
    const currentSize = queue.availableQueue.length + this.getProcessingCount(queue);

    if (currentSize >= queue.maxSize) {
      throw new QueueFullError(queueName, currentSize, queue.maxSize);
    }

    const now = Date.now();
    const message: Message = {
      id: uuidv4(),
      status: 'available',
      content,
      createdAt: now,
      updatedAt: now,
      retryCount: 0
    };

    queue.messages.set(message.id, message);
    queue.availableQueue.push(message.id);

    return message;
  }

  public consumeMessages(
    queueName: string,
    consumerId: string,
    batchSize: number = 1
  ): Message[] {
    const queue = this.getQueue(queueName);
    const size = Math.min(Math.max(1, batchSize), MAX_BATCH_SIZE);
    const result: Message[] = [];
    const now = Date.now();

    while (result.length < size && queue.availableQueue.length > 0) {
      const messageId = queue.availableQueue.shift()!;
      const message = queue.messages.get(messageId)!;

      if (this.isMessageExpired(message, queue.retentionTime)) {
        queue.messages.delete(messageId);
        continue;
      }

      message.status = 'processing';
      message.consumerId = consumerId;
      message.consumedAt = now;
      message.updatedAt = now;
      result.push(message);
    }

    return result;
  }

  public acknowledgeMessage(
    queueName: string,
    messageId: string,
    consumerId: string
  ): Message {
    const queue = this.queues.get(queueName);
    if (!queue) {
      return this.handleOrphanMessage(messageId);
    }

    const message = queue.messages.get(messageId);
    if (!message) {
      throw new MessageNotFoundError(messageId);
    }

    if (message.consumerId !== consumerId) {
      throw new NotOwnerError(messageId, consumerId);
    }

    if (message.status !== 'processing') {
      throw new InvalidStateTransitionError(message.status, 'acknowledged');
    }

    message.status = 'acknowledged';
    message.updatedAt = Date.now();
    queue.messages.delete(messageId);

    return message;
  }

  public retryOrDeadLetter(
    queueName: string,
    messageId: string,
    consumerId: string
  ): Message {
    const queue = this.queues.get(queueName);
    if (!queue) {
      return this.handleOrphanMessage(messageId);
    }

    const message = queue.messages.get(messageId);
    if (!message) {
      throw new MessageNotFoundError(messageId);
    }

    if (message.consumerId !== consumerId) {
      throw new NotOwnerError(messageId, consumerId);
    }

    if (message.status !== 'processing') {
      throw new InvalidStateTransitionError(message.status, message.retryCount >= MAX_RETRY ? 'dead-letter' : 'available');
    }

    message.retryCount++;
    message.updatedAt = Date.now();

    if (message.retryCount > MAX_RETRY) {
      message.status = 'dead-letter';
      queue.deadLetterQueue.push(messageId);
    } else {
      message.status = 'available';
      message.consumerId = undefined;
      message.consumedAt = undefined;
      queue.availableQueue.unshift(messageId);
    }

    return message;
  }

  public getDeadLetterMessages(queueName: string): Message[] {
    const queue = this.getQueue(queueName);
    return queue.deadLetterQueue
      .map(id => queue.messages.get(id)!)
      .filter(msg => msg !== undefined);
  }

  public processTimeouts(): void {
    const now = Date.now();

    for (const queue of this.queues.values()) {
      for (const [messageId, message] of queue.messages) {
        if (message.status === 'processing' && message.consumedAt) {
          if (now - message.consumedAt > ACK_TIMEOUT) {
            message.retryCount++;
            message.updatedAt = now;

            if (message.retryCount > MAX_RETRY) {
              message.status = 'dead-letter';
              queue.deadLetterQueue.push(messageId);
            } else {
              message.status = 'available';
              message.consumerId = undefined;
              message.consumedAt = undefined;
              queue.availableQueue.unshift(messageId);
            }
          }
        }
      }
    }
  }

  private handleOrphanMessage(messageId: string): Message {
    return {
      id: messageId,
      status: 'dead-letter',
      content: null,
      createdAt: Date.now(),
      updatedAt: Date.now(),
      retryCount: 0
    };
  }

  private getProcessingCount(queue: Queue): number {
    let count = 0;
    for (const message of queue.messages.values()) {
      if (message.status === 'processing') {
        count++;
      }
    }
    return count;
  }

  private isMessageExpired(message: Message, retentionTime: number): boolean {
    return Date.now() - message.createdAt > retentionTime;
  }
}
