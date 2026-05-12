import { Channel, ChannelType, ChannelInstance } from '../types';
import { v4 as uuidv4 } from 'uuid';
import { store } from '../storage/store';
import { alertService } from './AlertService';

export interface CreateChannelInput {
  name: string;
  type: ChannelType;
  primary: ChannelInstance;
  backup?: ChannelInstance;
  rateLimits?: {
    perMinute: number;
    perHour: number;
    perDay: number;
  };
}

export class ChannelService {
  create(input: CreateChannelInput): Channel {
    const now = Date.now();
    const channel: Channel = {
      id: uuidv4(),
      name: input.name,
      type: input.type,
      primary: input.primary,
      backup: input.backup,
      activeInstance: 'primary',
      rateLimits: input.rateLimits ?? {
        perMinute: 1,
        perHour: 5,
        perDay: 20,
      },
      status: 'active',
      createdAt: now,
      updatedAt: now,
    };
    store.addChannel(channel);
    return channel;
  }

  list(): Channel[] {
    return store.getChannels();
  }

  get(id: string): Channel | undefined {
    return store.getChannel(id);
  }

  switchToBackup(channelId: string): Channel | undefined {
    const channel = store.getChannel(channelId);
    if (!channel || !channel.backup) return undefined;
    if (channel.activeInstance === 'backup') return channel;

    const updated: Channel = {
      ...channel,
      activeInstance: 'backup',
      status: 'alerting',
      updatedAt: Date.now(),
    };
    store.updateChannel(channelId, updated);
    return updated;
  }

  switchToPrimary(channelId: string): Channel | undefined {
    const channel = store.getChannel(channelId);
    if (!channel) return undefined;
    if (channel.activeInstance === 'primary') return channel;

    const updated: Channel = {
      ...channel,
      activeInstance: 'primary',
      status: 'active',
      updatedAt: Date.now(),
    };
    store.updateChannel(channelId, updated);
    store.clearAlert(channelId);
    return updated;
  }

  updateConfig(
    id: string,
    updates: {
      name?: string;
      primary?: ChannelInstance;
      backup?: ChannelInstance;
      rateLimits?: { perMinute: number; perHour: number; perDay: number };
    }
  ): Channel | undefined {
    const existing = store.getChannel(id);
    if (!existing) return undefined;

    if (updates.rateLimits) {
      const { perMinute, perHour, perDay } = updates.rateLimits;
      if (perMinute <= 0 || perHour <= 0 || perDay <= 0) {
        return undefined;
      }
    }

    const updated: Channel = {
      ...existing,
      name: updates.name ?? existing.name,
      primary: updates.primary ?? existing.primary,
      backup: updates.backup !== undefined ? updates.backup : existing.backup,
      rateLimits: updates.rateLimits ?? existing.rateLimits,
      updatedAt: Date.now(),
    };
    store.updateChannel(id, updated);
    return updated;
  }

  recordResult(channelId: string, success: boolean): void {
    store.addChannelRecord(channelId, success);
    if (alertService.checkAndTrigger(channelId)) {
      this.switchToBackup(channelId);
    }
  }

  hasPendingNotifications(channelId: string): boolean {
    const pending = store.getPendingNotifications();
    return pending.some(n => n.channelId === channelId);
  }

  delete(id: string): boolean {
    if (!store.getChannel(id)) return false;
    store.deleteChannel(id);
    return true;
  }
}

export const channelService = new ChannelService();
