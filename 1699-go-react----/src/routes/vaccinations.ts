import { Router, Request, Response } from 'express';
import { getDb } from '../database';
import { VaccinationRecord, AdverseReaction } from '../types';
import { daysBetween, createNotification, todayStr, addDays, generateCertificateNumber } from '../utils';

const router = Router();

const VACCINE_INTERVALS: Record<string, number> = {
  '新冠疫苗': 21,
  'HPV疫苗': 60,
  '乙肝疫苗': 30,
  '狂犬疫苗': 3,
  '流感疫苗': 0
};

function getNextVaccinationDate(vaccineName: string, currentDate: string, dose: number): string | null {
  const interval = VACCINE_INTERVALS[vaccineName] ?? 0;
  if (interval === 0) return null;
  return addDays(currentDate, interval);
}

router.get('/', async (_req: Request, res: Response) => {
  const db = await getDb();
  const rows = await db.all('SELECT * FROM vaccination_records');
  
  const today = todayStr();
  const records: VaccinationRecord[] = [];
  
  for (const row of rows) {
    let isOverdue = !!row.is_overdue;
    if (!isOverdue && row.next_vaccination_date) {
      const daysLate = daysBetween(row.next_vaccination_date, today);
      if (daysLate > 7) {
        isOverdue = true;
        await db.run('UPDATE vaccination_records SET is_overdue = 1 WHERE id = ?', row.id);
        await createNotification(
          'overdue',
          `居民 ${row.resident_name} 的 ${row.vaccine_name} 疫苗接种已逾期未种`,
          row.resident_id
        );
      }
    }
    
    records.push({
      id: row.id,
      certificateNumber: row.certificate_number,
      reservationId: row.reservation_id,
      residentId: row.resident_id,
      residentName: row.resident_name,
      vaccineName: row.vaccine_name,
      batchNumber: row.batch_number,
      injectionSite: row.injection_site,
      dose: row.dose,
      vaccinationDate: row.vaccination_date,
      nextVaccinationDate: row.next_vaccination_date,
      isOverdue
    });
  }
  
  res.json(records);
});

router.get('/:recordId', async (req: Request, res: Response) => {
  const db = await getDb();
  const row = await db.get('SELECT * FROM vaccination_records WHERE id = ?', parseInt(req.params.recordId));
  
  if (!row) {
    return res.status(404).json({ error: 'Vaccination record not found' });
  }
  
  const today = todayStr();
  let isOverdue = !!row.is_overdue;
  if (!isOverdue && row.next_vaccination_date) {
    const daysLate = daysBetween(row.next_vaccination_date, today);
    if (daysLate > 7) {
      isOverdue = true;
      await db.run('UPDATE vaccination_records SET is_overdue = 1 WHERE id = ?', row.id);
    }
  }
  
  res.json({
    id: row.id,
    certificateNumber: row.certificate_number,
    reservationId: row.reservation_id,
    residentId: row.resident_id,
    residentName: row.resident_name,
    vaccineName: row.vaccine_name,
    batchNumber: row.batch_number,
    injectionSite: row.injection_site,
    dose: row.dose,
    vaccinationDate: row.vaccination_date,
    nextVaccinationDate: row.next_vaccination_date,
    isOverdue
  });
});

router.post('/', async (req: Request, res: Response) => {
  const { reservationId, residentId, residentName, vaccineName, injectionSite, dose, vaccinationDate, isBooster } = req.body;
  
  if (!residentId || !residentName || !vaccineName || !injectionSite || !dose) {
    return res.status(400).json({ error: 'Missing required fields' });
  }
  
  const vDate = vaccinationDate || todayStr();
  const db = await getDb();
  
  const activeBatches = await db.all(
    `SELECT * FROM vaccine_batches 
     WHERE name = ? AND stock > 0 AND status != 'suspended'
     ORDER BY expiry_date ASC`,
    vaccineName
  );
  
  let selectedBatch = null;
  for (const batch of activeBatches) {
    const today = todayStr();
    const status = daysBetween(today, batch.expiry_date) < 0 ? 'expired' : 
                   daysBetween(today, batch.expiry_date) <= 60 ? 'near_expiry' : 'normal';
    
    if (status === 'expired') continue;
    
    if (status === 'near_expiry' && !isBooster) {
      continue;
    }
    
    selectedBatch = batch;
    break;
  }
  
  if (!selectedBatch) {
    return res.status(400).json({ error: 'No available vaccine batch for this vaccination' });
  }
  
  await db.run('UPDATE vaccine_batches SET stock = stock - 1 WHERE id = ?', selectedBatch.id);
  
  const newStock = selectedBatch.stock - 1;
  if (newStock < selectedBatch.safe_stock && selectedBatch.stock >= selectedBatch.safe_stock) {
    await createNotification('stock', `疫苗 ${selectedBatch.name} (${selectedBatch.batch_number}) 库存低于安全库存量`, selectedBatch.batch_number);
  }
  
  const prevRecords = await db.all(
    `SELECT * FROM vaccination_records 
     WHERE resident_id = ? AND vaccine_name = ? 
     ORDER BY vaccination_date DESC LIMIT 1`,
    residentId,
    vaccineName
  );
  
  const actualDose = dose || (prevRecords.length > 0 ? (prevRecords[0].dose + 1) : 1);
  const nextDate = getNextVaccinationDate(vaccineName, vDate, actualDose);
  
  const certificateNumber = await generateCertificateNumber();
  
  const result = await db.run(
    `INSERT INTO vaccination_records 
     (certificate_number, reservation_id, resident_id, resident_name, vaccine_name, batch_number, 
      injection_site, dose, vaccination_date, next_vaccination_date, is_overdue)
     VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0)`,
    certificateNumber,
    reservationId || null,
    residentId,
    residentName,
    vaccineName,
    selectedBatch.batch_number,
    injectionSite,
    actualDose,
    vDate,
    nextDate
  );
  
  if (reservationId) {
    await db.run('UPDATE reservations SET status = ? WHERE id = ?', 'completed', reservationId);
  }
  
  res.status(201).json({
    id: result.lastID,
    certificateNumber,
    batchNumber: selectedBatch.batch_number,
    nextVaccinationDate: nextDate
  });
});

router.post('/:recordId/reactions/:reactionId/report', async (req: Request, res: Response) => {
  const recordId = parseInt(req.params.recordId);
  const reactionId = req.params.reactionId;
  
  const { symptoms, severity, reportDate } = req.body;
  
  if (!symptoms || !severity) {
    return res.status(400).json({ error: 'Missing required fields' });
  }
  
  if (!/^[a-zA-Z0-9_-]+$/.test(reactionId)) {
    return res.status(404).json({ error: 'Invalid reactionId format' });
  }
  
  const db = await getDb();
  const record = await db.get('SELECT * FROM vaccination_records WHERE id = ?', recordId);
  
  if (!record) {
    return res.status(404).json({ error: 'Vaccination record not found' });
  }
  
  const rDate = reportDate || todayStr();
  const daysSinceVaccination = daysBetween(record.vaccination_date, rDate);
  
  if (daysSinceVaccination > 14) {
    return res.status(400).json({ error: 'Cannot report adverse reaction after 14 days of vaccination' });
  }
  
  try {
    const result = await db.run(
      `INSERT INTO adverse_reactions (record_id, reaction_id, symptoms, severity, report_date, is_handled)
       VALUES (?, ?, ?, ?, ?, 0)`,
      recordId,
      reactionId,
      symptoms,
      severity,
      rDate
    );
    
    if (severity === 'severe') {
      await createNotification(
        'emergency',
        `紧急待办：接种记录 ${recordId} 报告了严重不良反应: ${symptoms}`,
        `record_${recordId}`
      );
    }
    
    const reactionCount = await db.get(
      `SELECT COUNT(*) as count FROM adverse_reactions 
       WHERE record_id IN (SELECT id FROM vaccination_records WHERE batch_number = ?)`,
      record.batch_number
    );
    
    if (reactionCount.count >= 5) {
      await db.run(
        `UPDATE vaccine_batches SET status = 'suspended' WHERE batch_number = ?`,
        record.batch_number
      );
      await createNotification(
        'emergency',
        `批次 ${record.batch_number} 不良反应超过5例，已自动暂停该批次`,
        record.batch_number
      );
    }
    
    res.status(201).json({ id: result.lastID, reactionId });
  } catch (err: any) {
    if (err.code === 'SQLITE_CONSTRAINT') {
      return res.status(409).json({ error: 'Reaction ID already exists' });
    }
    res.status(500).json({ error: err.message });
  }
});

router.get('/:recordId/reactions', async (req: Request, res: Response) => {
  const recordId = parseInt(req.params.recordId);
  const db = await getDb();
  
  const record = await db.get('SELECT * FROM vaccination_records WHERE id = ?', recordId);
  if (!record) {
    return res.status(404).json({ error: 'Vaccination record not found' });
  }
  
  const rows = await db.all('SELECT * FROM adverse_reactions WHERE record_id = ?', recordId);
  
  const reactions: AdverseReaction[] = rows.map((row: any) => ({
    id: row.id,
    recordId: row.record_id,
    reactionId: row.reaction_id,
    symptoms: row.symptoms,
    severity: row.severity,
    reportDate: row.report_date,
    isHandled: !!row.is_handled
  }));
  
  res.json(reactions);
});

router.get('/:recordId/reactions/:reactionId', async (req: Request, res: Response) => {
  const recordId = parseInt(req.params.recordId);
  const reactionId = req.params.reactionId;
  
  const db = await getDb();
  
  const record = await db.get('SELECT * FROM vaccination_records WHERE id = ?', recordId);
  if (!record) {
    return res.status(404).json({ error: 'Vaccination record not found' });
  }
  
  const row = await db.get('SELECT * FROM adverse_reactions WHERE record_id = ? AND reaction_id = ?', recordId, reactionId);
  
  if (!row) {
    return res.status(404).json({ error: 'Adverse reaction not found' });
  }
  
  res.json({
    id: row.id,
    recordId: row.record_id,
    reactionId: row.reaction_id,
    symptoms: row.symptoms,
    severity: row.severity,
    reportDate: row.report_date,
    isHandled: !!row.is_handled
  });
});

export default router;
