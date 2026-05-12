import { Database } from 'sqlite';
import { HealthRecord, IsolationRecord, IsolationStatus, IsolationType } from './types';
import {
  generateId,
  isValidIsolationType,
  calculateEndDate,
  isHighTemperature,
  hasRespiratorySymptoms,
  today,
} from './utils';

export interface CreateIsolationRequest {
  person_id: string;
  isolation_type: string;
  start_date: string;
}

export interface HealthCheckInRequest {
  person_id: string;
  isolation_id?: string;
  record_date?: string;
  temperature: number;
  symptoms?: string;
}

export async function createIsolation(
  db: Database,
  req: CreateIsolationRequest
): Promise<IsolationRecord> {
  if (!isValidIsolationType(req.isolation_type)) {
    const err = new Error('隔离类型不在有效范围内: home, centralized');
    (err as any).status = 400;
    throw err;
  }

  const startDate = req.start_date || today();
  const endDate = calculateEndDate(startDate, 14);
  const id = generateId();

  await db.run(
    'INSERT INTO isolation_records (id, person_id, isolation_type, start_date, end_date, status) VALUES (?, ?, ?, ?, ?, ?)',
    id,
    req.person_id,
    req.isolation_type as IsolationType,
    startDate,
    endDate,
    'active'
  );

  return db.get<IsolationRecord>('SELECT * FROM isolation_records WHERE id = ?', id) as Promise<IsolationRecord>;
}

export async function getIsolationById(
  db: Database,
  id: string
): Promise<IsolationRecord | undefined> {
  return db.get<IsolationRecord>('SELECT * FROM isolation_records WHERE id = ?', id);
}

export async function getActiveIsolationByPersonId(
  db: Database,
  personId: string
): Promise<IsolationRecord | undefined> {
  return db.get<IsolationRecord>(
    'SELECT * FROM isolation_records WHERE person_id = ? AND status IN (?, ?) ORDER BY start_date DESC LIMIT 1',
    personId,
    'active',
    'pending_discharge'
  );
}

export async function createHealthRecord(
  db: Database,
  req: HealthCheckInRequest
): Promise<{ healthRecord: HealthRecord; needsHospital: boolean }> {
  const recordDate = req.record_date || today();
  const isNormal = !isHighTemperature(req.temperature) && !hasRespiratorySymptoms(req.symptoms || '');
  const needsHospital = !isNormal;

  let isolationId = req.isolation_id;

  if (!isolationId) {
    const activeIsolation = await getActiveIsolationByPersonId(db, req.person_id);
    if (activeIsolation) {
      isolationId = activeIsolation.id;
    }
  }

  if (needsHospital && isolationId) {
    const isolation = await getIsolationById(db, isolationId);
    if (isolation && isolation.status !== 'transferred_hospital') {
      await db.run(
        'UPDATE isolation_records SET status = ? WHERE id = ?',
        'transferred_hospital',
        isolationId
      );
    }
  }

  const id = generateId();
  await db.run(
    'INSERT INTO health_records (id, person_id, isolation_id, record_date, temperature, symptoms, is_normal) VALUES (?, ?, ?, ?, ?, ?, ?)',
    id,
    req.person_id,
    isolationId || null,
    recordDate,
    req.temperature,
    req.symptoms || '',
    isNormal ? 1 : 0
  );

  const healthRecord = await db.get<HealthRecord>('SELECT * FROM health_records WHERE id = ?', id);

  return { healthRecord: healthRecord as HealthRecord, needsHospital };
}

export async function getLatestHealthRecordByIsolation(
  db: Database,
  isolationId: string
): Promise<HealthRecord | undefined> {
  return db.get<HealthRecord>(
    'SELECT * FROM health_records WHERE isolation_id = ? ORDER BY record_date DESC, created_at DESC LIMIT 1',
    isolationId
  );
}

export async function getHealthRecordsByIsolation(
  db: Database,
  isolationId: string
): Promise<HealthRecord[]> {
  return db.all<HealthRecord[]>(
    'SELECT * FROM health_records WHERE isolation_id = ? ORDER BY record_date DESC',
    isolationId
  );
}

export async function getExpiringIsolations(db: Database): Promise<IsolationRecord[]> {
  const todayStr = today();
  return db.all<IsolationRecord[]>(
    'SELECT * FROM isolation_records WHERE status = ? AND end_date <= ?',
    'active',
    todayStr
  );
}

export async function updateIsolationStatus(
  db: Database,
  isolationId: string,
  status: IsolationStatus
): Promise<void> {
  await db.run('UPDATE isolation_records SET status = ? WHERE id = ?', status, isolationId);
}
