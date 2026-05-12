import { getDatabase } from './database';
import { generateId } from './utils';

export async function logAudit(
  developerId: string | null,
  appId: string | null,
  action: string,
  details: Record<string, any> = {}
): Promise<void> {
  const db = getDatabase();
  await db.run(
    `INSERT INTO audit_logs (id, developer_id, app_id, action, details) VALUES (?, ?, ?, ?, ?)`,
    [generateId(), developerId, appId, action, JSON.stringify(details)]
  );
}
