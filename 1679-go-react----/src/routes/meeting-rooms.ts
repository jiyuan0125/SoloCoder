import { Router, Request, Response } from 'express';
import db from '../database';
import { MeetingRoom, RoomBooking } from '../types';
import { timeOverlap } from '../utils/time';

const router = Router();

router.get('/', (req: Request, res: Response) => {
  const rooms = db.prepare('SELECT * FROM meeting_rooms').all() as MeetingRoom[];
  res.json(rooms);
});

router.get('/:id/bookings', (req: Request, res: Response) => {
  const room = db.prepare('SELECT * FROM meeting_rooms WHERE id = ?').get(req.params.id) as MeetingRoom | undefined;
  if (!room) {
    return res.status(404).json({ error: '会议室不存在' });
  }

  const bookings = db.prepare(`
    SELECT * FROM room_bookings
    WHERE room_id = ?
    ORDER BY date, start_time
  `).all(req.params.id) as RoomBooking[];
  
  res.json(bookings);
});

router.post('/book', (req: Request, res: Response) => {
  const { room_id, date, start_time, end_time, team_name, purpose } = req.body;

  if (!room_id || !date || !start_time || !end_time || !team_name) {
    return res.status(400).json({ error: '缺少必填字段' });
  }

  const room = db.prepare('SELECT * FROM meeting_rooms WHERE id = ?').get(room_id) as MeetingRoom | undefined;
  if (!room) {
    return res.status(404).json({ error: '会议室不存在' });
  }

  const existingBookings = db.prepare(`
    SELECT * FROM room_bookings
    WHERE room_id = ? AND date = ?
  `).all(room_id, date) as RoomBooking[];

  for (const booking of existingBookings) {
    if (timeOverlap(start_time, end_time, booking.start_time, booking.end_time)) {
      return res.status(409).json({ 
        error: `该会议室在 ${date} ${start_time}-${end_time} 已有预约` 
      });
    }
  }

  const stmt = db.prepare(`
    INSERT INTO room_bookings (room_id, date, start_time, end_time, team_name, purpose)
    VALUES (?, ?, ?, ?, ?, ?)
  `);
  const result = stmt.run(room_id, date, start_time, end_time, team_name, purpose || null);

  const booking = db.prepare('SELECT * FROM room_bookings WHERE id = ?').get(result.lastInsertRowid) as RoomBooking;
  res.status(201).json(booking);
});

router.get('/bookings', (req: Request, res: Response) => {
  const bookings = db.prepare(`
    SELECT rb.*, mr.name as room_name
    FROM room_bookings rb
    JOIN meeting_rooms mr ON rb.room_id = mr.id
    ORDER BY rb.date, rb.start_time
  `).all();
  res.json(bookings);
});

export default router;
