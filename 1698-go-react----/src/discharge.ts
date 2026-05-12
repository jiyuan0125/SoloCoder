import { Database } from 'sqlite';
import { DischargeNotification, HealthRecord, IsolationRecord } from './types';
import { generateId, today } from './utils';
import { getIsolationById, getLatestHealthRecordByIsolation, updateIsolationStatus } from './isolation';

export interface ConfirmNotificationRequest {
  doctor_name: string;
}

export async function generateDischargeNotification(
  db: Database,
  isolationId: string
): Promise<DischargeNotification> {
  let transactionActive = false;

  try {
    await db.run('BEGIN TRANSACTION');
    transactionActive = true;

    const isolation = await getIsolationById(db, isolationId);

    if (!isolation) {
      const err = new Error('隔离记录不存在');
      (err as any).status = 404;
      throw err;
    }

    if (isolation.status !== 'active' && isolation.status !== 'pending_discharge') {
      const err = new Error('隔离记录状态不正确，无法生成解除通知');
      (err as any).status = 400;
      throw err;
    }

    const latestHealth = await getLatestHealthRecordByIsolation(db, isolationId);
    const healthStatus = buildHealthStatus(latestHealth);

    const existing = await db.get<DischargeNotification>(
      'SELECT * FROM discharge_notifications WHERE isolation_id = ?',
      isolationId
    );

    if (existing) {
      await db.run('COMMIT');
      transactionActive = false;
      return existing;
    }

    const id = generateId();
    const notificationDate = today();

    await db.run(
      `INSERT INTO discharge_notifications 
       (id, isolation_id, person_id, notification_date, health_status) 
       VALUES (?, ?, ?, ?, ?)`,
      id,
      isolationId,
      isolation.person_id,
      notificationDate,
      healthStatus
    );

    await updateIsolationStatus(db, isolationId, 'pending_discharge');

    await db.run('COMMIT');
    transactionActive = false;

    return db.get<DischargeNotification>('SELECT * FROM discharge_notifications WHERE id = ?', id) as Promise<DischargeNotification>;
  } catch (error) {
    if (transactionActive) {
      await db.run('ROLLBACK');
    }
    throw error;
  }
}

function buildHealthStatus(health: HealthRecord | undefined): string {
  if (!health) {
    return '无健康记录数据';
  }

  const parts: string[] = [];
  parts.push(`体温: ${health.temperature}°C`);

  if (health.symptoms) {
    parts.push(`症状: ${health.symptoms}`);
  }

  parts.push(health.is_normal ? '健康状况: 正常' : '健康状况: 异常');

  return parts.join(' | ');
}

export async function getDischargeNotificationByIsolation(
  db: Database,
  isolationId: string
): Promise<DischargeNotification | undefined> {
  return db.get<DischargeNotification>(
    'SELECT * FROM discharge_notifications WHERE isolation_id = ?',
    isolationId
  );
}

export async function getDischargeNotificationById(
  db: Database,
  id: string
): Promise<DischargeNotification | undefined> {
  return db.get<DischargeNotification>('SELECT * FROM discharge_notifications WHERE id = ?', id);
}

export async function confirmDischargeNotification(
  db: Database,
  notificationId: string,
  req: ConfirmNotificationRequest
): Promise<{ notification: DischargeNotification; isolation: IsolationRecord }> {
  const notification = await getDischargeNotificationById(db, notificationId);

  if (!notification) {
    const err = new Error('解除通知不存在');
    (err as any).status = 404;
    throw err;
  }

  if (notification.doctor_confirmed) {
    return {
      notification,
      isolation: (await getIsolationById(db, notification.isolation_id)) as IsolationRecord,
    };
  }

  let transactionActive = false;

  try {
    await db.run('BEGIN TRANSACTION');
    transactionActive = true;

    const confirmedAt = new Date().toISOString();

    await db.run(
      'UPDATE discharge_notifications SET doctor_confirmed = 1, doctor_name = ?, confirmed_at = ? WHERE id = ?',
      req.doctor_name,
      confirmedAt,
      notificationId
    );

    await updateIsolationStatus(db, notification.isolation_id, 'discharged');

    await db.run('COMMIT');
    transactionActive = false;

    const updatedNotification = await getDischargeNotificationById(db, notificationId);
    const isolation = await getIsolationById(db, notification.isolation_id);

    return {
      notification: updatedNotification as DischargeNotification,
      isolation: isolation as IsolationRecord,
    };
  } catch (error) {
    if (transactionActive) {
      await db.run('ROLLBACK');
    }
    throw error;
  }
}
