import { v4 as uuidv4 } from 'uuid';
import { run, all, get } from '../database';
import { ChannelType, NotificationQueue } from '../types';
import { getChannelConfig } from './channelService';

const channelSentCounts: Record<ChannelType, { count: number; windowStart: number }> = {
  email: { count: 0, windowStart: 0 },
  sms: { count: 0, windowStart: 0 },
  webhook: { count: 0, windowStart: 0 }
};

const resetWindowIfNeeded = (channelType: ChannelType) => {
  const now = Date.now();
  const window = channelSentCounts[channelType];
  if (now - window.windowStart >= 60000) {
    window.count = 0;
    window.windowStart = now;
  }
};

const sendEmail = async (content: string, config: any): Promise<boolean> => {
  console.log(`[邮件] 发送到: ${config.emailRecipients?.join(', ') || '未配置收件人'}`);
  console.log(`[邮件] 内容: ${content.substring(0, 200)}...`);
  return true;
};

const sendSms = async (content: string, config: any): Promise<boolean> => {
  console.log(`[短信] 发送到: ${config.smsNumbers?.join(', ') || '未配置号码'}`);
  console.log(`[短信] 内容: ${content.substring(0, 160)}...`);
  return true;
};

const sendWebhook = async (content: string, config: any): Promise<boolean> => {
  console.log(`[Webhook] 发送到: ${config.webhookUrl || '未配置URL'}`);
  console.log(`[Webhook] 内容: ${content.substring(0, 500)}...`);
  return true;
};

const sendInternal = async (channelType: ChannelType, content: string): Promise<boolean> => {
  const config = await getChannelConfig(channelType);
  if (!config || !config.enabled) {
    return false;
  }

  resetWindowIfNeeded(channelType);
  const window = channelSentCounts[channelType];

  if (window.count >= config.rateLimitPerMinute) {
    return false;
  }

  try {
    let success = false;
    switch (channelType) {
      case 'email':
        success = await sendEmail(content, config);
        break;
      case 'sms':
        success = await sendSms(content, config);
        break;
      case 'webhook':
        success = await sendWebhook(content, config);
        break;
    }

    if (success) {
      window.count++;
    }

    return success;
  } catch (err) {
    console.error(`[通知发送失败] ${channelType}:`, err);
    return false;
  }
};

export const queueNotification = async (
  channelType: ChannelType,
  content: string,
  priority: number = 0,
  alertId?: string
): Promise<void> => {
  const now = Date.now();
  const id = uuidv4();

  await run(
    `INSERT INTO notification_queue (id, channelType, content, priority, scheduledTime, createdAt, sent, alertId)
     VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
    [id, channelType, content, priority, now, now, 0, alertId || null]
  );
};

export const processNotificationQueue = async (): Promise<void> => {
  const now = Date.now();

  const pending = await all<any>(
    `SELECT * FROM notification_queue 
     WHERE sent = 0 AND scheduledTime <= ?
     ORDER BY priority DESC, scheduledTime ASC`,
    [now]
  );

  for (const notification of pending) {
    const sent = await sendInternal(notification.channelType, notification.content);

    if (sent) {
      await run(
        `UPDATE notification_queue SET sent = 1 WHERE id = ?`,
        [notification.id]
      );
    } else {
      const config = await getChannelConfig(notification.channelType as ChannelType);
      if (config && config.enabled) {
        const nextMinute = now + 60000;
        await run(
          `UPDATE notification_queue SET scheduledTime = ? WHERE id = ?`,
          [nextMinute, notification.id]
        );
      }
    }
  }
};

export const buildAggregatedAlertMessage = (agg: any, isEscalation: boolean = false): string => {
  const levelMap: Record<string, string> = {
    info: '信息',
    warning: '警告',
    critical: '严重',
    emergency: '紧急'
  };
  const levelLabel = levelMap[agg.level] || '未知';

  const prefix = isEscalation ? '[告警升级] ' : '[告警通知] ';
  const lines = [
    `${prefix}${levelLabel} - ${agg.name}`,
    `来源系统: ${agg.sourceSystem}`,
    `告警次数: ${agg.count}`,
    `首次时间: ${new Date(agg.createdAt).toISOString()}`,
    '',
    '告警描述:'
  ];

  agg.descriptions.slice(0, 5).forEach((desc: string, i: number) => {
    lines.push(`  ${i + 1}. ${desc}`);
  });

  if (agg.descriptions.length > 5) {
    lines.push(`  ... (还有 ${agg.descriptions.length - 5} 条)`);
  }

  return lines.join('\n');
};

export const sendAggregatedAlert = async (agg: any, isEscalation: boolean = false): Promise<void> => {
  const content = buildAggregatedAlertMessage(agg, isEscalation);
  const priority = agg.level === 'emergency' ? 3 : agg.level === 'critical' ? 2 : agg.level === 'warning' ? 1 : 0;

  const channels: ChannelType[] = ['email', 'sms', 'webhook'];
  for (const channel of channels) {
    await queueNotification(channel, content, priority);
  }
};
