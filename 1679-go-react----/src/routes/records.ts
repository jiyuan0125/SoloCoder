import { Router, Request, Response } from 'express';
import db from '../database';
import { Activity, ActivityRegistration, Volunteer } from '../types';
import { calculateServiceMinutes, roundServiceMinutes } from '../utils/time';

const router = Router();

router.post('/:id/complete', (req: Request, res: Response) => {
  const { volunteer_id, actual_start_time, actual_end_time } = req.body;
  const activityId = parseInt(req.params.id);

  if (!volunteer_id || !actual_start_time || !actual_end_time) {
    return res.status(400).json({ error: '缺少必填字段' });
  }

  const registration = db.prepare(`
    SELECT * FROM activity_registrations
    WHERE volunteer_id = ? AND activity_id = ?
  `).get(volunteer_id, activityId) as ActivityRegistration | undefined;

  if (!registration || registration.status === 'completed' || registration.status === 'cancelled') {
    return res.status(404).json({ error: '活动参与记录不存在' });
  }

  const actualStart = new Date(actual_start_time);
  const actualEnd = new Date(actual_end_time);

  if (actualEnd <= actualStart) {
    return res.status(400).json({ error: '结束时间必须大于开始时间' });
  }

  const transaction = db.transaction(() => {
    const overlappingCompleteds = db.prepare(`
      SELECT ar.actual_start_time, ar.actual_end_time, a.name
      FROM activity_registrations ar
      JOIN activities a ON ar.activity_id = a.id
      WHERE ar.volunteer_id = ? 
        AND ar.status = 'completed' 
        AND ar.activity_id != ?
        AND ar.actual_start_time IS NOT NULL
        AND ar.actual_end_time IS NOT NULL
    `).all(volunteer_id, activityId) as Array<{
      actual_start_time: string;
      actual_end_time: string;
      name: string;
    }>;

    const overlaps: Array<{ start: Date; end: Date }> = [];
    for (const record of overlappingCompleteds) {
      overlaps.push({
        start: new Date(record.actual_start_time),
        end: new Date(record.actual_end_time)
      });
    }

    const rawMinutes = calculateServiceMinutes(actualStart, actualEnd, overlaps);
    const roundedMinutes = roundServiceMinutes(rawMinutes);
    const hours = roundedMinutes / 60;

    db.prepare(`
      UPDATE activity_registrations 
      SET status = 'completed',
          actual_start_time = ?,
          actual_end_time = ?,
          calculated_hours = ?
      WHERE volunteer_id = ? AND activity_id = ?
    `).run(actual_start_time, actual_end_time, hours, volunteer_id, activityId);

    const volunteer = db.prepare('SELECT * FROM volunteers WHERE id = ?').get(volunteer_id) as Volunteer;
    const newTotalHours = volunteer.total_hours + hours;
    const isBackbone = newTotalHours >= 100 ? 1 : 0;

    db.prepare(`
      UPDATE volunteers 
      SET total_hours = ?, is_backbone = ?
      WHERE id = ?
    `).run(newTotalHours, isBackbone, volunteer_id);

    const year = actualStart.getFullYear();
    const existingRecord = db.prepare(`
      SELECT id FROM service_records 
      WHERE volunteer_id = ? AND activity_id = ?
    `).get(volunteer_id, activityId);

    if (!existingRecord) {
      db.prepare(`
        INSERT INTO service_records (volunteer_id, activity_id, hours, year)
        VALUES (?, ?, ?, ?)
      `).run(volunteer_id, activityId, hours, year);
    } else {
      db.prepare(`
        UPDATE service_records SET hours = ?, year = ?
        WHERE volunteer_id = ? AND activity_id = ?
      `).run(hours, year, volunteer_id, activityId);
    }

    const updatedVolunteer = db.prepare('SELECT * FROM volunteers WHERE id = ?').get(volunteer_id) as Volunteer;
    const updatedRegistration = db.prepare(`
      SELECT * FROM activity_registrations 
      WHERE volunteer_id = ? AND activity_id = ?
    `).get(volunteer_id, activityId) as ActivityRegistration;

    return {
      registration: updatedRegistration,
      volunteer: {
        ...updatedVolunteer,
        skills: JSON.parse(updatedVolunteer.skills),
        availability: JSON.parse(updatedVolunteer.availability),
        is_backbone: !!updatedVolunteer.is_backbone
      },
      raw_minutes: rawMinutes,
      rounded_minutes: roundedMinutes,
      calculated_hours: hours,
      deducted_overlaps: overlaps.length
    };
  });

  try {
    const result = transaction();
    res.json(result);
  } catch (err) {
    res.status(500).json({ error: '处理时长统计时出错' });
  }
});

router.get('/volunteer/:volunteer_id', (req: Request, res: Response) => {
  const records = db.prepare(`
    SELECT sr.*, a.name as activity_name, a.date as activity_date
    FROM service_records sr
    JOIN activities a ON sr.activity_id = a.id
    WHERE sr.volunteer_id = ?
    ORDER BY sr.created_at DESC
  `).all(req.params.volunteer_id);
  res.json(records);
});

router.get('/ranking', (req: Request, res: Response) => {
  const currentYear = new Date().getFullYear();
  const year = req.query.year ? parseInt(req.query.year as string) : currentYear;

  const ranking = db.prepare(`
    SELECT 
      v.id, 
      v.name, 
      v.phone,
      COALESCE(SUM(sr.hours), 0) as annual_hours,
      v.total_hours
    FROM volunteers v
    LEFT JOIN service_records sr ON v.id = sr.volunteer_id AND sr.year = ?
    GROUP BY v.id
    ORDER BY annual_hours DESC
  `).all(year);

  res.json({ year, ranking });
});

export default router;
