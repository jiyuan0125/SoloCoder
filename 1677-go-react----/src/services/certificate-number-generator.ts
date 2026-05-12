import { db } from '../database';
import dayjs from 'dayjs';

let sequenceLock = false;

export function generateCertificateNumber(): string {
  const dateStr = dayjs().format('YYYYMMDD');
  
  while (sequenceLock) {
    // 简单自旋等待，在单线程 Node.js 中这个不会阻塞太久
  }
  sequenceLock = true;
  
  try {
    const transaction = db.transaction(() => {
      const existing = db.prepare(
        'SELECT sequence FROM certificate_sequences WHERE date = ?'
      ).get(dateStr) as { sequence: number } | undefined;
      
      let nextSequence: number;
      if (existing) {
        nextSequence = existing.sequence + 1;
        db.prepare(
          'UPDATE certificate_sequences SET sequence = ? WHERE date = ?'
        ).run(nextSequence, dateStr);
      } else {
        const todayStart = dayjs().startOf('day').format('YYYY-MM-DD HH:mm:ss');
        const todayEnd = dayjs().endOf('day').format('YYYY-MM-DD HH:mm:ss');
        
        const todayCount = db.prepare(
          `SELECT COUNT(*) as cnt FROM certificates 
           WHERE issued_at >= ? AND issued_at <= ?`
        ).get(todayStart, todayEnd) as { cnt: number };
        
        nextSequence = todayCount.cnt + 1;
        
        db.prepare(
          'INSERT OR REPLACE INTO certificate_sequences (date, sequence) VALUES (?, ?)'
        ).run(dateStr, nextSequence);
      }
      
      const sequenceStr = nextSequence.toString().padStart(6, '0');
      return `DJ${dateStr}${sequenceStr}`;
    });
    
    return transaction();
  } finally {
    sequenceLock = false;
  }
}

export function generateVoucherNumber(): string {
  const dateStr = dayjs().format('YYYYMMDD');
  
  while (sequenceLock) {
    // 简单自旋等待
  }
  sequenceLock = true;
  
  try {
    const transaction = db.transaction(() => {
      const existing = db.prepare(
        'SELECT sequence FROM vouchers_sequences WHERE date = ?'
      ).get(dateStr) as { sequence: number } | undefined;
      
      let nextSequence: number;
      if (existing) {
        nextSequence = existing.sequence + 1;
        db.prepare(
          'UPDATE vouchers_sequences SET sequence = ? WHERE date = ?'
        ).run(nextSequence, dateStr);
      } else {
        const todayStart = dayjs().startOf('day').format('YYYY-MM-DD HH:mm:ss');
        const todayEnd = dayjs().endOf('day').format('YYYY-MM-DD HH:mm:ss');
        
        const todayCount = db.prepare(
          `SELECT COUNT(*) as cnt FROM tax_deduction_vouchers 
           WHERE issued_at >= ? AND issued_at <= ?`
        ).get(todayStart, todayEnd) as { cnt: number };
        
        nextSequence = todayCount.cnt + 1;
        
        db.prepare(
          'INSERT OR REPLACE INTO vouchers_sequences (date, sequence) VALUES (?, ?)'
        ).run(dateStr, nextSequence);
      }
      
      const sequenceStr = nextSequence.toString().padStart(6, '0');
      return `KD${dateStr}${sequenceStr}`;
    });
    
    return transaction();
  } finally {
    sequenceLock = false;
  }
}
