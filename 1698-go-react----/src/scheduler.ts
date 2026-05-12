import { Database } from 'sqlite';
import cron from 'node-cron';
import { yesterday, today, calculateIsolationDay } from './utils';
import { getExpiringIsolations } from './isolation';
import { generateDischargeNotification } from './discharge';

export interface DailySummary {
  date: string;
  new_investigations: number;
  new_isolations: number;
  new_tests: number;
  new_health_records: number;
  key_persons: number;
  active_isolations: number;
  pending_discharge: number;
}

export interface TestSchedule {
  person_id: string;
  person_name: string;
  isolation_id?: string;
  test_day: number;
  reason: string;
}

export async function generateDailySummary(db: Database): Promise<DailySummary> {
  const yesterdayStr = yesterday();
  const todayStr = today();

  const [
    newInvestigations,
    newIsolations,
    newTests,
    newHealthRecords,
    keyPersons,
    activeIsolations,
    pendingDischarge,
  ] = await Promise.all([
    db.get<{ count: number }>('SELECT COUNT(*) as count FROM investigations WHERE DATE(investigation_date) = ?', yesterdayStr),
    db.get<{ count: number }>('SELECT COUNT(*) as count FROM isolation_records WHERE DATE(start_date) = ?', yesterdayStr),
    db.get<{ count: number }>('SELECT COUNT(*) as count FROM test_records WHERE DATE(test_date) = ?', yesterdayStr),
    db.get<{ count: number }>('SELECT COUNT(*) as count FROM health_records WHERE DATE(record_date) = ?', yesterdayStr),
    db.get<{ count: number }>('SELECT COUNT(*) as count FROM persons WHERE is_key_person = 1'),
    db.get<{ count: number }>('SELECT COUNT(*) as count FROM isolation_records WHERE status = ?', 'active'),
    db.get<{ count: number }>('SELECT COUNT(*) as count FROM isolation_records WHERE status = ?', 'pending_discharge'),
  ]);

  return {
    date: todayStr,
    new_investigations: newInvestigations?.count || 0,
    new_isolations: newIsolations?.count || 0,
    new_tests: newTests?.count || 0,
    new_health_records: newHealthRecords?.count || 0,
    key_persons: keyPersons?.count || 0,
    active_isolations: activeIsolations?.count || 0,
    pending_discharge: pendingDischarge?.count || 0,
  };
}

export async function processExpiringIsolations(db: Database): Promise<number> {
  const expiring = await getExpiringIsolations(db);
  let processed = 0;

  for (const iso of expiring) {
    try {
      await generateDischargeNotification(db, iso.id);
      processed++;
    } catch (err) {
      console.error(`Failed to generate notification for isolation ${iso.id}:`, err);
    }
  }

  return processed;
}

export async function generateTestSchedule(db: Database): Promise<TestSchedule[]> {
  const schedules: TestSchedule[] = [];
  const todayStr = today();

  const activeIsolations = await db.all<any[]>(
    `SELECT ir.*, p.name as person_name 
     FROM isolation_records ir
     JOIN persons p ON ir.person_id = p.id
     WHERE ir.status = ?`,
    'active'
  );

  for (const iso of activeIsolations) {
    const day = calculateIsolationDay(iso.start_date, todayStr);
    if ([1, 7, 14].includes(day)) {
      schedules.push({
        person_id: iso.person_id,
        person_name: iso.person_name,
        isolation_id: iso.id,
        test_day: day,
        reason: '隔离人员定期检测',
      });
    }
  }

  const keyPersons = await db.all<any[]>(
    'SELECT id, name FROM persons WHERE is_key_person = 1'
  );

  for (const person of keyPersons) {
    const lastTest = await db.get<any>(
      'SELECT * FROM test_records WHERE person_id = ? ORDER BY test_date DESC LIMIT 1',
      person.id
    );

    let needsTest = false;
    let testDay = 1;

    if (lastTest) {
      const daysDiff = Math.floor(
        (new Date(todayStr).getTime() - new Date(lastTest.test_date).getTime()) / (1000 * 60 * 60 * 24)
      );
      if (daysDiff >= 3) {
        needsTest = true;
        testDay = lastTest.test_day + daysDiff;
      }
    } else {
      needsTest = true;
    }

    if (needsTest) {
      schedules.push({
        person_id: person.id,
        person_name: person.name,
        test_day: testDay,
        reason: '重点人员每3天检测',
      });
    }
  }

  return schedules;
}

export async function generateHealthCheckReminders(db: Database): Promise<string[]> {
  const todayStr = today();
  const keyPersons = await db.all<any[]>('SELECT id, name FROM persons WHERE is_key_person = 1');

  const reminders: string[] = [];

  for (const person of keyPersons) {
    const todayRecord = await db.get<any>(
      'SELECT * FROM health_records WHERE person_id = ? AND DATE(record_date) = ?',
      person.id,
      todayStr
    );

    if (!todayRecord) {
      reminders.push(`请 ${person.name} (ID: ${person.id}) 完成今日健康打卡`);
    }
  }

  return reminders;
}

export async function runDailyTasks(db: Database): Promise<{
  summary: DailySummary;
  notificationsGenerated: number;
  testSchedule: TestSchedule[];
  healthReminders: string[];
}> {
  console.log('Running daily epidemic management tasks...');

  const summary = await generateDailySummary(db);
  const notificationsGenerated = await processExpiringIsolations(db);
  const testSchedule = await generateTestSchedule(db);
  const healthReminders = await generateHealthCheckReminders(db);

  console.log('Daily summary:', JSON.stringify(summary, null, 2));
  console.log('Notifications generated:', notificationsGenerated);
  console.log('Test schedule entries:', testSchedule.length);
  console.log('Health reminders:', healthReminders.length);

  return { summary, notificationsGenerated, testSchedule, healthReminders };
}

export function setupDailyScheduler(db: Database): cron.ScheduledTask {
  return cron.schedule('0 8 * * *', async () => {
    try {
      await runDailyTasks(db);
    } catch (err) {
      console.error('Daily scheduler error:', err);
    }
  }, {
    scheduled: true,
    timezone: 'Asia/Shanghai'
  });
}
