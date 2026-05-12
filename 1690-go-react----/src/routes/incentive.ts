import { Router, Request, Response } from 'express';
import db from '../database';
import { generateId, getNow, getStarLevel, addDays, isWithinDays } from '../utils';

const router = Router();

router.get('/volunteers/:id/points', (req: Request, res: Response) => {
  const volunteerId = req.params.id;

  const volunteer = db.prepare('SELECT * FROM volunteers WHERE id = ?').get(volunteerId);
  if (!volunteer) {
    return res.status(404).json({ error: '志愿者不存在' });
  }

  const pointsRecords = db.prepare('SELECT * FROM points WHERE volunteer_id = ? ORDER BY created_at DESC')
    .all(volunteerId);

  const totalPoints = pointsRecords.reduce((sum: number, record: any) => sum + record.amount, 0);

  res.json({
    total_points: totalPoints,
    records: pointsRecords
  });
});

router.post('/points/redemption', (req: Request, res: Response) => {
  const { volunteer_id, type, item_name, points_cost } = req.body;

  if (!['prize', 'certificate'].includes(type)) {
    return res.status(400).json({ error: 'type 必须是 prize 或 certificate' });
  }

  if (points_cost <= 0) {
    return res.status(400).json({ error: '兑换积分必须为正数' });
  }

  const volunteer = db.prepare('SELECT * FROM volunteers WHERE id = ?').get(volunteer_id);
  if (!volunteer) {
    return res.status(404).json({ error: '志愿者不存在' });
  }

  const pointsRecords = db.prepare('SELECT * FROM points WHERE volunteer_id = ?').all(volunteer_id);
  const totalPoints = pointsRecords.reduce((sum: number, record: any) => sum + record.amount, 0);

  if (totalPoints < points_cost) {
    return res.status(400).json({ error: '积分不足，无法兑换' });
  }

  const id = generateId();
  const now = getNow();

  db.prepare(`
    INSERT INTO points (id, volunteer_id, amount, description, source, created_at)
    VALUES (?, ?, ?, ?, 'redemption', ?)
  `).run(id, volunteer_id, -points_cost, `兑换${type === 'prize' ? '奖品' : '时长证明'}: ${item_name}`, now);

  res.json({
    success: true,
    type,
    item_name,
    points_cost,
    remaining_points: totalPoints - points_cost
  });
});

router.get('/volunteers/:id/stars', (req: Request, res: Response) => {
  const volunteerId = req.params.id;

  const volunteer = db.prepare('SELECT * FROM volunteers WHERE id = ?').get(volunteerId);
  if (!volunteer) {
    return res.status(404).json({ error: '志愿者不存在' });
  }

  const registrations = db.prepare(`
    SELECT hours FROM registrations 
    WHERE volunteer_id = ? AND status = 'approved'
  `).all(volunteerId);

  const totalHours = registrations.reduce((sum: number, r: any) => sum + (r.hours || 0), 0);
  const starLevel = getStarLevel(totalHours);

  const starNames = ['无', '一星', '二星', '三星', '四星', '五星'];

  res.json({
    total_hours: totalHours,
    star_level: starLevel,
    star_name: starNames[starLevel],
    next_star_hours: starLevel < 5 ? [50, 100, 200, 500, 1000][starLevel] : null
  });
});

router.post('/reminders/generate', (req: Request, res: Response) => {
  const now = new Date();
  const volunteers = db.prepare('SELECT * FROM volunteers WHERE entry_training_completed = 1').all();
  let generatedCount = 0;

  for (const volunteer of volunteers as any[]) {
    if (!volunteer.last_annual_training_date) continue;

    const nextAnnualDate = addDays(volunteer.last_annual_training_date, 365);
    const reminderDueDate = addDays(nextAnnualDate, -30);

    if (isWithinDays(reminderDueDate, 30)) {
      const existingReminder = db.prepare(`
        SELECT id FROM reminders 
        WHERE volunteer_id = ? AND type = 'annual_training' 
        AND is_read = 0 AND created_at >= ?
      `).get(volunteer.id, addDays(now.toISOString(), -30));

      if (!existingReminder) {
        const id = generateId();
        const daysUntil = Math.ceil((new Date(nextAnnualDate).getTime() - now.getTime()) / (1000 * 60 * 60 * 24));

        db.prepare(`
          INSERT INTO reminders (id, volunteer_id, type, message, due_date, created_at)
          VALUES (?, ?, 'annual_training', ?, ?, ?)
        `).run(
          id,
          volunteer.id,
          `您的年度复训将在${daysUntil}天后到期，请及时完成培训`,
          nextAnnualDate,
          now.toISOString()
        );
        generatedCount++;
      }
    }
  }

  res.json({ generated: generatedCount });
});

router.get('/reminders/:volunteer_id', (req: Request, res: Response) => {
  const reminders = db.prepare(`
    SELECT * FROM reminders 
    WHERE volunteer_id = ? 
    ORDER BY created_at DESC
  `).all(req.params.volunteer_id);

  res.json(reminders);
});

router.put('/reminders/:id/read', (req: Request, res: Response) => {
  const reminder = db.prepare('SELECT * FROM reminders WHERE id = ?').get(req.params.id);
  if (!reminder) {
    return res.status(404).json({ error: '提醒不存在' });
  }

  db.prepare('UPDATE reminders SET is_read = 1 WHERE id = ?').run(req.params.id);
  res.json({ success: true });
});

export default router;
