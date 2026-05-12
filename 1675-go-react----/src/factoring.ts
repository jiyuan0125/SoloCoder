import db from './db';
import { formatDate, isDateValid, isRateValid, calculateTotalInterest } from './utils';

export interface Receivable {
  id: number;
  supplier_id: string;
  buyer_name: string;
  amount: number;
  due_date: string;
  contract_no: string;
  created_at: string;
}

export interface FactoringFinance {
  id: number;
  receivable_id: number;
  financing_rate: number;
  loan_amount: number;
  annual_interest_rate: number;
  penalty_rate: number;
  status: 'pending' | 'funded' | 'settled';
  created_at: string;
  funded_at?: string;
  settled_at?: string;
  buyer_paid_at?: string;
  interest_amount?: number;
  settlement_amount?: number;
}

export function registerReceivable(data: {
  supplier_id: string;
  buyer_name: string;
  amount: number;
  due_date: string;
  contract_no: string;
}): { code: number; message: string; data?: Receivable } {
  if (!data.buyer_name || !data.contract_no) {
    return { code: 400, message: '买家名称或合同编号不能为空' };
  }

  if (!data.amount || data.amount <= 0) {
    return { code: 400, message: '应收金额必须为正数' };
  }

  if (!isDateValid(data.due_date)) {
    return { code: 400, message: '到期日格式无效' };
  }

  const stmt = db.prepare(`
    INSERT INTO accounts_receivable (supplier_id, buyer_name, amount, due_date, contract_no, created_at)
    VALUES (?, ?, ?, ?, ?, ?)
  `);

  const result = stmt.run(
    data.supplier_id,
    data.buyer_name,
    data.amount,
    data.due_date,
    data.contract_no,
    formatDate()
  );

  const receivable = db.prepare('SELECT * FROM accounts_receivable WHERE id = ?').get(result.lastInsertRowid) as Receivable;

  return { code: 201, message: '应收账款登记成功', data: receivable };
}

export function getReceivableById(id: number): Receivable | undefined {
  return db.prepare('SELECT * FROM accounts_receivable WHERE id = ?').get(id) as Receivable | undefined;
}

export function listReceivables(supplierId?: string): Receivable[] {
  if (supplierId) {
    return db.prepare('SELECT * FROM accounts_receivable WHERE supplier_id = ?').all(supplierId) as Receivable[];
  }
  return db.prepare('SELECT * FROM accounts_receivable').all() as Receivable[];
}

export function submitFinancingApplication(data: {
  receivable_id: number;
  financing_rate: number;
  annual_interest_rate?: number;
  penalty_rate?: number;
}): { code: number; message: string; data?: FactoringFinance } {
  const receivable = getReceivableById(data.receivable_id);
  if (!receivable) {
    return { code: 404, message: '应收账款不存在' };
  }

  const existingFinance = db.prepare(
    'SELECT * FROM factoring_finance WHERE receivable_id = ?'
  ).get(data.receivable_id);

  if (existingFinance) {
    return { code: 409, message: '该应收账款已存在融资申请' };
  }

  if (!isRateValid(data.financing_rate, 0, 1)) {
    return { code: 400, message: '融资比例必须在0到1之间' };
  }

  const loanAmount = Math.round(receivable.amount * data.financing_rate);

  const stmt = db.prepare(`
    INSERT INTO factoring_finance (receivable_id, financing_rate, loan_amount, annual_interest_rate, penalty_rate, status, created_at)
    VALUES (?, ?, ?, ?, ?, 'pending', ?)
  `);

  const result = stmt.run(
    data.receivable_id,
    data.financing_rate,
    loanAmount,
    data.annual_interest_rate || 0.08,
    data.penalty_rate || 0.15,
    formatDate()
  );

  const finance = db.prepare('SELECT * FROM factoring_finance WHERE id = ?').get(result.lastInsertRowid) as FactoringFinance;

  return { code: 201, message: '融资申请提交成功', data: finance };
}

export function getFinancingById(id: number): FactoringFinance | undefined {
  return db.prepare('SELECT * FROM factoring_finance WHERE id = ?').get(id) as FactoringFinance | undefined;
}

export function listFinancings(status?: string): FactoringFinance[] {
  if (status) {
    return db.prepare('SELECT * FROM factoring_finance WHERE status = ?').all(status) as FactoringFinance[];
  }
  return db.prepare('SELECT * FROM factoring_finance').all() as FactoringFinance[];
}

export function approveAndFund(id: number): { code: number; message: string; data?: FactoringFinance } {
  const finance = getFinancingById(id);
  if (!finance) {
    return { code: 404, message: '融资申请不存在' };
  }

  if (finance.status !== 'pending') {
    return { code: 400, message: '非法状态跳转，只能从待审核状态放款' };
  }

  const stmt = db.prepare(`
    UPDATE factoring_finance SET status = 'funded', funded_at = ? WHERE id = ?
  `);

  stmt.run(formatDate(), id);

  const updated = getFinancingById(id);
  return { code: 200, message: '融资已放款', data: updated };
}

export function settleFinancing(id: number, paymentDate?: string): { code: number; message: string; data?: FactoringFinance } {
  const finance = getFinancingById(id);
  if (!finance) {
    return { code: 404, message: '融资申请不存在' };
  }

  if (finance.status !== 'funded') {
    return { code: 400, message: '非法状态跳转，只能从已放款状态结算' };
  }

  if (!finance.funded_at) {
    return { code: 400, message: '融资尚未放款' };
  }

  const receivable = getReceivableById(finance.receivable_id);
  if (!receivable) {
    return { code: 404, message: '应收账款不存在' };
  }

  const settleDate = paymentDate || formatDate();
  const interestAmount = calculateTotalInterest(
    finance.loan_amount,
    finance.annual_interest_rate,
    finance.penalty_rate,
    finance.funded_at,
    receivable.due_date,
    settleDate
  );

  const settlementAmount = receivable.amount - finance.loan_amount - interestAmount;

  const transaction = db.transaction(() => {
    const stmt = db.prepare(`
      UPDATE factoring_finance 
      SET status = 'settled', 
          settled_at = ?, 
          buyer_paid_at = ?, 
          interest_amount = ?, 
          settlement_amount = ?
      WHERE id = ?
    `);
    stmt.run(formatDate(), settleDate, interestAmount, settlementAmount, id);
  });

  transaction();

  const updated = getFinancingById(id);
  return { code: 200, message: '结算完成', data: updated };
}
