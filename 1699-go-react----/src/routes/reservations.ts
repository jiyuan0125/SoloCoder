import { Router, Request, Response } from 'express';
import { getDb } from '../database';
import { Reservation } from '../types';
import { daysBetween, createNotification, todayStr } from '../utils';

const router = Router();
const MAX_PER_SLOT = 20;

function parseTimeSlot(slot: string): number {
  const [hours, minutes] = slot.split(':').map(Number);
  return hours * 60 + minutes;
}

function getAdjacentSlot(timeSlot: string, allSlots: string[]): string | null {
  const currentMinutes = parseTimeSlot(timeSlot);
  let bestSlot: string | null = null;
  let bestDiff = Infinity;
  
  for (const slot of allSlots) {
    if (slot === timeSlot) continue;
    const diff = Math.abs(parseTimeSlot(slot) - currentMinutes);
    if (diff < bestDiff) {
      bestDiff = diff;
      bestSlot = slot;
    }
  }
  
  return bestSlot;
}

async function checkAndMergeSlots(date: string, vaccineName: string): Promise<void> {
  const db = await getDb();
  const today = todayStr();
  const daysUntil = daysBetween(today, date);
  
  if (daysUntil >= 1) return;
  
  const slots = await db.all(
    `SELECT time_slot, COUNT(*) as count FROM reservations 
     WHERE date = ? AND vaccine_name = ? AND status = 'active'
     GROUP BY time_slot`,
    date,
    vaccineName
  );
  
  const allSlotNames = slots.map((s: any) => s.time_slot);
  
  for (const slot of slots) {
    if (slot.count < 5) {
      const adjacent = getAdjacentSlot(slot.time_slot, allSlotNames);
      if (!adjacent) continue;
      
      const adjacentCount = slots.find((s: any) => s.time_slot === adjacent)?.count || 0;
      if (adjacentCount >= MAX_PER_SLOT) continue;
      
      const reservations = await db.all(
        `SELECT * FROM reservations WHERE date = ? AND vaccine_name = ? AND time_slot = ? AND status = 'active'`,
        date,
        vaccineName,
        slot.time_slot
      );
      
      for (const r of reservations) {
        await db.run(
          `UPDATE reservations SET time_slot = ?, merged_from = ? WHERE id = ?`,
          adjacent,
          slot.time_slot,
          r.id
        );
        await createNotification(
          'appointment_merge',
          `您的 ${vaccineName} 预约时段已从 ${slot.time_slot} 合并到 ${adjacent}`,
          r.resident_id
        );
      }
    }
  }
}

router.get('/', async (_req: Request, res: Response) => {
  const db = await getDb();
  const rows = await db.all('SELECT * FROM reservations');
  
  const reservations: Reservation[] = rows.map((row: any) => ({
    id: row.id,
    residentId: row.resident_id,
    residentName: row.resident_name,
    vaccineName: row.vaccine_name,
    date: row.date,
    timeSlot: row.time_slot,
    status: row.status,
    isBooster: !!row.is_booster,
    mergedFrom: row.merged_from
  }));
  
  res.json(reservations);
});

router.get('/:id', async (req: Request, res: Response) => {
  const db = await getDb();
  const row = await db.get('SELECT * FROM reservations WHERE id = ?', parseInt(req.params.id));
  
  if (!row) {
    return res.status(404).json({ error: 'Reservation not found' });
  }
  
  res.json({
    id: row.id,
    residentId: row.resident_id,
    residentName: row.resident_name,
    vaccineName: row.vaccine_name,
    date: row.date,
    timeSlot: row.time_slot,
    status: row.status,
    isBooster: !!row.is_booster,
    mergedFrom: row.merged_from
  });
});

router.post('/', async (req: Request, res: Response) => {
  const { residentId, residentName, vaccineName, date, timeSlot, isBooster } = req.body;
  
  if (!residentId || !residentName || !vaccineName || !date || !timeSlot) {
    return res.status(400).json({ error: 'Missing required fields' });
  }
  
  const db = await getDb();
  
  const existingSameDay = await db.get(
    `SELECT * FROM reservations 
     WHERE resident_id = ? AND date = ? AND status = 'active'`,
    residentId,
    date
  );
  
  if (existingSameDay) {
    return res.status(409).json({ error: 'Same resident cannot have multiple appointments on the same day' });
  }
  
  await checkAndMergeSlots(date, vaccineName);
  
  const slotCount = await db.get(
    `SELECT COUNT(*) as count FROM reservations 
     WHERE date = ? AND vaccine_name = ? AND time_slot = ? AND status = 'active'`,
    date,
    vaccineName,
    timeSlot
  );
  
  if (slotCount.count >= MAX_PER_SLOT) {
    return res.status(400).json({ error: 'Time slot is full' });
  }
  
  const result = await db.run(
    `INSERT INTO reservations (resident_id, resident_name, vaccine_name, date, time_slot, status, is_booster)
     VALUES (?, ?, ?, ?, ?, 'active', ?)`,
    residentId,
    residentName,
    vaccineName,
    date,
    timeSlot,
    isBooster ? 1 : 0
  );
  
  res.status(201).json({ id: result.lastID, date, timeSlot });
});

router.delete('/:id', async (req: Request, res: Response) => {
  const db = await getDb();
  const row = await db.get('SELECT * FROM reservations WHERE id = ?', parseInt(req.params.id));
  
  if (!row) {
    return res.status(404).json({ error: 'Reservation not found' });
  }
  
  const today = todayStr();
  const daysUntil = daysBetween(today, row.date);
  
  if (daysUntil <= 1) {
    return res.status(400).json({ error: 'Cannot cancel appointment within 24 hours of vaccination' });
  }
  
  await db.run('UPDATE reservations SET status = ? WHERE id = ?', 'cancelled', row.id);
  res.json({ success: true });
});

export default router;
