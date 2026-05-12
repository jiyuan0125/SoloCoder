import { Router, Request, Response } from 'express';
import db from '../database';
import { Activity, ActivityRegistration, Volunteer } from '../types';
import { isDatePassed, isWithin24Hours, parseDateTime, timeOverlap } from '../utils/time';

const router = Router();

router.post('/', (req: Request, res: Response) => {
  const { name, category, date, start_time, end_time, location, required_people, skills_required, created_by } = req.body;

  if (!name || !category || !date || !start_time || !end_time || !location || required_people === undefined || skills_required === undefined || !created_by) {
    return res.status(400).json({ error: '缺少必填字段' });
  }

  if (isDatePassed(date)) {
    return res.status(400).json({ error: '活动日期已过' });
  }

  if (required_people <= 0) {
    return res.status(400).json({ error: '所需人数必须大于0' });
  }

  const creator = db.prepare('SELECT * FROM volunteers WHERE id = ?').get(created_by) as Volunteer | undefined;
  if (!creator) {
    return res.status(404).json({ error: '创建者不存在' });
  }

  if (!creator.is_backbone) {
    return res.status(403).json({ error: '只有骨干志愿者才能创建活动' });
  }

  const stmt = db.prepare(`
    INSERT INTO activities (name, category, date, start_time, end_time, location, required_people, skills_required, created_by)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
  `);
  const result = stmt.run(name, category, date, start_time, end_time, location, required_people, JSON.stringify(skills_required), created_by);

  const activity = db.prepare('SELECT * FROM activities WHERE id = ?').get(result.lastInsertRowid) as Activity;
  res.status(201).json({
    ...activity,
    skills_required: JSON.parse(activity.skills_required)
  });
});

router.get('/', (req: Request, res: Response) => {
  const activities = db.prepare('SELECT * FROM activities').all() as Activity[];
  res.json(activities.map(a => ({
    ...a,
    skills_required: JSON.parse(a.skills_required)
  })));
});

router.get('/:id', (req: Request, res: Response) => {
  const activity = db.prepare('SELECT * FROM activities WHERE id = ?').get(req.params.id) as Activity | undefined;
  if (!activity) {
    return res.status(404).json({ error: '活动不存在' });
  }
  res.json({
    ...activity,
    skills_required: JSON.parse(activity.skills_required)
  });
});

router.post('/:id/register', (req: Request, res: Response) => {
  const { volunteer_id } = req.body;
  const activityId = parseInt(req.params.id);

  if (!volunteer_id) {
    return res.status(400).json({ error: '缺少志愿者ID' });
  }

  const activity = db.prepare('SELECT * FROM activities WHERE id = ?').get(activityId) as Activity | undefined;
  if (!activity) {
    return res.status(404).json({ error: '活动不存在' });
  }

  const volunteer = db.prepare('SELECT * FROM volunteers WHERE id = ?').get(volunteer_id) as Volunteer | undefined;
  if (!volunteer) {
    return res.status(404).json({ error: '志愿者不存在' });
  }

  const existing = db.prepare('SELECT id FROM activity_registrations WHERE volunteer_id = ? AND activity_id = ?').get(volunteer_id, activityId);
  if (existing) {
    return res.status(409).json({ error: '已报名该活动' });
  }

  const availability = JSON.parse(volunteer.availability);
  const activityStartTime = activity.start_time;
  const activityEndTime = activity.end_time;

  let hasConflict = false;
  let conflictReason = '';

  for (const slot of availability) {
    const slotStart = slot.start;
    const slotEnd = slot.end;
    if (timeOverlap(activityStartTime, activityEndTime, slotStart, slotEnd)) {
      hasConflict = true;
      conflictReason = `活动时间 (${activityStartTime}-${activityEndTime}) 与可服务时间段 (${slotStart}-${slotEnd}) 冲突`;
      break;
    }
  }

  const otherRegistrations = db.prepare(`
    SELECT a.date, a.start_time, a.end_time, a.id
    FROM activity_registrations ar
    JOIN activities a ON ar.activity_id = a.id
    WHERE ar.volunteer_id = ? AND ar.status = 'registered' AND a.id != ?
  `).all(volunteer_id, activityId) as Activity[];

  for (const other of otherRegistrations) {
    if (other.date === activity.date && timeOverlap(activityStartTime, activityEndTime, other.start_time, other.end_time)) {
      hasConflict = true;
      conflictReason = `与已报名的活动"${other.name}" (${other.date} ${other.start_time}-${other.end_time}) 时间冲突`;
      break;
    }
  }

  if (hasConflict) {
    return res.status(409).json({ error: conflictReason });
  }

  const stmt = db.prepare(`
    INSERT INTO activity_registrations (volunteer_id, activity_id, status)
    VALUES (?, ?, 'registered')
  `);
  const result = stmt.run(volunteer_id, activityId);

  const registration = db.prepare('SELECT * FROM activity_registrations WHERE id = ?').get(result.lastInsertRowid) as ActivityRegistration;
  res.status(201).json(registration);
});

router.delete('/:id/register', (req: Request, res: Response) => {
  const { volunteer_id } = req.body;
  const activityId = parseInt(req.params.id);

  if (!volunteer_id) {
    return res.status(400).json({ error: '缺少志愿者ID' });
  }

  const registration = db.prepare(`
    SELECT * FROM activity_registrations
    WHERE volunteer_id = ? AND activity_id = ?
  `).get(volunteer_id, activityId) as ActivityRegistration | undefined;

  if (!registration) {
    return res.status(404).json({ error: '报名记录不存在' });
  }

  if (registration.status === 'cancelled' || registration.status === 'completed') {
    return res.status(404).json({ error: '报名记录不存在' });
  }

  const activity = db.prepare('SELECT * FROM activities WHERE id = ?').get(activityId) as Activity;
  const activityStartTime = parseDateTime(activity.date, activity.start_time);

  if (isWithin24Hours(activityStartTime)) {
    return res.status(400).json({ error: '活动开始前24小时内不能取消报名' });
  }

  db.prepare(`
    UPDATE activity_registrations SET status = 'cancelled'
    WHERE volunteer_id = ? AND activity_id = ?
  `).run(volunteer_id, activityId);

  res.json({ message: '取消报名成功' });
});

router.get('/:id/registrations', (req: Request, res: Response) => {
  const registrations = db.prepare(`
    SELECT ar.*, v.name as volunteer_name, v.phone
    FROM activity_registrations ar
    JOIN volunteers v ON ar.volunteer_id = v.id
    WHERE ar.activity_id = ?
  `).all(req.params.id);
  res.json(registrations);
});

export default router;
