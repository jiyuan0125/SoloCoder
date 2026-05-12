import { Router, Request, Response } from 'express';
import db from '../db';

const router = Router();

const DAY_MS = 24 * 60 * 60 * 1000;
const LATE_FEE_RATE = 0.0005;

interface FeeBill {
  id: number;
  resident_id: number;
  year: number;
  quarter: number;
  property_fee: number;
  water_fee: number;
  electricity_fee: number;
  gas_fee: number;
  other_fee: number;
  late_fee: number;
  is_paid: number;
  paid_at: number | null;
  due_date: number;
  created_at: number;
}

interface Resident {
  id: number;
  name: string;
  building: string;
  unit: string;
  room: string;
  created_at: number;
}

function calculateTotalFee(bill: FeeBill): number {
  return bill.property_fee + bill.water_fee + bill.electricity_fee + 
         bill.gas_fee + bill.other_fee + bill.late_fee;
}

function calculateLateFee(bill: FeeBill, now: number): number {
  if (bill.is_paid) return 0;
  if (now <= bill.due_date) return 0;

  const overdueDays = Math.floor((now - bill.due_date) / DAY_MS);
  if (overdueDays <= 0) return 0;

  const principal = bill.property_fee + bill.water_fee + bill.electricity_fee + 
                    bill.gas_fee + bill.other_fee;
  
  if (principal <= 0) return 0;

  const dailyFee = principal * LATE_FEE_RATE;
  const monthlyFee = dailyFee * 30;
  
  if (monthlyFee < 1) return 0;
  
  return Math.floor(dailyFee * overdueDays);
}

function getQuarterDueDate(year: number, quarter: number): number {
  const dueMonth = quarter * 3;
  const dueYear = quarter === 4 ? year + 1 : year;
  const lastDay = new Date(dueYear, dueMonth, 0).getDate();
  return new Date(dueYear, dueMonth - 1, lastDay, 23, 59, 59, 999).getTime();
}

router.post('/generate', (req: Request, res: Response) => {
  const { resident_id, year, quarter, property_fee, water_fee, electricity_fee, gas_fee, other_fee } = req.body;

  if (resident_id === undefined || isNaN(parseInt(resident_id))) {
    return res.status(400).json({ error: '居民ID无效' });
  }

  const resident = db.prepare('SELECT * FROM residents WHERE id = ?').get(parseInt(resident_id)) as Resident | undefined;
  if (!resident) {
    return res.status(404).json({ error: '居民不存在' });
  }

  if (!year || !quarter || ![1, 2, 3, 4].includes(parseInt(quarter))) {
    return res.status(400).json({ error: '年份或季度无效' });
  }

  const feeFields = { property_fee, water_fee, electricity_fee, gas_fee, other_fee };
  for (const [key, value] of Object.entries(feeFields)) {
    if (value !== undefined && (isNaN(parseInt(value)) || parseInt(value) < 0)) {
      return res.status(400).json({ error: `${key} 必须是非负整数（单位：分）` });
    }
  }

  const residentId = parseInt(resident_id);
  const qYear = parseInt(year);
  const qQuarter = parseInt(quarter);

  const existing = db.prepare(
    'SELECT * FROM fee_bills WHERE resident_id = ? AND year = ? AND quarter = ?'
  ).get(residentId, qYear, qQuarter) as FeeBill | undefined;

  if (existing) {
    return res.status(409).json({ error: '该季度缴费单已存在' });
  }

  const dueDate = getQuarterDueDate(qYear, qQuarter);
  const now = Date.now();

  const result = db.prepare(`
    INSERT INTO fee_bills 
    (resident_id, year, quarter, property_fee, water_fee, electricity_fee, gas_fee, other_fee, late_fee, is_paid, due_date, created_at)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 0, ?, ?)
  `).run(
    residentId,
    qYear,
    qQuarter,
    parseInt(property_fee) || 0,
    parseInt(water_fee) || 0,
    parseInt(electricity_fee) || 0,
    parseInt(gas_fee) || 0,
    parseInt(other_fee) || 0,
    0,
    dueDate,
    now
  );

  const bill = db.prepare('SELECT * FROM fee_bills WHERE id = ?').get(result.lastInsertRowid) as FeeBill;
  res.status(201).json(formatBill(bill));
});

router.get('/', (req: Request, res: Response) => {
  const { resident_id } = req.query;
  const now = Date.now();

  let bills: FeeBill[];
  if (resident_id) {
    bills = db.prepare(
      'SELECT * FROM fee_bills WHERE resident_id = ? ORDER BY year DESC, quarter DESC'
    ).all(parseInt(resident_id as string)) as FeeBill[];
  } else {
    bills = db.prepare(
      'SELECT * FROM fee_bills ORDER BY year DESC, quarter DESC'
    ).all() as FeeBill[];
  }

  const updatedBills = bills.map(bill => {
    if (!bill.is_paid && now > bill.due_date) {
      const newLateFee = calculateLateFee(bill, now);
      if (newLateFee !== bill.late_fee) {
        db.prepare('UPDATE fee_bills SET late_fee = ? WHERE id = ?').run(newLateFee, bill.id);
        return { ...bill, late_fee: newLateFee };
      }
    }
    return bill;
  });

  res.json(updatedBills.map(formatBill));
});

router.get('/:id', (req: Request, res: Response) => {
  const id = parseInt(req.params.id);
  if (isNaN(id)) {
    return res.status(404).json({ error: '缴费单不存在' });
  }

  let bill = db.prepare('SELECT * FROM fee_bills WHERE id = ?').get(id) as FeeBill | undefined;
  if (!bill) {
    return res.status(404).json({ error: '缴费单不存在' });
  }

  const now = Date.now();
  if (!bill.is_paid && now > bill.due_date) {
    const newLateFee = calculateLateFee(bill, now);
    if (newLateFee !== bill.late_fee) {
      db.prepare('UPDATE fee_bills SET late_fee = ? WHERE id = ?').run(newLateFee, bill.id);
      bill = { ...bill, late_fee: newLateFee };
    }
  }

  res.json(formatBill(bill));
});

router.post('/:id/pay', (req: Request, res: Response) => {
  const id = parseInt(req.params.id);
  if (isNaN(id)) {
    return res.status(404).json({ error: '缴费单不存在' });
  }

  let bill = db.prepare('SELECT * FROM fee_bills WHERE id = ?').get(id) as FeeBill | undefined;
  if (!bill) {
    return res.status(404).json({ error: '缴费单不存在' });
  }

  if (bill.is_paid) {
    return res.status(409).json({ error: '缴费单已支付' });
  }

  const { payment_method } = req.body;
  const paymentMethod = payment_method || '在线支付';

  const now = Date.now();
  const lateFee = calculateLateFee(bill, now);

  if (lateFee !== bill.late_fee) {
    db.prepare('UPDATE fee_bills SET late_fee = ? WHERE id = ?').run(lateFee, bill.id);
    bill = { ...bill, late_fee: lateFee };
  }

  const totalAmount = calculateTotalFee(bill);

  const tx = db.transaction(() => {
    db.prepare('UPDATE fee_bills SET is_paid = 1, paid_at = ? WHERE id = ?').run(now, id);
    db.prepare(`
      INSERT INTO receipts (bill_id, amount, payment_method, created_at)
      VALUES (?, ?, ?, ?)
    `).run(id, totalAmount, paymentMethod, now);
  });
  tx();

  const updatedBill = db.prepare('SELECT * FROM fee_bills WHERE id = ?').get(id) as FeeBill;
  const receipt = db.prepare('SELECT * FROM receipts WHERE bill_id = ? ORDER BY id DESC LIMIT 1').get(id);

  res.json({
    bill: formatBill(updatedBill),
    receipt: {
      id: (receipt as any).id,
      bill_id: (receipt as any).bill_id,
      amount: (receipt as any).amount,
      amount_yuan: ((receipt as any).amount / 100).toFixed(2),
      payment_method: (receipt as any).payment_method,
      created_at: (receipt as any).created_at
    }
  });
});

router.get('/:id/receipts', (req: Request, res: Response) => {
  const id = parseInt(req.params.id);
  if (isNaN(id)) {
    return res.status(404).json({ error: '缴费单不存在' });
  }

  const bill = db.prepare('SELECT * FROM fee_bills WHERE id = ?').get(id) as FeeBill | undefined;
  if (!bill) {
    return res.status(404).json({ error: '缴费单不存在' });
  }

  const receipts = db.prepare('SELECT * FROM receipts WHERE bill_id = ? ORDER BY created_at DESC').all(id);
  res.json(receipts.map((r: any) => ({
    id: r.id,
    bill_id: r.bill_id,
    amount: r.amount,
    amount_yuan: (r.amount / 100).toFixed(2),
    payment_method: r.payment_method,
    created_at: r.created_at
  })));
});

function formatBill(bill: FeeBill) {
  const total = calculateTotalFee(bill);
  return {
    id: bill.id,
    resident_id: bill.resident_id,
    year: bill.year,
    quarter: bill.quarter,
    fees: {
      property_fee: bill.property_fee,
      water_fee: bill.water_fee,
      electricity_fee: bill.electricity_fee,
      gas_fee: bill.gas_fee,
      other_fee: bill.other_fee,
      late_fee: bill.late_fee
    },
    fees_yuan: {
      property_fee: (bill.property_fee / 100).toFixed(2),
      water_fee: (bill.water_fee / 100).toFixed(2),
      electricity_fee: (bill.electricity_fee / 100).toFixed(2),
      gas_fee: (bill.gas_fee / 100).toFixed(2),
      other_fee: (bill.other_fee / 100).toFixed(2),
      late_fee: (bill.late_fee / 100).toFixed(2)
    },
    total_fee: total,
    total_fee_yuan: (total / 100).toFixed(2),
    is_paid: !!bill.is_paid,
    paid_at: bill.paid_at,
    due_date: bill.due_date,
    created_at: bill.created_at
  };
}

router.get('/residents/list', (_req: Request, res: Response) => {
  const residents = db.prepare('SELECT * FROM residents').all();
  res.json(residents);
});

export default router;
