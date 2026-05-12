import { Router, Request, Response } from 'express';
import db from '../database';
import { Volunteer } from '../types';
import { getBadge } from '../utils/time';

const router = Router();

router.post('/', (req: Request, res: Response) => {
  const { name, phone, id_card, skills, availability } = req.body;

  if (!name || !phone || id_card === undefined || skills === undefined || availability === undefined) {
    return res.status(400).json({ error: '缺少必填字段' });
  }

  if (!id_card || id_card.trim() === '') {
    return res.status(400).json({ error: '身份证号不能为空' });
  }

  const existing = db.prepare('SELECT id FROM volunteers WHERE phone = ?').get(phone);
  if (existing) {
    return res.status(409).json({ error: '该手机号已注册' });
  }

  const stmt = db.prepare(`
    INSERT INTO volunteers (name, phone, id_card, skills, availability)
    VALUES (?, ?, ?, ?, ?)
  `);
  const result = stmt.run(name, phone, id_card, JSON.stringify(skills), JSON.stringify(availability));

  const volunteer = db.prepare('SELECT * FROM volunteers WHERE id = ?').get(result.lastInsertRowid) as Volunteer;
  res.status(201).json({
    ...volunteer,
    skills: JSON.parse(volunteer.skills),
    availability: JSON.parse(volunteer.availability),
    is_backbone: !!volunteer.is_backbone
  });
});

router.get('/', (req: Request, res: Response) => {
  const volunteers = db.prepare('SELECT * FROM volunteers').all() as Volunteer[];
  res.json(volunteers.map(v => ({
    ...v,
    skills: JSON.parse(v.skills),
    availability: JSON.parse(v.availability),
    is_backbone: !!v.is_backbone,
    badge: getBadge(v.total_hours)
  })));
});

router.get('/:id', (req: Request, res: Response) => {
  const volunteer = db.prepare('SELECT * FROM volunteers WHERE id = ?').get(req.params.id) as Volunteer | undefined;
  if (!volunteer) {
    return res.status(404).json({ error: '志愿者不存在' });
  }
  res.json({
    ...volunteer,
    skills: JSON.parse(volunteer.skills),
    availability: JSON.parse(volunteer.availability),
    is_backbone: !!volunteer.is_backbone,
    badge: getBadge(volunteer.total_hours)
  });
});

router.get('/:id/stats', (req: Request, res: Response) => {
  const volunteer = db.prepare('SELECT * FROM volunteers WHERE id = ?').get(req.params.id) as Volunteer | undefined;
  if (!volunteer) {
    return res.status(404).json({ error: '志愿者不存在' });
  }

  const currentYear = new Date().getFullYear();
  const annualRecords = db.prepare(`
    SELECT COALESCE(SUM(hours), 0) as annual_hours
    FROM service_records
    WHERE volunteer_id = ? AND year = ?
  `).get(req.params.id, currentYear) as { annual_hours: number };

  const annualHours = annualRecords.annual_hours;

  const activityCount = db.prepare(`
    SELECT COUNT(*) as count
    FROM activity_registrations
    WHERE volunteer_id = ? AND status = 'completed'
  `).get(req.params.id) as { count: number };

  res.json({
    total_hours: volunteer.total_hours,
    annual_hours: annualHours,
    is_backbone: !!volunteer.is_backbone,
    badge: getBadge(annualHours),
    completed_activities: activityCount.count
  });
});

export default router;
