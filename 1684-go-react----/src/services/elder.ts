import { getDb, runInTransaction } from '../database';
import { Elder, CareLevel, CheckInRequest, CheckOutRequest, MONTHLY_FEES, VALID_CARE_LEVELS, BillingRecord, OccupancyStats } from '../types';
import { getCurrentDate, getCurrentDateTime, daysBetweenInclusive } from '../utils/date';

export class ElderService {
  private db = getDb();

  checkIn(request: CheckInRequest): { elder: Elder; deposit: BillingRecord } {
    if (!request.idCard || request.idCard.trim() === '') {
      throw new Error('BAD_REQUEST: 身份证号不能为空');
    }

    const careLevel = request.careLevel as CareLevel;
    if (!VALID_CARE_LEVELS.includes(careLevel)) {
      throw new Error('BAD_REQUEST: 护理等级无效，必须是: 自理(self-care)、半护理(semi-care)或全护理(full-care)');
    }

    const checkInDate = request.checkInDate || getCurrentDate();
    const monthlyFee = MONTHLY_FEES[careLevel];
    const depositAmount = monthlyFee * 2;

    return runInTransaction(() => {
      const insertElder = this.db.prepare(`
        INSERT INTO elders 
          (name, idCard, gender, birthDate, emergencyContact, emergencyContactPhone, 
           medicalHistory, careLevel, status, checkInDate, roomNumber, bedNumber, 
           depositAmount, createdAt)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
      `);

      const result = insertElder.run(
        request.name,
        request.idCard,
        request.gender || null,
        request.birthDate || null,
        request.emergencyContact,
        request.emergencyContactPhone,
        request.medicalHistory || null,
        careLevel,
        checkInDate,
        request.roomNumber || null,
        request.bedNumber || null,
        depositAmount,
        getCurrentDateTime()
      );

      const elderId = Number(result.lastInsertRowid);

      const insertBilling = this.db.prepare(`
        INSERT INTO billingRecords 
          (elderId, type, amount, description, createdAt)
        VALUES (?, 'deposit', ?, ?, ?)
      `);

      const billingResult = insertBilling.run(
        elderId,
        depositAmount,
        `入住押金 (两个月: ${monthlyFee}分)`,
        getCurrentDateTime()
      );

      const elder = this.db.prepare('SELECT * FROM elders WHERE id = ?').get(elderId) as Elder;
      const deposit = this.db.prepare('SELECT * FROM billingRecords WHERE id = ?').get(Number(billingResult.lastInsertRowid)) as BillingRecord;

      return { elder, deposit };
    });
  }

  checkOut(elderId: number, request: CheckOutRequest): { elder: Elder; finalBill: BillingRecord[] } {
    const elder = this.getElder(elderId);
    if (!elder) {
      throw new Error('NOT_FOUND: 老人不存在');
    }
    if (elder.status !== 'active') {
      throw new Error('BAD_REQUEST: 该老人已退住');
    }

    const checkOutDate = request.checkOutDate || getCurrentDate();
    
    const unpaidAmount = this.calculateUnpaidAmount(elderId);
    if (unpaidAmount > 0) {
      throw new Error('BAD_REQUEST: 存在未结清费用');
    }

    const stayDays = daysBetweenInclusive(elder.checkInDate, checkOutDate);
    const monthlyFee = MONTHLY_FEES[elder.careLevel as CareLevel];
    const dailyRate = Math.ceil(monthlyFee / 30);
    const totalCharge = dailyRate * stayDays;

    const bills: BillingRecord[] = [];

    return runInTransaction(() => {
      const chargeStmt = this.db.prepare(`
        INSERT INTO billingRecords (elderId, type, amount, description, createdAt)
        VALUES (?, 'charge', ?, ?, ?)
      `);
      const chargeResult = chargeStmt.run(
        elderId,
        totalCharge,
        `退住结算 (住${stayDays}天，每日${dailyRate}分)`,
        getCurrentDateTime()
      );
      bills.push(this.db.prepare('SELECT * FROM billingRecords WHERE id = ?').get(Number(chargeResult.lastInsertRowid)) as BillingRecord);

      const remainingDeposit = this.getCurrentDeposit(elderId);
      let depositAfterCharge = remainingDeposit - totalCharge;

      if (depositAfterCharge > 0) {
        const refundStmt = this.db.prepare(`
          INSERT INTO billingRecords (elderId, type, amount, description, createdAt)
          VALUES (?, 'refund', ?, ?, ?)
        `);
        const refundResult = refundStmt.run(
          elderId,
          depositAfterCharge,
          '押金退还',
          getCurrentDateTime()
        );
        bills.push(this.db.prepare('SELECT * FROM billingRecords WHERE id = ?').get(Number(refundResult.lastInsertRowid)) as BillingRecord);
        depositAfterCharge = 0;
      } else if (depositAfterCharge < 0) {
        const arrearsStmt = this.db.prepare(`
          INSERT INTO billingRecords (elderId, type, amount, description, createdAt)
          VALUES (?, 'arrears', ?, ?, ?)
        `);
        const arrearsResult = arrearsStmt.run(
          elderId,
          Math.abs(depositAfterCharge),
          '欠费记录',
          getCurrentDateTime()
        );
        bills.push(this.db.prepare('SELECT * FROM billingRecords WHERE id = ?').get(Number(arrearsResult.lastInsertRowid)) as BillingRecord);
        depositAfterCharge = 0;
      }

      const updateElder = this.db.prepare(`
        UPDATE elders 
        SET status = 'discharged', checkOutDate = ?, depositAmount = 0
        WHERE id = ?
      `);
      updateElder.run(checkOutDate, elderId);

      const updatedElder = this.getElder(elderId)!;
      return { elder: updatedElder, finalBill: bills };
    });
  }

  getElder(elderId: number): Elder | undefined {
    return this.db.prepare('SELECT * FROM elders WHERE id = ?').get(elderId) as Elder | undefined;
  }

  getElders(status?: string): Elder[] {
    if (status) {
      return this.db.prepare('SELECT * FROM elders WHERE status = ?').all(status) as Elder[];
    }
    return this.db.prepare('SELECT * FROM elders').all() as Elder[];
  }

  getCurrentDeposit(elderId: number): number {
    const records = this.db.prepare(`
      SELECT type, amount FROM billingRecords 
      WHERE elderId = ?
    `).all(elderId) as { type: string; amount: number }[];

    let balance = 0;
    for (const record of records) {
      if (record.type === 'deposit' || record.type === 'refund') {
        balance += record.amount;
      } else if (record.type === 'charge' || record.type === 'arrears') {
        balance -= record.amount;
      }
    }
    return balance;
  }

  calculateUnpaidAmount(elderId: number): number {
    const records = this.db.prepare(`
      SELECT type, amount FROM billingRecords 
      WHERE elderId = ?
    `).all(elderId) as { type: string; amount: number }[];

    let balance = 0;
    for (const record of records) {
      if (record.type === 'arrears') {
        balance += record.amount;
      }
    }
    return balance;
  }

  getOccupancyStats(): OccupancyStats {
    const totalBedsResult = this.db.prepare(`
      SELECT value FROM settings WHERE key = 'totalBeds'
    `).get() as { value: string } | undefined;

    const totalBeds = totalBedsResult ? parseInt(totalBedsResult.value, 10) : 100;

    const occupiedResult = this.db.prepare(`
      SELECT COUNT(*) as count FROM elders WHERE status = 'active'
    `).get() as { count: number };

    const occupiedBeds = occupiedResult.count;
    const occupancyRate = totalBeds > 0 ? (occupiedBeds / totalBeds) * 100 : 0;

    return {
      totalBeds,
      occupiedBeds,
      occupancyRate
    };
  }

  getBillingRecords(elderId: number): BillingRecord[] {
    return this.db.prepare(`
      SELECT * FROM billingRecords 
      WHERE elderId = ? 
      ORDER BY createdAt DESC
    `).all(elderId) as BillingRecord[];
  }
}

export const elderService = new ElderService();
