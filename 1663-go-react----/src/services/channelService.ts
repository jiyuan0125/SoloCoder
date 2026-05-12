import { v4 as uuidv4 } from 'uuid';
import db from '../database';
import { Channel, CreateChannelRequest, UpdateChannelRequest } from '../types';

export class ChannelService {
  createChannel(data: CreateChannelRequest): Channel {
    // 验证佣金比例
    if (data.commission_rate < 0 || data.commission_rate > 1) {
      throw new Error('佣金比例必须在 0 到 1 之间');
    }

    // 检查名称重复
    const existingChannel = db.prepare('SELECT id FROM channels WHERE name = ?').get(data.name);
    if (existingChannel) {
      throw new Error('渠道名称已存在');
    }

    const id = uuidv4();
    const channel: Channel = {
      id,
      name: data.name,
      contact: data.contact,
      settlement_method: data.settlement_method,
      commission_rate: data.commission_rate,
      status: data.status || 'active',
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    };

    const stmt = db.prepare(`
      INSERT INTO channels (id, name, contact, settlement_method, commission_rate, status, created_at, updated_at)
      VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `);

    stmt.run(
      channel.id,
      channel.name,
      channel.contact,
      channel.settlement_method,
      channel.commission_rate,
      channel.status,
      channel.created_at,
      channel.updated_at
    );

    return channel;
  }

  getChannelById(id: string): Channel | undefined {
    return db.prepare('SELECT * FROM channels WHERE id = ?').get(id) as Channel | undefined;
  }

  getAllChannels(): Channel[] {
    return db.prepare('SELECT * FROM channels ORDER BY created_at DESC').all() as Channel[];
  }

  updateChannel(id: string, data: UpdateChannelRequest): Channel {
    const existingChannel = this.getChannelById(id);
    if (!existingChannel) {
      throw new Error('渠道不存在');
    }

    // 验证佣金比例
    if (data.commission_rate !== undefined && (data.commission_rate < 0 || data.commission_rate > 1)) {
      throw new Error('佣金比例必须在 0 到 1 之间');
    }

    // 检查名称重复
    if (data.name && data.name !== existingChannel.name) {
      const duplicateChannel = db.prepare('SELECT id FROM channels WHERE name = ? AND id != ?').get(data.name, id);
      if (duplicateChannel) {
        throw new Error('渠道名称已存在');
      }
    }

    const updatedChannel: Channel = {
      ...existingChannel,
      ...data,
      updated_at: new Date().toISOString(),
    };

    const stmt = db.prepare(`
      UPDATE channels
      SET name = ?, contact = ?, settlement_method = ?, commission_rate = ?, status = ?, updated_at = ?
      WHERE id = ?
    `);

    stmt.run(
      updatedChannel.name,
      updatedChannel.contact,
      updatedChannel.settlement_method,
      updatedChannel.commission_rate,
      updatedChannel.status,
      updatedChannel.updated_at,
      id
    );

    return updatedChannel;
  }

  deleteChannel(id: string): void {
    const existingChannel = this.getChannelById(id);
    if (!existingChannel) {
      throw new Error('渠道不存在');
    }

    db.prepare('DELETE FROM channels WHERE id = ?').run(id);
  }
}

export const channelService = new ChannelService();
