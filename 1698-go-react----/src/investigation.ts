import { Database } from 'sqlite';
import { Investigation, Person } from './types';
import { generateId, isValidInvestigationResult, hasHighRiskTravel } from './utils';

export interface CreateInvestigationRequest {
  person_id: string;
  investigation_date: string;
  investigation_method: string;
  travel_history: string;
  health_status: string;
  result: string;
  person_info?: {
    name: string;
    id_card: string;
    phone?: string;
    address?: string;
  };
}

export async function createInvestigation(
  db: Database,
  req: CreateInvestigationRequest
): Promise<{ person: Person; investigation: Investigation }> {
  if (!req.investigation_date) {
    const err = new Error('排查日期不能为空');
    (err as any).status = 400;
    throw err;
  }

  if (!isValidInvestigationResult(req.result)) {
    const err = new Error('排查结果不在有效范围内: normal, need_isolation, need_hospital');
    (err as any).status = 400;
    throw err;
  }

  const isKeyPerson = hasHighRiskTravel(req.travel_history);

  let transactionActive = false;

  try {
    await db.run('BEGIN TRANSACTION');
    transactionActive = true;

    let person: Person;

    const existingPerson = await db.get<Person>('SELECT * FROM persons WHERE id = ?', req.person_id);

    if (existingPerson) {
      person = existingPerson;
      if (isKeyPerson && !existingPerson.is_key_person) {
        await db.run('UPDATE persons SET is_key_person = 1 WHERE id = ?', req.person_id);
        person = { ...person, is_key_person: true };
      }
    } else if (req.person_info) {
      const newPersonId = generateId();
      await db.run(
        'INSERT INTO persons (id, name, id_card, phone, address, is_key_person) VALUES (?, ?, ?, ?, ?, ?)',
        newPersonId,
        req.person_info.name,
        req.person_info.id_card,
        req.person_info.phone || null,
        req.person_info.address || null,
        isKeyPerson ? 1 : 0
      );
      person = await db.get<Person>('SELECT * FROM persons WHERE id = ?', newPersonId) as Person;
    } else {
      const err = new Error('未找到人员信息');
      (err as any).status = 404;
      throw err;
    }

    const investigationId = generateId();
    await db.run(
      'INSERT INTO investigations (id, person_id, investigation_date, investigation_method, travel_history, health_status, result) VALUES (?, ?, ?, ?, ?, ?, ?)',
      investigationId,
      person.id,
      req.investigation_date,
      req.investigation_method,
      req.travel_history,
      req.health_status,
      req.result
    );

    await db.run('COMMIT');
    transactionActive = false;

    const investigation = await db.get<Investigation>(
      'SELECT * FROM investigations WHERE id = ?',
      investigationId
    );

    return { person, investigation: investigation as Investigation };
  } catch (error) {
    if (transactionActive) {
      await db.run('ROLLBACK');
    }
    throw error;
  }
}

export async function getInvestigationById(
  db: Database,
  id: string
): Promise<Investigation | undefined> {
  return db.get<Investigation>('SELECT * FROM investigations WHERE id = ?', id);
}

export async function getInvestigationsByPersonId(
  db: Database,
  personId: string
): Promise<Investigation[]> {
  return db.all<Investigation[]>('SELECT * FROM investigations WHERE person_id = ? ORDER BY investigation_date DESC', personId);
}
