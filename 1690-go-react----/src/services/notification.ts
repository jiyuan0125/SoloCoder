import db from '../database';
import { generateId, getNow } from '../utils';
import { Volunteer, Activity } from '../types';

export function sendNotification(volunteerId: string, activityId: string): boolean {
  const activity = db.prepare('SELECT * FROM activities WHERE id = ?').get(activityId) as Activity | undefined;
  if (!activity) {
    return false;
  }

  const volunteer = db.prepare('SELECT * FROM volunteers WHERE id = ?').get(volunteerId) as Volunteer | undefined;
  if (!volunteer) {
    return false;
  }

  const content = `尊敬的${volunteer.name}：\n\n您报名的活动「${activity.name}」已审核通过！\n\n活动详情：\n- 名称：${activity.name}\n- 类别：${activity.category}\n- 时间：${activity.date_time}\n- 地点：${activity.location}\n\n请准时前往集合地点参加活动。`;

  const id = generateId();
  const now = getNow();

  try {
    db.prepare(`
      INSERT INTO notifications (id, volunteer_id, activity_id, content, sent_at, status)
      VALUES (?, ?, ?, ?, ?, 'sent')
    `).run(id, volunteerId, activityId, content, now);

    console.log(`[通知] 已发送给志愿者 ${volunteer.name}: ${activity.name}`);
    return true;
  } catch (e) {
    console.error(`[通知] 发送失败:`, e);

    db.prepare(`
      INSERT INTO notifications (id, volunteer_id, activity_id, content, status)
      VALUES (?, ?, ?, ?, 'failed')
    `).run(id, volunteerId, activityId, content);

    return false;
  }
}

export function getNotifications(volunteerId: string): any[] {
  return db.prepare('SELECT * FROM notifications WHERE volunteer_id = ? ORDER BY sent_at DESC').all(volunteerId);
}

export function getAllNotifications(): any[] {
  return db.prepare('SELECT n.*, v.name as volunteer_name, a.name as activity_name FROM notifications n LEFT JOIN volunteers v ON n.volunteer_id = v.id LEFT JOIN activities a ON n.activity_id = a.id ORDER BY sent_at DESC').all();
}
