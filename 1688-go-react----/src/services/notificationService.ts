import { db } from '../database';
import { Alert, AssessmentResult, Scale } from '../types';
import { v4 as uuidv4 } from 'uuid';
import { addNotifiedUser } from './alertService';

export const sendAlertNotification = async (
  alert: Alert,
  result: AssessmentResult,
  scale: Scale,
  recipientId: string
): Promise<void> => {
  const levelText: { [key: string]: string } = {
    moderate: '中度异常',
    severe: '重度异常'
  };

  const priorityText: { [key: string]: string } = {
    normal: '普通',
    urgent: '紧急（需48小时内介入）'
  };

  const content = `【心理测评预警通知】\n\n` +
    `用户ID：${result.userId}\n` +
    `测评量表：${scale.name}\n` +
    `测评结果等级：${levelText[alert.level]}\n` +
    `预警优先级：${priorityText[alert.priority]}\n` +
    `原始分数：${result.rawScore}\n` +
    `标准分数：${result.standardScore}\n` +
    `预警时间：${new Date(alert.createdAt).toLocaleString()}\n\n` +
    `请及时关注并处理该预警。`;

  const notification = {
    id: uuidv4(),
    alertId: alert.id,
    userId: result.userId,
    recipientId,
    content,
    sentAt: Date.now()
  };

  return new Promise((resolve, reject) => {
    db.run(
      `INSERT INTO notifications (id, alert_id, user_id, recipient_id, content, sent_at)
       VALUES (?, ?, ?, ?, ?, ?)`,
      [
        notification.id,
        notification.alertId,
        notification.userId,
        notification.recipientId,
        notification.content,
        notification.sentAt
      ],
      async (err) => {
        if (err) {
          reject(err);
          return;
        }

        try {
          await addNotifiedUser(alert.id, recipientId);
          console.log(`[Notification] Sent to ${recipientId}: ${content}`);
          resolve();
        } catch (addErr) {
          reject(addErr);
        }
      }
    );
  });
};

export const getNotificationsByRecipient = (recipientId: string): Promise<any[]> => {
  return new Promise((resolve, reject) => {
    db.all(
      'SELECT * FROM notifications WHERE recipient_id = ? ORDER BY sent_at DESC',
      [recipientId],
      (err, rows: any[]) => {
        if (err) reject(err);
        else resolve(rows.map(row => ({
          id: row.id,
          alertId: row.alert_id,
          userId: row.user_id,
          recipientId: row.recipient_id,
          content: row.content,
          sentAt: row.sent_at,
          readAt: row.read_at
        })));
      }
    );
  });
};

export const markNotificationRead = (notificationId: string): Promise<void> => {
  return new Promise((resolve, reject) => {
    db.run(
      'UPDATE notifications SET read_at = ? WHERE id = ?',
      [Date.now(), notificationId],
      (err) => {
        if (err) reject(err);
        else resolve();
      }
    );
  });
};
