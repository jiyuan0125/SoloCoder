import { getDb, runInTransaction } from '../database';
import { HealthWarning, DoctorNotification, Elder, CareRecord } from '../types';
import { getCurrentDateTime } from '../utils/date';

export class NotificationService {
  private db = getDb();

  notifyDoctor(warningId: number, doctorName: string): { warning: HealthWarning; notification: DoctorNotification } {
    const warning = this.db.prepare('SELECT * FROM healthWarnings WHERE id = ?').get(warningId) as HealthWarning | undefined;
    
    if (!warning) {
      throw new Error('NOT_FOUND: 预警ID不存在');
    }

    const now = getCurrentDateTime();

    return runInTransaction(() => {
      const elder = this.db.prepare('SELECT * FROM elders WHERE id = ?').get(warning.elderId) as Elder | undefined;
      const records = this.db.prepare(`
        SELECT * FROM careRecords 
        WHERE elderId = ? 
        ORDER BY recordDate DESC 
        LIMIT 5
      `).all(warning.elderId) as CareRecord[];

      const notificationStmt = this.db.prepare(`
        INSERT INTO doctorNotifications (warningId, doctorName, notifiedAt)
        VALUES (?, ?, ?)
      `);

      const result = notificationStmt.run(warningId, doctorName, now);

      this.db.prepare(`
        UPDATE healthWarnings 
        SET status = 'notified', notifiedAt = ?, notifiedTo = ?
        WHERE id = ?
      `).run(now, doctorName, warningId);

      const updatedWarning = this.db.prepare('SELECT * FROM healthWarnings WHERE id = ?').get(warningId) as HealthWarning;
      const notification = this.db.prepare('SELECT * FROM doctorNotifications WHERE id = ?').get(Number(result.lastInsertRowid)) as DoctorNotification;

      return { warning: updatedWarning, notification };
    });
  }

  getDoctorNotifications(warningId?: number): DoctorNotification[] {
    let query = 'SELECT * FROM doctorNotifications WHERE 1=1';
    const params: (string | number)[] = [];

    if (warningId) {
      query += ' AND warningId = ?';
      params.push(warningId);
    }

    query += ' ORDER BY notifiedAt DESC';

    return this.db.prepare(query).all(...params) as DoctorNotification[];
  }

  resolveWarning(warningId: number): HealthWarning | undefined {
    const warning = this.db.prepare('SELECT * FROM healthWarnings WHERE id = ?').get(warningId) as HealthWarning | undefined;
    if (!warning) {
      throw new Error('NOT_FOUND: 预警ID不存在');
    }

    this.db.prepare(`
      UPDATE healthWarnings 
      SET status = 'resolved'
      WHERE id = ?
    `).run(warningId);

    return this.db.prepare('SELECT * FROM healthWarnings WHERE id = ?').get(warningId) as HealthWarning;
  }

  getWarningSnapshot(warningId: number): unknown {
    const warning = this.db.prepare('SELECT * FROM healthWarnings WHERE id = ?').get(warningId) as HealthWarning | undefined;
    
    if (!warning) {
      throw new Error('NOT_FOUND: 预警ID不存在');
    }

    try {
      return JSON.parse(warning.snapshot);
    } catch {
      return warning.snapshot;
    }
  }
}

export const notificationService = new NotificationService();
