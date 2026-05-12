import { Router, Request, Response } from 'express';
import { getDb } from '../database';
import { VaccineBatch } from '../types';
import { daysBetween, createNotification, todayStr } from '../utils';

const router = Router();

function calculateStatus(expiryDate: string): 'normal' | 'near_expiry' | 'expired' {
  const today = todayStr();
  const days = daysBetween(today, expiryDate);
  if (days < 0) return 'expired';
  if (days <= 60) return 'near_expiry';
  return 'normal';
}

router.get('/', async (_req: Request, res: Response) => {
  const db = await getDb();
  const today = todayStr();
  
  const rows = await db.all('SELECT * FROM vaccine_batches');
  const batches: VaccineBatch[] = rows.map((row: any) => {
    let status: VaccineBatch['status'] = calculateStatus(row.expiry_date);
    if (row.status === 'suspended') status = 'suspended';
    
    return {
      id: row.id,
      name: row.name,
      manufacturer: row.manufacturer,
      batchNumber: row.batch_number,
      specification: row.specification,
      expiryDate: row.expiry_date,
      stock: row.stock,
      storageTemperature: row.storage_temperature,
      safeStock: row.safe_stock,
      status
    };
  });
  
  res.json(batches);
});

router.get('/:batchNumber', async (req: Request, res: Response) => {
  const db = await getDb();
  const row = await db.get('SELECT * FROM vaccine_batches WHERE batch_number = ?', req.params.batchNumber);
  
  if (!row) {
    return res.status(404).json({ error: 'Vaccine batch not found' });
  }
  
  let status: VaccineBatch['status'] = calculateStatus(row.expiry_date);
  if (row.status === 'suspended') status = 'suspended';
  
  res.json({
    id: row.id,
    name: row.name,
    manufacturer: row.manufacturer,
    batchNumber: row.batch_number,
    specification: row.specification,
    expiryDate: row.expiry_date,
    stock: row.stock,
    storageTemperature: row.storage_temperature,
    safeStock: row.safe_stock,
    status
  });
});

router.post('/', async (req: Request, res: Response) => {
  const { name, manufacturer, batchNumber, specification, expiryDate, stock, storageTemperature, safeStock } = req.body;
  
  if (!name || !manufacturer || !batchNumber || !specification || !expiryDate || stock === undefined || !storageTemperature) {
    return res.status(400).json({ error: 'Missing required fields' });
  }
  
  if (stock < 0) {
    return res.status(400).json({ error: 'Stock cannot be negative' });
  }
  
  const db = await getDb();
  
  try {
    const result = await db.run(
      `INSERT INTO vaccine_batches (name, manufacturer, batch_number, specification, expiry_date, stock, storage_temperature, safe_stock, status)
       VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'normal')`,
      name,
      manufacturer,
      batchNumber,
      specification,
      expiryDate,
      stock,
      storageTemperature,
      safeStock || 10
    );
    
    const status = calculateStatus(expiryDate);
    if (status !== 'normal') {
      await db.run('UPDATE vaccine_batches SET status = ? WHERE id = ?', status, result.lastID);
    }
    
    if (stock < (safeStock || 10)) {
      await createNotification('stock', `疫苗 ${name} (${batchNumber}) 库存低于安全库存量`, batchNumber);
    }
    
    res.status(201).json({ id: result.lastID, batchNumber });
  } catch (err: any) {
    if (err.code === 'SQLITE_CONSTRAINT') {
      return res.status(409).json({ error: 'Batch number already exists' });
    }
    res.status(500).json({ error: err.message });
  }
});

router.put('/:batchNumber/stock', async (req: Request, res: Response) => {
  const { change } = req.body;
  if (change === undefined) {
    return res.status(400).json({ error: 'Missing change field' });
  }
  
  const db = await getDb();
  const batch = await db.get('SELECT * FROM vaccine_batches WHERE batch_number = ?', req.params.batchNumber);
  
  if (!batch) {
    return res.status(404).json({ error: 'Vaccine batch not found' });
  }
  
  const newStock = batch.stock + change;
  if (newStock < 0) {
    return res.status(400).json({ error: 'Stock cannot be negative' });
  }
  
  await db.run('UPDATE vaccine_batches SET stock = ? WHERE id = ?', newStock, batch.id);
  
  if (newStock < batch.safe_stock && batch.stock >= batch.safe_stock) {
    await createNotification('stock', `疫苗 ${batch.name} (${batch.batch_number}) 库存低于安全库存量`, batch.batch_number);
  }
  
  res.json({ batchNumber: batch.batch_number, stock: newStock });
});

export default router;
