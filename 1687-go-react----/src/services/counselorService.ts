import { v4 as uuidv4 } from 'uuid';
import db from '../database';
import { Counselor } from '../types';

export function registerCounselor(data: {
  name: string;
  certificate: string;
  specialty: string;
  experience: number;
  fee: number;
}): Counselor {
  const id = uuidv4();
  const stmt = db.prepare(`
    INSERT INTO counselors (id, name, certificate, specialty, experience, fee)
    VALUES (?, ?, ?, ?, ?, ?)
  `);
  stmt.run(id, data.name, data.certificate, data.specialty, data.experience, data.fee);
  return getCounselorById(id) as Counselor;
}

export function getCounselorById(id: string): Counselor | undefined {
  const row = db.prepare('SELECT * FROM counselors WHERE id = ?').get(id) as any;
  if (!row) return undefined;
  return {
    ...row,
    experience: Number(row.experience),
    fee: Number(row.fee)
  };
}

export function getAllCounselors(): Counselor[] {
  const rows = db.prepare('SELECT * FROM counselors').all() as any[];
  return rows.map(row => ({
    ...row,
    experience: Number(row.experience),
    fee: Number(row.fee)
  }));
}

export function reviewCounselor(id: string, approved: boolean): Counselor | null {
  const counselor = getCounselorById(id);
  if (!counselor) return null;
  
  const newStatus = approved ? 'approved' : 'rejected';
  db.prepare('UPDATE counselors SET status = ? WHERE id = ?').run(newStatus, id);
  return getCounselorById(id) as Counselor;
}

export function setTimeSlot(counselorId: string, data: {
  isRecurring: boolean;
  dayOfWeek?: number;
  date?: string;
  startTime: string;
  endTime: string;
}): void {
  const existing = db.prepare(`
    SELECT * FROM time_slots 
    WHERE counselor_id = ? 
    AND is_recurring = ?
    ${data.isRecurring ? 'AND day_of_week = ?' : 'AND date = ?'}
    AND start_time = ? AND end_time = ?
  `).get(
    counselorId,
    data.isRecurring ? 1 : 0,
    data.isRecurring ? data.dayOfWeek : data.date,
    data.startTime,
    data.endTime
  );

  if (existing) {
    return;
  }

  const id = uuidv4();
  db.prepare(`
    INSERT INTO time_slots (id, counselor_id, is_recurring, day_of_week, date, start_time, end_time)
    VALUES (?, ?, ?, ?, ?, ?, ?)
  `).run(
    id,
    counselorId,
    data.isRecurring ? 1 : 0,
    data.dayOfWeek ?? null,
    data.date ?? null,
    data.startTime,
    data.endTime
  );
}

export function getAvailableTimeSlots(counselorId: string, date: string): any[] {
  const dateObj = new Date(date);
  const dayOfWeek = dateObj.getDay();

  const recurringSlots = db.prepare(`
    SELECT * FROM time_slots 
    WHERE counselor_id = ? 
    AND is_recurring = 1 
    AND day_of_week = ?
    AND is_booked = 0
  `).all(counselorId, dayOfWeek) as any[];

  const specificSlots = db.prepare(`
    SELECT * FROM time_slots 
    WHERE counselor_id = ? 
    AND is_recurring = 0 
    AND date = ?
    AND is_booked = 0
  `).all(counselorId, date) as any[];

  const bookedRecurringSlots = db.prepare(`
    SELECT ts.start_time, ts.end_time FROM time_slots ts
    JOIN appointments a ON ts.id = a.time_slot_id
    WHERE ts.counselor_id = ?
    AND a.appointment_date = ?
    AND a.status NOT IN ('cancelled', 'no_show')
  `).all(counselorId, date) as any[];

  const bookedTimes = new Set(bookedRecurringSlots.map(s => `${s.start_time}-${s.end_time}`));

  const availableRecurring = recurringSlots.filter(
    slot => !bookedTimes.has(`${slot.start_time}-${slot.end_time}`)
  );

  return [...availableRecurring, ...specificSlots];
}
