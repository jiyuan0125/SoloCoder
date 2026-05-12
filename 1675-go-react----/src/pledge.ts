import db from './db';
import { formatDate, isRateValid } from './utils';

export interface WarehouseReceipt {
  id: number;
  supplier_id: string;
  product_name: string;
  quantity: number;
  unit_price: number;
  warehouse_address: string;
  market_value: number;
  created_at: string;
}

export interface PledgeFinance {
  id: number;
  receipt_id: number;
  pledge_rate: number;
  loan_amount: number;
  margin_threshold: number;
  status: 'pending' | 'active' | 'closed';
  margin_call_count: number;
  created_at: string;
}

export function registerWarehouseReceipt(data: {
  supplier_id: string;
  product_name: string;
  quantity: number;
  unit_price: number;
  warehouse_address: string;
}): { code: number; message: string; data?: WarehouseReceipt } {
  if (!data.product_name) {
    return { code: 400, message: '商品名称不能为空' };
  }

  if (!data.quantity || data.quantity <= 0) {
    return { code: 400, message: '数量必须为正数' };
  }

  if (!data.unit_price || data.unit_price <= 0) {
    return { code: 400, message: '单价必须为正数' };
  }

  if (!data.warehouse_address) {
    return { code: 400, message: '仓库地址不能为空' };
  }

  const marketValue = Math.round(data.quantity * data.unit_price);

  const stmt = db.prepare(`
    INSERT INTO warehouse_receipts (supplier_id, product_name, quantity, unit_price, warehouse_address, market_value, created_at)
    VALUES (?, ?, ?, ?, ?, ?, ?)
  `);

  const result = stmt.run(
    data.supplier_id,
    data.product_name,
    data.quantity,
    data.unit_price,
    data.warehouse_address,
    marketValue,
    formatDate()
  );

  const receipt = db.prepare('SELECT * FROM warehouse_receipts WHERE id = ?').get(result.lastInsertRowid) as WarehouseReceipt;

  return { code: 201, message: '仓单登记成功', data: receipt };
}

export function getReceiptById(id: number): WarehouseReceipt | undefined {
  return db.prepare('SELECT * FROM warehouse_receipts WHERE id = ?').get(id) as WarehouseReceipt | undefined;
}

export function listReceipts(supplierId?: string): WarehouseReceipt[] {
  if (supplierId) {
    return db.prepare('SELECT * FROM warehouse_receipts WHERE supplier_id = ?').all(supplierId) as WarehouseReceipt[];
  }
  return db.prepare('SELECT * FROM warehouse_receipts').all() as WarehouseReceipt[];
}

export function updateReceiptMarketValue(id: number, newUnitPrice: number): { code: number; message: string; data?: WarehouseReceipt } {
  const receipt = getReceiptById(id);
  if (!receipt) {
    return { code: 404, message: '仓单不存在' };
  }

  if (!newUnitPrice || newUnitPrice <= 0) {
    return { code: 400, message: '单价必须为正数' };
  }

  const newMarketValue = Math.round(receipt.quantity * newUnitPrice);

  const stmt = db.prepare(`
    UPDATE warehouse_receipts SET unit_price = ?, market_value = ? WHERE id = ?
  `);

  stmt.run(newUnitPrice, newMarketValue, id);

  const updated = getReceiptById(id);

  const activePledges = db.prepare(
    'SELECT id FROM pledge_finance WHERE receipt_id = ? AND status = \'active\''
  ).all(id) as { id: number }[];

  for (const pledge of activePledges) {
    if (updated) {
      marginCheck(pledge.id, updated.market_value);
    }
  }

  return { code: 200, message: '市值更新成功', data: updated };
}

export function createPledge(data: {
  receipt_id: number;
  pledge_rate: number;
  margin_threshold?: number;
}): { code: number; message: string; data?: PledgeFinance } {
  const receipt = getReceiptById(data.receipt_id);
  if (!receipt) {
    return { code: 404, message: '仓单不存在' };
  }

  if (!isRateValid(data.pledge_rate, 0, 1)) {
    return { code: 400, message: '质押率必须在0到1之间' };
  }

  const existingPledge = db.prepare(
    'SELECT * FROM pledge_finance WHERE receipt_id = ? AND status IN (\'pending\', \'active\')'
  ).get(data.receipt_id);

  if (existingPledge) {
    return { code: 409, message: '该仓单已有质押中' };
  }

  const loanAmount = Math.round(receipt.market_value * data.pledge_rate);

  const stmt = db.prepare(`
    INSERT INTO pledge_finance (receipt_id, pledge_rate, loan_amount, margin_threshold, status, margin_call_count, created_at)
    VALUES (?, ?, ?, ?, 'active', 0, ?)
  `);

  const result = stmt.run(
    data.receipt_id,
    data.pledge_rate,
    loanAmount,
    data.margin_threshold || 0.6,
    formatDate()
  );

  const pledge = db.prepare('SELECT * FROM pledge_finance WHERE id = ?').get(result.lastInsertRowid) as PledgeFinance;

  return { code: 201, message: '质押融资创建成功', data: pledge };
}

export function getPledgeById(id: number): PledgeFinance | undefined {
  return db.prepare('SELECT * FROM pledge_finance WHERE id = ?').get(id) as PledgeFinance | undefined;
}

export function listPledges(status?: string): PledgeFinance[] {
  if (status) {
    return db.prepare('SELECT * FROM pledge_finance WHERE status = ?').all(status) as PledgeFinance[];
  }
  return db.prepare('SELECT * FROM pledge_finance').all() as PledgeFinance[];
}

export function marginCheck(pledgeId: number, currentMarketValue?: number): { code: number; message: string; triggered: boolean } {
  const pledge = getPledgeById(pledgeId);
  if (!pledge) {
    return { code: 404, message: '质押不存在', triggered: false };
  }

  const receipt = getReceiptById(pledge.receipt_id);
  if (!receipt) {
    return { code: 404, message: '仓单不存在', triggered: false };
  }

  const marketValue = currentMarketValue ?? receipt.market_value;
  const currentPledgeRate = pledge.loan_amount / marketValue;

  if (currentPledgeRate > pledge.margin_threshold) {
    const stmt = db.prepare(
      'UPDATE pledge_finance SET margin_call_count = margin_call_count + 1 WHERE id = ?'
    );
    stmt.run(pledgeId);
    return { code: 200, message: '追加保证金通知已触发', triggered: true };
  }

  return { code: 200, message: '质押率正常', triggered: false };
}

export function closePledge(id: number): { code: number; message: string; data?: PledgeFinance } {
  const pledge = getPledgeById(id);
  if (!pledge) {
    return { code: 404, message: '质押不存在' };
  }

  const stmt = db.prepare(`
    UPDATE pledge_finance SET status = 'closed' WHERE id = ?
  `);

  stmt.run(id);

  const updated = getPledgeById(id);
  return { code: 200, message: '质押已结清', data: updated };
}
