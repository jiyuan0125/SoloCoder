import { run, get } from '../database';
import { ChannelConfig, ChannelType } from '../types';

export const getChannelConfig = async (type: ChannelType): Promise<ChannelConfig | null> => {
  const row = await get<any>(`SELECT * FROM channel_configs WHERE type = ?`, [type]);
  if (!row) return null;

  return {
    type: row.type as ChannelType,
    enabled: Boolean(row.enabled),
    rateLimitPerMinute: row.rateLimitPerMinute,
    emailRecipients: row.emailRecipients ? JSON.parse(row.emailRecipients) : undefined,
    smsNumbers: row.smsNumbers ? JSON.parse(row.smsNumbers) : undefined,
    webhookUrl: row.webhookUrl || undefined
  };
};

export const updateChannelConfig = async (config: ChannelConfig): Promise<ChannelConfig> => {
  const existing = await get<any>(`SELECT * FROM channel_configs WHERE type = ?`, [config.type]);

  if (existing) {
    await run(
      `UPDATE channel_configs 
       SET enabled = ?, rateLimitPerMinute = ?, emailRecipients = ?, smsNumbers = ?, webhookUrl = ?
       WHERE type = ?`,
      [
        config.enabled ? 1 : 0,
        config.rateLimitPerMinute,
        config.emailRecipients ? JSON.stringify(config.emailRecipients) : null,
        config.smsNumbers ? JSON.stringify(config.smsNumbers) : null,
        config.webhookUrl || null,
        config.type
      ]
    );
  } else {
    await run(
      `INSERT INTO channel_configs (type, enabled, rateLimitPerMinute, emailRecipients, smsNumbers, webhookUrl)
       VALUES (?, ?, ?, ?, ?, ?)`,
      [
        config.type,
        config.enabled ? 1 : 0,
        config.rateLimitPerMinute,
        config.emailRecipients ? JSON.stringify(config.emailRecipients) : null,
        config.smsNumbers ? JSON.stringify(config.smsNumbers) : null,
        config.webhookUrl || null
      ]
    );
  }

  return config;
};

export const getAllChannelConfigs = async (): Promise<ChannelConfig[]> => {
  const { all } = await import('../database');
  const rows = await all<any>(`SELECT * FROM channel_configs`);

  return rows.map(row => ({
    type: row.type as ChannelType,
    enabled: Boolean(row.enabled),
    rateLimitPerMinute: row.rateLimitPerMinute,
    emailRecipients: row.emailRecipients ? JSON.parse(row.emailRecipients) : undefined,
    smsNumbers: row.smsNumbers ? JSON.parse(row.smsNumbers) : undefined,
    webhookUrl: row.webhookUrl || undefined
  }));
};
