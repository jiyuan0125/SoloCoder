import { Template, Channel, Notification, AlertRecord, ChannelType } from '../types';

class DataStore {
  private templates: Map<string, Template> = new Map();
  private channels: Map<string, Channel> = new Map();
  private notifications: Map<string, Notification> = new Map();
  private sendRecords: Map<string, Map<ChannelType, number[]>> = new Map();
  private channelFailures: Map<string, { time: number; success: boolean }[]> = new Map();
  private channelAlerts: Map<string, AlertRecord> = new Map();

  getTemplates(): Template[] {
    return Array.from(this.templates.values());
  }

  getTemplate(id: string): Template | undefined {
    return this.templates.get(id);
  }

  addTemplate(template: Template): void {
    this.templates.set(template.id, template);
  }

  updateTemplate(id: string, template: Template): void {
    this.templates.set(id, template);
  }

  deleteTemplate(id: string): void {
    this.templates.delete(id);
  }

  getChannels(): Channel[] {
    return Array.from(this.channels.values());
  }

  getChannel(id: string): Channel | undefined {
    return this.channels.get(id);
  }

  addChannel(channel: Channel): void {
    this.channels.set(channel.id, channel);
    this.channelFailures.set(channel.id, []);
  }

  updateChannel(id: string, channel: Channel): void {
    this.channels.set(id, channel);
  }

  deleteChannel(id: string): void {
    this.channels.delete(id);
    this.channelFailures.delete(id);
    this.channelAlerts.delete(id);
  }

  getNotifications(): Notification[] {
    return Array.from(this.notifications.values());
  }

  getPendingNotifications(): Notification[] {
    return Array.from(this.notifications.values()).filter(
      n => n.status === 'pending' || n.status === 'delayed'
    );
  }

  getNotification(id: string): Notification | undefined {
    return this.notifications.get(id);
  }

  addNotification(notification: Notification): void {
    this.notifications.set(notification.id, notification);
  }

  updateNotification(id: string, notification: Notification): void {
    this.notifications.set(id, notification);
  }

  getSendRecords(userId: string, channelType: ChannelType): number[] {
    const userRecords = this.sendRecords.get(userId);
    if (!userRecords) return [];
    return userRecords.get(channelType) || [];
  }

  addSendRecord(userId: string, channelType: ChannelType, timestamp: number): void {
    let userRecords = this.sendRecords.get(userId);
    if (!userRecords) {
      userRecords = new Map();
      this.sendRecords.set(userId, userRecords);
    }
    const records = userRecords.get(channelType) || [];
    records.push(timestamp);
    userRecords.set(channelType, records);
  }

  addChannelRecord(channelId: string, success: boolean): void {
    const records = this.channelFailures.get(channelId) || [];
    records.push({ time: Date.now(), success });
    if (records.length > 1000) {
      records.shift();
    }
    this.channelFailures.set(channelId, records);
  }

  getChannelRecords(channelId: string): { time: number; success: boolean }[] {
    return this.channelFailures.get(channelId) || [];
  }

  getAlert(channelId: string): AlertRecord | undefined {
    return this.channelAlerts.get(channelId);
  }

  setAlert(alert: AlertRecord): void {
    this.channelAlerts.set(alert.channelId, alert);
  }

  clearAlert(channelId: string): void {
    this.channelAlerts.delete(channelId);
  }
}

export const store = new DataStore();
