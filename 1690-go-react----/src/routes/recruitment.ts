import { Router, Request, Response } from 'express';
import db from '../database';
import { generateId, getNow, calculateAge } from '../utils';
import { sendNotification } from '../services/notification';
import { Volunteer, Registration, Activity } from '../types';

const router = Router();

router.post('/activities', (req: Request, res: Response) => {
  const { name, category, date_time, location, max_volunteers, registration_deadline } = req.body;

  if (!max_volunteers || max_volunteers <= 0) {
    return res.status(400).json({ error: '招募人数必须为正整数' });
  }

  const existing = db.prepare('SELECT id FROM activities WHERE name = ?').get(name);
  if (existing) {
    return res.status(409).json({ error: '活动名称已存在' });
  }

  const id = generateId();
  const now = getNow();

  db.prepare(`
    INSERT INTO activities (id, name, category, date_time, location, max_volunteers, registration_deadline, created_at, updated_at)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
  `).run(id, name, category, date_time, location, max_volunteers, registration_deadline, now, now);

  res.status(201).json({ id, name, category, date_time, location, max_volunteers, registration_deadline });
});

router.get('/activities', (req: Request, res: Response) => {
  const activities = db.prepare('SELECT * FROM activities').all();
  res.json(activities);
});

router.get('/activities/:id', (req: Request, res: Response) => {
  const activity = db.prepare('SELECT * FROM activities WHERE id = ?').get(req.params.id);
  if (!activity) {
    return res.status(404).json({ error: '活动不存在' });
  }
  res.json(activity);
});

router.put('/activities/:id', (req: Request, res: Response) => {
  const { name, category, date_time, location, max_volunteers, registration_deadline } = req.body;
  const activityId = req.params.id;

  const existing = db.prepare('SELECT * FROM activities WHERE id = ?').get(activityId);
  if (!existing) {
    return res.status(404).json({ error: '活动不存在' });
  }

  if (name) {
    const nameConflict = db.prepare('SELECT id FROM activities WHERE name = ? AND id != ?').get(name, activityId);
    if (nameConflict) {
      return res.status(409).json({ error: '活动名称已存在' });
    }
  }

  if (max_volunteers !== undefined && max_volunteers <= 0) {
    return res.status(400).json({ error: '招募人数必须为正整数' });
  }

  const now = getNow();
  db.prepare(`
    UPDATE activities 
    SET name = COALESCE(?, name),
        category = COALESCE(?, category),
        date_time = COALESCE(?, date_time),
        location = COALESCE(?, location),
        max_volunteers = COALESCE(?, max_volunteers),
        registration_deadline = COALESCE(?, registration_deadline),
        updated_at = ?
    WHERE id = ?
  `).run(name, category, date_time, location, max_volunteers, registration_deadline, now, activityId);

  const updated = db.prepare('SELECT * FROM activities WHERE id = ?').get(activityId);
  res.json(updated);
});

router.post('/volunteers', (req: Request, res: Response) => {
  const { name, phone, id_card, emergency_contact, guardian_name, guardian_phone } = req.body;

  const existing = db.prepare('SELECT id FROM volunteers WHERE phone = ?').get(phone);
  if (existing) {
    return res.status(409).json({ error: '手机号已存在' });
  }

  let age: number;
  try {
    age = calculateAge(id_card);
  } catch (e: any) {
    return res.status(400).json({ error: e.message });
  }

  if (age < 18 && (!guardian_name || !guardian_phone)) {
    return res.status(400).json({ error: '未满18周岁需提供监护人信息' });
  }

  const id = generateId();
  const now = getNow();

  db.prepare(`
    INSERT INTO volunteers (id, name, phone, id_card, age, guardian_name, guardian_phone, emergency_contact, created_at)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
  `).run(id, name, phone, id_card, age, age < 18 ? guardian_name : null, age < 18 ? guardian_phone : null, emergency_contact, now);

  res.status(201).json({ id, name, phone, age, emergency_contact });
});

router.get('/volunteers', (req: Request, res: Response) => {
  const volunteers = db.prepare('SELECT * FROM volunteers').all();
  res.json(volunteers);
});

router.post('/registrations', (req: Request, res: Response) => {
  const { activity_id, volunteer_id } = req.body;

  const activity = db.prepare('SELECT * FROM activities WHERE id = ?').get(activity_id);
  if (!activity) {
    return res.status(404).json({ error: '活动不存在' });
  }

  const volunteer = db.prepare('SELECT * FROM volunteers WHERE id = ?').get(volunteer_id) as Volunteer | undefined;
  if (!volunteer) {
    return res.status(404).json({ error: '志愿者不存在' });
  }

  if (!volunteer.entry_training_completed) {
    return res.status(403).json({ error: '未完成入门培训，不能报名活动' });
  }

  if (volunteer.last_annual_training_date) {
    const lastTraining = new Date(volunteer.last_annual_training_date);
    const now = new Date();
    const diffDays = Math.floor((now.getTime() - lastTraining.getTime()) / (1000 * 60 * 60 * 24));
    if (diffDays > 365) {
      return res.status(403).json({ error: '未完成年度复训，不能报名新活动' });
    }
  }

  const existing = db.prepare('SELECT id FROM registrations WHERE activity_id = ? AND volunteer_id = ?').get(activity_id, volunteer_id);
  if (existing) {
    return res.status(409).json({ error: '同一活动同一志愿者只能报名一次' });
  }

  const id = generateId();
  const now = getNow();

  db.prepare(`
    INSERT INTO registrations (id, activity_id, volunteer_id, status, registered_at)
    VALUES (?, ?, ?, 'pending', ?)
  `).run(id, activity_id, volunteer_id, now);

  res.status(201).json({ id, activity_id, volunteer_id, status: 'pending' });
});

router.get('/registrations', (req: Request, res: Response) => {
  const registrations = db.prepare(`
    SELECT r.*, a.name as activity_name, v.name as volunteer_name
    FROM registrations r
    JOIN activities a ON r.activity_id = a.id
    JOIN volunteers v ON r.volunteer_id = v.id
  `).all();
  res.json(registrations);
});

router.post('/registrations/:id/review', (req: Request, res: Response) => {
  const { action, hours, is_special_position } = req.body;
  const registrationId = req.params.id;

  const registration = db.prepare('SELECT * FROM registrations WHERE id = ?').get(registrationId) as Registration | undefined;
  if (!registration) {
    return res.status(404).json({ error: '报名记录不存在' });
  }

  const activity = db.prepare('SELECT * FROM activities WHERE id = ?').get(registration.activity_id) as Activity | undefined;
  if (!activity) {
    return res.status(404).json({ error: '活动不存在' });
  }

  if (action !== 'approve' && action !== 'reject') {
    return res.status(400).json({ error: 'action 必须是 approve 或 reject' });
  }

  const now = getNow();

  if (action === 'approve') {
    db.prepare(`
      UPDATE registrations 
      SET status = 'approved', reviewed_at = ?, hours = ?, is_special_position = ?
      WHERE id = ?
    `).run(now, hours || 0, is_special_position ? 1 : 0, registrationId);

    if (hours && hours > 0) {
      const pointsPerHour = 10;
      const specialBonus = is_special_position ? 5 : 0;
      const totalPoints = hours * (pointsPerHour + specialBonus);
      const pointId = generateId();
      db.prepare(`
        INSERT INTO points (id, volunteer_id, amount, description, source, created_at)
        VALUES (?, ?, ?, ?, 'activity', ?)
      `).run(pointId, registration.volunteer_id, totalPoints, `${activity.name} 服务积分`, now);
    }

    const notificationSent = sendNotification(registration.volunteer_id, registration.activity_id);

    if (!notificationSent) {
      return res.status(200).json({ 
        message: '审核已通过但通知发送失败',
        registration: db.prepare('SELECT * FROM registrations WHERE id = ?').get(registrationId)
      });
    }

    return res.status(200).json({ 
      message: '审核已通过，通知已发送',
      registration: db.prepare('SELECT * FROM registrations WHERE id = ?').get(registrationId)
    });
  } else {
    db.prepare(`
      UPDATE registrations 
      SET status = 'rejected', reviewed_at = ?
      WHERE id = ?
    `).run(now, registrationId);

    return res.status(200).json({ 
      message: '审核已拒绝',
      registration: db.prepare('SELECT * FROM registrations WHERE id = ?').get(registrationId)
    });
  }
});

export default router;
