import { getDb } from './database';

export function todayStr(): string {
  return new Date().toISOString().split('T')[0];
}

export function daysBetween(date1: string, date2: string): number {
  const d1 = new Date(date1);
  const d2 = new Date(date2);
  const diffTime = d2.getTime() - d1.getTime();
  return Math.ceil(diffTime / (1000 * 60 * 60 * 24));
}

export function addDays(dateStr: string, days: number): string {
  const d = new Date(dateStr);
  d.setDate(d.getDate() + days);
  return d.toISOString().split('T')[0];
}

export async function generateCertificateNumber(): Promise<string> {
  const db = await getDb();
  const today = todayStr().replace(/-/g, '');
  const dateKey = today;
  
  const existing = await db.get('SELECT current_sequence FROM certificate_sequence WHERE date_key = ?', dateKey);
  let sequence: number;
  
  if (existing) {
    sequence = existing.current_sequence + 1;
    await db.run('UPDATE certificate_sequence SET current_sequence = ? WHERE date_key = ?', sequence, dateKey);
  } else {
    sequence = 1;
    await db.run('INSERT INTO certificate_sequence (date_key, current_sequence) VALUES (?, ?)', dateKey, sequence);
  }
  
  const seqStr = sequence.toString().padStart(6, '0');
  return `YM${today}${seqStr}`;
}

export async function createNotification(type: string, message: string, targetId: string): Promise<void> {
  const db = await getDb();
  await db.run(
    'INSERT INTO notifications (type, message, target_id, created_at, is_read) VALUES (?, ?, ?, ?, 0)',
    type,
    message,
    targetId,
    new Date().toISOString()
  );
}
