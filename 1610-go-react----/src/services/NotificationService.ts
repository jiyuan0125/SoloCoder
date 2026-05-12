import { Notification, Channel } from '../types';
import { v4 as uuidv4 } from 'uuid';
import { store } from '../storage/store';
import { templateService } from './TemplateService';
import { channelService } from './ChannelService';
import { rateLimiter } from './RateLimiter';

export interface SendNotificationInput {
  channelId: string;
  templateId: string;
  userId: string;
  variables: Record<string, any>;
}

export interface SendResult {
  success: boolean;
  notificationId?: string;
  missingVariables?: string[];
  waitSeconds?: number;
  renderedContent?: string;
}

class NotificationSender {
  async send(channel: Channel, content: string): Promise<boolean> {
    const instance = channel.activeInstance === 'primary' ? channel.primary : channel.backup;
    if (!instance) return false;
    return true;
  }
}

const sender = new NotificationSender();

export class NotificationService {
  private delayedQueueTimer: NodeJS.Timeout | null = null;
  private isProcessing = false;

  create(input: SendNotificationInput): Notification {
    const now = Date.now();
    const notification: Notification = {
      id: uuidv4(),
      channelId: input.channelId,
      templateId: input.templateId,
      userId: input.userId,
      variables: input.variables,
      status: 'pending',
      scheduledAt: now,
      retries: 0,
      createdAt: now,
      updatedAt: now,
    };
    store.addNotification(notification);
    return notification;
  }

  async send(input: SendNotificationInput): Promise<SendResult> {
    const channel = channelService.get(input.channelId);
    if (!channel) {
      return { success: false };
    }

    const template = templateService.get(input.templateId);
    if (!template) {
      return { success: false };
    }

    const missing = templateService.validateVariables(template, input.variables);
    if (missing.length > 0) {
      return { success: false, missingVariables: missing };
    }

    const rateCheck = rateLimiter.check(input.userId, channel.type, channel.rateLimits);
    if (!rateCheck.allowed) {
      const notification = this.create(input);
      this.scheduleForDelay(notification.id, rateCheck.waitSeconds!);
      return { success: false, waitSeconds: rateCheck.waitSeconds, notificationId: notification.id };
    }

    const notification = this.create(input);
    const renderedContent = templateService.render(template, input.variables);

    const result = await this.trySendToChannel(channel, input.userId, renderedContent);
    this.updateNotificationStatus(notification.id, result.success ? 'sent' : 'failed');

    if (result.success) {
      return { success: true, notificationId: notification.id, renderedContent };
    }

    if (result.shouldFallback && channel.backup && channel.activeInstance === 'primary') {
      const updated = channelService.switchToBackup(channel.id);
      if (updated) {
        const fallbackResult = await this.trySendToChannel(updated, input.userId, renderedContent);
        this.updateNotificationStatus(notification.id, fallbackResult.success ? 'sent' : 'failed');
        if (fallbackResult.success) {
          return { success: true, notificationId: notification.id, renderedContent };
        }
      }
    }

    return { success: false, notificationId: notification.id };
  }

  private async trySendToChannel(
    channel: Channel,
    userId: string,
    content: string
  ): Promise<{ success: boolean; shouldFallback: boolean }> {
    try {
      const result = await sender.send(channel, content);
      if (result) {
        rateLimiter.record(userId, channel.type);
        channelService.recordResult(channel.id, true);
        return { success: true, shouldFallback: false };
      } else {
        channelService.recordResult(channel.id, false);
        return { success: false, shouldFallback: true };
      }
    } catch {
      channelService.recordResult(channel.id, false);
      return { success: false, shouldFallback: true };
    }
  }

  private scheduleForDelay(notificationId: string, waitSeconds: number): void {
    const notification = store.getNotification(notificationId);
    if (!notification) return;

    const updated: Notification = {
      ...notification,
      status: 'delayed',
      scheduledAt: Date.now() + waitSeconds * 1000,
      retries: notification.retries + 1,
      updatedAt: Date.now(),
    };
    store.updateNotification(notificationId, updated);
    this.startDelayQueue();
  }

  private startDelayQueue(): void {
    if (this.delayedQueueTimer) return;
    this.delayedQueueTimer = setInterval(() => {
      this.processDelayedQueue();
    }, 1000);
  }

  private async processDelayedQueue(): Promise<void> {
    if (this.isProcessing) return;
    this.isProcessing = true;

    try {
      const now = Date.now();
      const delayed = store.getNotifications().filter(
        n => n.status === 'delayed' && n.scheduledAt <= now
      );

      for (const notification of delayed) {
        const channel = channelService.get(notification.channelId);
        const template = templateService.get(notification.templateId);
        if (!channel || !template) {
          this.updateNotificationStatus(notification.id, 'cancelled');
          continue;
        }

        const rateCheck = rateLimiter.check(notification.userId, channel.type, channel.rateLimits);
        if (!rateCheck.allowed) {
          const updated: Notification = {
            ...notification,
            scheduledAt: Date.now() + rateCheck.waitSeconds! * 1000,
            retries: notification.retries + 1,
            updatedAt: Date.now(),
          };
          store.updateNotification(notification.id, updated);
          continue;
        }

        const renderedContent = templateService.render(template, notification.variables);
        const result = await this.trySendToChannel(channel, notification.userId, renderedContent);
        this.updateNotificationStatus(notification.id, result.success ? 'sent' : 'failed');
      }
    } finally {
      this.isProcessing = false;
    }
  }

  private updateNotificationStatus(id: string, status: Notification['status']): void {
    const existing = store.getNotification(id);
    if (!existing) return;
    const updated: Notification = {
      ...existing,
      status,
      updatedAt: Date.now(),
    };
    store.updateNotification(id, updated);
  }

  list(): Notification[] {
    return store.getNotifications();
  }

  get(id: string): Notification | undefined {
    return store.getNotification(id);
  }
}

export const notificationService = new NotificationService();
