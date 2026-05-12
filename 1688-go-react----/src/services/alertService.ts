import { db } from '../database';
import { Alert, AssessmentResult } from '../types';
import { v4 as uuidv4 } from 'uuid';

const FORTY_EIGHT_HOURS = 48 * 60 * 60 * 1000;

export const createAlert = (result: AssessmentResult): Promise<Alert | null> => {
  return new Promise(async (resolve, reject) => {
    if (result.level !== 'moderate' && result.level !== 'severe') {
      resolve(null);
      return;
    }

    try {
      const isUrgent = await checkIfUrgent(result.userId, result.scaleId, result.level);
      
      const alert: Alert = {
        id: uuidv4(),
        assessmentResultId: result.id,
        userId: result.userId,
        scaleId: result.scaleId,
        level: result.level,
        priority: isUrgent ? 'urgent' : 'normal',
        status: 'unprocessed',
        notifiedUsers: [],
        createdAt: Date.now()
      };

      db.run(
        `INSERT INTO alerts (id, assessment_result_id, user_id, scale_id, level, priority, status, notified_users, created_at) 
         VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
        [
          alert.id,
          alert.assessmentResultId,
          alert.userId,
          alert.scaleId,
          alert.level,
          alert.priority,
          alert.status,
          JSON.stringify(alert.notifiedUsers),
          alert.createdAt
        ],
        (err) => {
          if (err) reject(err);
          else resolve(alert);
        }
      );
    } catch (err) {
      reject(err);
    }
  });
};

const checkIfUrgent = (userId: string, scaleId: string, currentLevel: string): Promise<boolean> => {
  return new Promise((resolve, reject) => {
    if (currentLevel === 'severe') {
      resolve(true);
      return;
    }

    db.all(
      `SELECT level FROM alerts 
       WHERE user_id = ? AND scale_id = ? AND status != 'resolved'
       ORDER BY created_at DESC 
       LIMIT 2`,
      [userId, scaleId],
      (err, rows: any[]) => {
        if (err) {
          reject(err);
          return;
        }

        const moderateCount = rows.filter(r => r.level === 'moderate').length;
        resolve(moderateCount >= 1);
      }
    );
  });
};

export const getAlertById = (id: string): Promise<Alert | null> => {
  return new Promise((resolve, reject) => {
    db.get('SELECT * FROM alerts WHERE id = ?', [id], (err, row: any) => {
      if (err) reject(err);
      else if (!row) resolve(null);
      else resolve({
        id: row.id,
        assessmentResultId: row.assessment_result_id,
        userId: row.user_id,
        scaleId: row.scale_id,
        level: row.level,
        priority: row.priority,
        status: row.status,
        notifiedUsers: JSON.parse(row.notified_users || '[]'),
        createdAt: row.created_at,
        processingRecord: row.processing_record,
        resolvedAt: row.resolved_at,
        processedBy: row.processed_by
      });
    });
  });
};

export const updateAlertStatus = (
  alertId: string,
  newStatus: 'processing' | 'resolved',
  processedBy?: string,
  processingRecord?: string
): Promise<Alert | null> => {
  return new Promise(async (resolve, reject) => {
    const alert = await getAlertById(alertId);
    if (!alert) {
      resolve(null);
      return;
    }

    const validTransitions: { [key: string]: string[] } = {
      unprocessed: ['processing'],
      processing: ['resolved'],
      resolved: []
    };

    if (!validTransitions[alert.status].includes(newStatus)) {
      reject(new Error('Invalid status transition'));
      return;
    }

    if (newStatus === 'resolved' && !processingRecord) {
      reject(new Error('Processing record is required for resolution'));
      return;
    }

    const now = Date.now();

    db.run(
      `UPDATE alerts SET status = ?, processed_by = ?, processing_record = ?, resolved_at = ? WHERE id = ?`,
      [
        newStatus,
        processedBy || null,
        processingRecord || null,
        newStatus === 'resolved' ? now : null,
        alertId
      ],
      async (err) => {
        if (err) reject(err);
        else {
          const updatedAlert = await getAlertById(alertId);
          resolve(updatedAlert);
        }
      }
    );
  });
};

export const getRelatedUsers = (userId: string): Promise<string[]> => {
  return new Promise((resolve, reject) => {
    db.get(
      'SELECT supervisor_id, class_teacher_id FROM user_relations WHERE user_id = ?',
      [userId],
      (err, row: any) => {
        if (err) {
          reject(err);
          return;
        }

        const users: string[] = [];
        if (row?.supervisor_id) users.push(row.supervisor_id);
        if (row?.class_teacher_id) users.push(row.class_teacher_id);
        resolve(users);
      }
    );
  });
};

export const addNotifiedUser = (alertId: string, userId: string): Promise<void> => {
  return new Promise(async (resolve, reject) => {
    const alert = await getAlertById(alertId);
    if (!alert) {
      reject(new Error('Alert not found'));
      return;
    }

    if (!alert.notifiedUsers.includes(userId)) {
      alert.notifiedUsers.push(userId);
      db.run(
        'UPDATE alerts SET notified_users = ? WHERE id = ?',
        [JSON.stringify(alert.notifiedUsers), alertId],
        (err) => {
          if (err) reject(err);
          else resolve();
        }
      );
    } else {
      resolve();
    }
  });
};
