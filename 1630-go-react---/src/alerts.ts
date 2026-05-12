import { getDb } from './db';

export function evaluateThreshold(value: number, operator: string, threshold: number): boolean {
  switch (operator) {
    case 'gt':
      return value > threshold;
    case 'lt':
      return value < threshold;
    case 'eq':
      return value === threshold;
    default:
      return false;
  }
}

interface AlertRule {
  id: number;
  metric_name: string;
  operator: string;
  threshold: number;
  duration: number;
  level: string;
}

interface AlertRecord {
  id: number;
  rule_id: number;
  service_name: string;
  current_value: number;
  level: string;
  status: string;
  timestamp: number;
  last_value: number;
  last_timestamp: number;
}

export async function checkAlertRules(
  serviceName: string,
  metricName: string,
  value: number,
  timestamp: number
): Promise<void> {
  const db = getDb();

  const rules = await db.all<AlertRule[]>(
    'SELECT * FROM alert_rules WHERE metric_name = ?',
    [metricName]
  );

  if (rules.length === 0) return;

  for (const rule of rules) {
    const existingPending = await db.get<AlertRecord>(
      'SELECT * FROM alerts WHERE rule_id = ? AND service_name = ? AND status = ?',
      [rule.id, serviceName, 'pending']
    );

    const existingProcessing = await db.get<AlertRecord>(
      'SELECT * FROM alerts WHERE rule_id = ? AND service_name = ? AND status = ?',
      [rule.id, serviceName, 'processing']
    );

    const isOverThreshold = evaluateThreshold(value, rule.operator, rule.threshold);

    if (existingPending || existingProcessing) {
      const existing = existingPending || existingProcessing!;
      if (isOverThreshold) {
        await db.run(
          'UPDATE alerts SET last_value = ?, last_timestamp = ? WHERE id = ?',
          [value, timestamp, existing.id]
        );

        const durationOver = timestamp - existing.timestamp;
        if (durationOver >= rule.duration && existing.status === 'pending') {
          await db.run(
            'UPDATE alerts SET status = ? WHERE id = ?',
            ['processing', existing.id]
          );
        }
      } else {
        await db.run(
          'UPDATE alerts SET status = ? WHERE id = ?',
          ['recovered', existing.id]
        );
      }
    } else if (isOverThreshold) {
      try {
        await db.run(
          `INSERT INTO alerts (
            rule_id, service_name, current_value, level, status, timestamp,
            last_value, last_timestamp
          ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
          [
            rule.id,
            serviceName,
            value,
            rule.level,
            'pending',
            timestamp,
            value,
            timestamp
          ]
        );
      } catch (e: any) {
        if (!e.message?.includes('UNIQUE')) {
          throw e;
        }
      }
    }
  }
}

export async function cleanupOldMetrics(): Promise<void> {
  const db = getDb();
  const sevenDaysAgo = Date.now() - 7 * 24 * 60 * 60 * 1000;
  await db.run('DELETE FROM metrics WHERE timestamp < ?', [sevenDaysAgo]);
}
