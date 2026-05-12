import { getDb, runInTransaction } from '../database';
import { CarePlan, CarePlanChange, CarePlanUpdateRequest, CareRecord, CareRecordRequest, HealthWarning, Elder } from '../types';
import { getCurrentDate, getCurrentDateTime, addMonths, isDateBefore } from '../utils/date';

const TEMPERATURE_THRESHOLD = 37.5;
const CONSECUTIVE_DAYS_FOR_WARNING = 3;

export class CareService {
  private db = getDb();

  getOrCreateCarePlan(elderId: number): CarePlan {
    let plan = this.db.prepare('SELECT * FROM carePlans WHERE elderId = ?').get(elderId) as CarePlan | undefined;
    
    if (!plan) {
      const now = getCurrentDateTime();
      const result = this.db.prepare(`
        INSERT INTO carePlans (elderId, createdAt, updatedAt)
        VALUES (?, ?, ?)
      `).run(elderId, now, now);
      
      plan = this.db.prepare('SELECT * FROM carePlans WHERE id = ?').get(Number(result.lastInsertRowid)) as CarePlan;
    }
    
    return plan;
  }

  updateCarePlan(elderId: number, request: CarePlanUpdateRequest): { plan: CarePlan; change: CarePlanChange } {
    const plan = this.getOrCreateCarePlan(elderId);
    const now = getCurrentDateTime();

    return runInTransaction(() => {
      const snapshot = JSON.stringify({
        medicationReminder: plan.medicationReminder,
        dietArrangement: plan.dietArrangement,
        rehabilitationProject: plan.rehabilitationProject
      });

      const updateStmt = this.db.prepare(`
        UPDATE carePlans 
        SET medicationReminder = ?, dietArrangement = ?, rehabilitationProject = ?, updatedAt = ?
        WHERE id = ?
      `);

      updateStmt.run(
        request.medicationReminder || plan.medicationReminder || null,
        request.dietArrangement || plan.dietArrangement || null,
        request.rehabilitationProject || plan.rehabilitationProject || null,
        now,
        plan.id
      );

      const changeStmt = this.db.prepare(`
        INSERT INTO carePlanChanges (carePlanId, changeReason, approvedBy, changedAt, snapshot)
        VALUES (?, ?, ?, ?, ?)
      `);

      const changeResult = changeStmt.run(
        plan.id,
        request.changeReason,
        request.approvedBy,
        now,
        snapshot
      );

      const updatedPlan = this.db.prepare('SELECT * FROM carePlans WHERE id = ?').get(plan.id) as CarePlan;
      const change = this.db.prepare('SELECT * FROM carePlanChanges WHERE id = ?').get(Number(changeResult.lastInsertRowid)) as CarePlanChange;

      return { plan: updatedPlan, change };
    });
  }

  getCarePlanChanges(carePlanId: number): CarePlanChange[] {
    return this.db.prepare(`
      SELECT * FROM carePlanChanges 
      WHERE carePlanId = ? 
      ORDER BY changedAt DESC
    `).all(carePlanId) as CarePlanChange[];
  }

  createCareRecord(elderId: number, request: CareRecordRequest): { record: CareRecord; warning?: HealthWarning } {
    let temperatureValue: number;
    if (request.temperature === undefined || request.temperature === null) {
      throw new Error('BAD_REQUEST: 体温不能为空');
    }
    if (typeof request.temperature === 'string') {
      if (request.temperature.trim() === '') {
        throw new Error('BAD_REQUEST: 体温不能为空');
      }
      temperatureValue = Number(request.temperature);
      if (Number.isNaN(temperatureValue)) {
        throw new Error('BAD_REQUEST: 体温必须是有效的数字');
      }
    } else {
      temperatureValue = request.temperature;
    }

    if (!request.bloodPressure || request.bloodPressure.trim() === '') {
      throw new Error('BAD_REQUEST: 血压不能为空');
    }

    const recordDate = request.recordDate || getCurrentDate();
    const now = getCurrentDateTime();

    return runInTransaction(() => {
      const insertStmt = this.db.prepare(`
        INSERT INTO careRecords 
          (elderId, recordDate, temperature, bloodPressure, diet, specialNotes, createdBy, createdAt)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
      `);

      const result = insertStmt.run(
        elderId,
        recordDate,
        temperatureValue,
        request.bloodPressure,
        request.diet || null,
        request.specialNotes || null,
        request.createdBy || null,
        now
      );

      const record = this.db.prepare('SELECT * FROM careRecords WHERE id = ?').get(Number(result.lastInsertRowid)) as CareRecord;

      let warning: HealthWarning | undefined;
      if (temperatureValue > TEMPERATURE_THRESHOLD) {
        warning = this.checkAndCreateHealthWarning(elderId, recordDate);
      }

      return { record, warning };
    });
  }

  private checkAndCreateHealthWarning(elderId: number, checkDate: string): HealthWarning | undefined {
    const recentRecords = this.db.prepare(`
      SELECT * FROM careRecords 
      WHERE elderId = ? AND temperature > ?
      ORDER BY recordDate DESC
      LIMIT ?
    `).all(elderId, TEMPERATURE_THRESHOLD, CONSECUTIVE_DAYS_FOR_WARNING) as CareRecord[];

    if (recentRecords.length < CONSECUTIVE_DAYS_FOR_WARNING) {
      return undefined;
    }

    const existingWarning = this.db.prepare(`
      SELECT * FROM healthWarnings 
      WHERE elderId = ? AND warningType = 'fever' AND status = 'pending'
    `).get(elderId) as HealthWarning | undefined;

    if (existingWarning) {
      return undefined;
    }

    const elder = this.db.prepare('SELECT * FROM elders WHERE id = ?').get(elderId) as Elder | undefined;
    if (!elder) {
      return undefined;
    }

    const snapshot = JSON.stringify({
      elder: {
        name: elder.name,
        idCard: elder.idCard,
        careLevel: elder.careLevel,
        emergencyContact: elder.emergencyContact,
        emergencyContactPhone: elder.emergencyContactPhone
      },
      records: recentRecords.map(r => ({
        recordDate: r.recordDate,
        temperature: r.temperature,
        bloodPressure: r.bloodPressure,
        diet: r.diet,
        specialNotes: r.specialNotes
      }))
    });

    const insertStmt = this.db.prepare(`
      INSERT INTO healthWarnings 
        (elderId, warningType, description, snapshot, status, createdAt)
      VALUES (?, 'fever', ?, ?, 'pending', ?)
    `);

    const description = `连续${CONSECUTIVE_DAYS_FOR_WARNING}天体温超过${TEMPERATURE_THRESHOLD}度，最近一次体温: ${recentRecords[0].temperature}度`;
    const result = insertStmt.run(elderId, description, snapshot, getCurrentDateTime());

    return this.db.prepare('SELECT * FROM healthWarnings WHERE id = ?').get(Number(result.lastInsertRowid)) as HealthWarning;
  }

  getCareRecords(elderId: number, startDate?: string, endDate?: string): CareRecord[] {
    let query = 'SELECT * FROM careRecords WHERE elderId = ?';
    const params: (string | number)[] = [elderId];

    if (startDate) {
      query += ' AND recordDate >= ?';
      params.push(startDate);
    }
    if (endDate) {
      query += ' AND recordDate <= ?';
      params.push(endDate);
    }

    query += ' ORDER BY recordDate DESC, createdAt DESC';

    return this.db.prepare(query).all(...params) as CareRecord[];
  }

  getHealthWarnings(elderId?: number, status?: string): HealthWarning[] {
    let query = 'SELECT * FROM healthWarnings WHERE 1=1';
    const params: (string | number)[] = [];

    if (elderId) {
      query += ' AND elderId = ?';
      params.push(elderId);
    }
    if (status) {
      query += ' AND status = ?';
      params.push(status);
    }

    query += ' ORDER BY createdAt DESC';

    return this.db.prepare(query).all(...params) as HealthWarning[];
  }

  getHealthWarning(warningId: number): HealthWarning | undefined {
    return this.db.prepare('SELECT * FROM healthWarnings WHERE id = ?').get(warningId) as HealthWarning | undefined;
  }

  checkFamilyViewReminders(): { familyMemberId: number; elderId: number; elderName: string; lastViewDate: string | null; monthsSinceLastView: number }[] {
    const currentDate = getCurrentDate();
    const twoMonthsAgo = addMonths(currentDate, -2);

    const query = `
      SELECT fm.id as familyMemberId, fm.elderId, e.name as elderName, 
             fm.lastViewDate, fm.viewCount
      FROM familyMembers fm
      JOIN elders e ON fm.elderId = e.id
      WHERE e.status = 'active'
        AND (fm.lastViewDate IS NULL OR fm.lastViewDate < ?)
    `;

    const results = this.db.prepare(query).all(twoMonthsAgo) as {
      familyMemberId: number;
      elderId: number;
      elderName: string;
      lastViewDate: string | null;
      viewCount: number;
    }[];

    return results.map(r => ({
      familyMemberId: r.familyMemberId,
      elderId: r.elderId,
      elderName: r.elderName,
      lastViewDate: r.lastViewDate,
      monthsSinceLastView: r.lastViewDate 
        ? this.calculateMonthsSince(r.lastViewDate, currentDate)
        : 999
    }));
  }

  private calculateMonthsSince(startDate: string, endDate: string): number {
    const start = new Date(startDate);
    const end = new Date(endDate);
    const yearDiff = end.getFullYear() - start.getFullYear();
    const monthDiff = end.getMonth() - start.getMonth();
    return yearDiff * 12 + monthDiff;
  }
}

export const careService = new CareService();
