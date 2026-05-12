import { Database } from 'sqlite';
import { TestRecord } from './types';
import { generateId, isValidIsolationTestDay, isValidTestResult, ISOLATION_TEST_DAYS, today } from './utils';
import { createIsolation, getActiveIsolationByPersonId, getIsolationById } from './isolation';

export interface CreateTestRequest {
  person_id: string;
  isolation_id?: string;
  test_day: number;
  test_date?: string;
  result: string;
}

export async function createTestRecord(
  db: Database,
  req: CreateTestRequest
): Promise<{ testRecord: TestRecord; flowInvestigation: boolean; startedIsolation: boolean }> {
  if (!isValidTestResult(req.result)) {
    const err = new Error('检测结果不在有效范围内: positive, negative, pending');
    (err as any).status = 400;
    throw err;
  }

  let isolationId = req.isolation_id;
  let isKeyPerson = false;

  if (isolationId) {
    const isolation = await getIsolationById(db, isolationId);
    if (!isolation) {
      const err = new Error('隔离记录不存在');
      (err as any).status = 404;
      throw err;
    }
  }

  const person = await db.get<any>('SELECT is_key_person FROM persons WHERE id = ?', req.person_id);
  if (person) {
    isKeyPerson = person.is_key_person;
  }

  if (!isValidIsolationTestDay(req.test_day, isKeyPerson)) {
    const err = new Error(`检测天次不在规定范围内。隔离人员: ${ISOLATION_TEST_DAYS.join(', ')}；重点人员: 任意大于等于1的天数`);
    (err as any).status = 400;
    throw err;
  }

  const testDate = req.test_date || today();
  const id = generateId();
  const flowInvestigation = req.result === 'positive';

  let transactionActive = false;

  try {
    await db.run('BEGIN TRANSACTION');
    transactionActive = true;

    await db.run(
      'INSERT INTO test_records (id, person_id, isolation_id, test_day, test_date, result) VALUES (?, ?, ?, ?, ?, ?)',
      id,
      req.person_id,
      isolationId || null,
      req.test_day,
      testDate,
      req.result
    );

    if (req.result === 'positive') {
      const existingIsolation = await getActiveIsolationByPersonId(db, req.person_id);
      if (!existingIsolation) {
        await createIsolation(db, {
          person_id: req.person_id,
          isolation_type: 'centralized',
          start_date: testDate,
        });
      }
    }

    await db.run('COMMIT');
    transactionActive = false;

    const testRecord = await db.get<TestRecord>('SELECT * FROM test_records WHERE id = ?', id);

    return {
      testRecord: testRecord as TestRecord,
      flowInvestigation,
      startedIsolation: req.result === 'positive',
    };
  } catch (error) {
    if (transactionActive) {
      await db.run('ROLLBACK');
    }
    throw error;
  }
}

export async function getTestRecordById(
  db: Database,
  id: string
): Promise<TestRecord | undefined> {
  return db.get<TestRecord>('SELECT * FROM test_records WHERE id = ?', id);
}

export async function getTestRecordsByPersonId(
  db: Database,
  personId: string
): Promise<TestRecord[]> {
  return db.all<TestRecord[]>(
    'SELECT * FROM test_records WHERE person_id = ? ORDER BY test_date DESC',
    personId
  );
}

export async function getTestRecordsByIsolationId(
  db: Database,
  isolationId: string
): Promise<TestRecord[]> {
  return db.all<TestRecord[]>(
    'SELECT * FROM test_records WHERE isolation_id = ? ORDER BY test_day ASC',
    isolationId
  );
}
