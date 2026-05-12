import { getDb, runInTransaction } from '../database';
import { FamilyMember, FamilyMessage, VisitAppointment, VisitAppointmentRequest } from '../types';
import { getCurrentDate, getCurrentDateTime, isValidTimeSlot } from '../utils/date';

export class FamilyService {
  private db = getDb();

  addFamilyMember(elderId: number, name: string, phone: string, relation: string): FamilyMember {
    const now = getCurrentDateTime();
    
    const result = this.db.prepare(`
      INSERT INTO familyMembers (elderId, name, phone, relation, createdAt)
      VALUES (?, ?, ?, ?, ?)
    `).run(elderId, name, phone, relation, now);

    return this.db.prepare('SELECT * FROM familyMembers WHERE id = ?').get(Number(result.lastInsertRowid)) as FamilyMember;
  }

  getFamilyMembers(elderId: number): FamilyMember[] {
    return this.db.prepare(`
      SELECT * FROM familyMembers 
      WHERE elderId = ? 
      ORDER BY createdAt DESC
    `).all(elderId) as FamilyMember[];
  }

  updateLastView(familyMemberId: number): FamilyMember | undefined {
    const now = getCurrentDateTime();
    const currentDate = getCurrentDate();

    const member = this.db.prepare('SELECT * FROM familyMembers WHERE id = ?').get(familyMemberId) as FamilyMember | undefined;
    if (!member) {
      return undefined;
    }

    this.db.prepare(`
      UPDATE familyMembers 
      SET lastViewDate = ?, viewCount = viewCount + 1
      WHERE id = ?
    `).run(currentDate, familyMemberId);

    return this.db.prepare('SELECT * FROM familyMembers WHERE id = ?').get(familyMemberId) as FamilyMember;
  }

  sendMessage(familyMemberId: number, elderId: number, content: string): FamilyMessage {
    const now = getCurrentDateTime();

    const result = this.db.prepare(`
      INSERT INTO familyMessages (familyMemberId, elderId, content, createdAt)
      VALUES (?, ?, ?, ?)
    `).run(familyMemberId, elderId, content, now);

    return this.db.prepare('SELECT * FROM familyMessages WHERE id = ?').get(Number(result.lastInsertRowid)) as FamilyMessage;
  }

  getMessages(elderId: number, familyMemberId?: number): FamilyMessage[] {
    let query = 'SELECT * FROM familyMessages WHERE elderId = ?';
    const params: (string | number)[] = [elderId];

    if (familyMemberId) {
      query += ' AND familyMemberId = ?';
      params.push(familyMemberId);
    }

    query += ' ORDER BY createdAt DESC';

    return this.db.prepare(query).all(...params) as FamilyMessage[];
  }

  createAppointment(request: VisitAppointmentRequest): VisitAppointment {
    const timeSlot = request.timeSlot;
    if (!isValidTimeSlot(timeSlot)) {
      throw new Error('BAD_REQUEST: 无效的探视时段，必须是morning或afternoon');
    }

    const existing = this.db.prepare(`
      SELECT * FROM visitAppointments 
      WHERE elderId = ? AND visitDate = ? AND timeSlot = ? AND status = 'pending'
    `).get(request.elderId, request.visitDate, timeSlot) as VisitAppointment | undefined;

    if (existing) {
      throw new Error('CONFLICT: 该时段已经预约过了');
    }

    const now = getCurrentDateTime();

    return runInTransaction(() => {
      const result = this.db.prepare(`
        INSERT INTO visitAppointments 
          (familyMemberId, elderId, visitDate, timeSlot, visitorName, visitorPhone, status, createdAt)
        VALUES (?, ?, ?, ?, ?, ?, 'pending', ?)
      `).run(
        request.familyMemberId,
        request.elderId,
        request.visitDate,
        timeSlot,
        request.visitorName,
        request.visitorPhone || null,
        now
      );

      return this.db.prepare('SELECT * FROM visitAppointments WHERE id = ?').get(Number(result.lastInsertRowid)) as VisitAppointment;
    });
  }

  getAppointments(elderId?: number, familyMemberId?: number, date?: string): VisitAppointment[] {
    let query = 'SELECT * FROM visitAppointments WHERE 1=1';
    const params: (string | number)[] = [];

    if (elderId) {
      query += ' AND elderId = ?';
      params.push(elderId);
    }
    if (familyMemberId) {
      query += ' AND familyMemberId = ?';
      params.push(familyMemberId);
    }
    if (date) {
      query += ' AND visitDate = ?';
      params.push(date);
    }

    query += ' ORDER BY visitDate DESC, timeSlot ASC';

    return this.db.prepare(query).all(...params) as VisitAppointment[];
  }

  cancelAppointment(appointmentId: number): VisitAppointment | undefined {
    const appointment = this.db.prepare('SELECT * FROM visitAppointments WHERE id = ?').get(appointmentId) as VisitAppointment | undefined;
    if (!appointment) {
      return undefined;
    }

    this.db.prepare(`
      UPDATE visitAppointments 
      SET status = 'cancelled'
      WHERE id = ?
    `).run(appointmentId);

    return this.db.prepare('SELECT * FROM visitAppointments WHERE id = ?').get(appointmentId) as VisitAppointment;
  }
}

export const familyService = new FamilyService();
