import { Router, Request, Response } from 'express';
import { getDatabase, DatabaseWrapper } from '../database';
import { generateId, getHourStart, addHours } from '../utils';
import { logAudit } from '../audit';

const router = Router({ mergeParams: true });

const FAILURE_RATE_THRESHOLD = 0.20;
const FAILURE_CHECK_HOURS = 1;

router.get('/', async (req: Request, res: Response) => {
  try {
    const { devId, appId } = req.params;
    const db = getDatabase();
    
    const app = await db.get<any>(
      `SELECT id FROM apps WHERE id = ? AND developer_id = ?`,
      [appId, devId]
    );
    
    if (!app) {
      return res.status(404).json({ error: 'App not found' });
    }
    
    const limit = parseInt(req.query.limit as string) || 24;
    const stats = await db.all<any>(
      `SELECT 
         hour_start,
         total_calls,
         success_calls,
         (total_calls - success_calls) as failure_calls,
         CASE WHEN total_calls > 0 THEN (success_calls * 100.0 / total_calls) ELSE 0 END as success_rate,
         CASE WHEN total_calls > 0 THEN (total_response_time / total_calls) ELSE 0 END as avg_response_time_ms
       FROM api_stats 
       WHERE app_id = ? 
       ORDER BY hour_start DESC 
       LIMIT ?`,
      [appId, limit]
    );
    
    const summary = await db.get<any>(
      `SELECT 
         COALESCE(SUM(total_calls), 0) as total_calls,
         COALESCE(SUM(success_calls), 0) as success_calls,
         COALESCE(SUM(total_calls) - SUM(success_calls), 0) as failure_calls,
         CASE WHEN COALESCE(SUM(total_calls), 0) > 0 
              THEN (SUM(success_calls) * 100.0 / SUM(total_calls)) 
              ELSE 0 END as overall_success_rate,
         CASE WHEN COALESCE(SUM(total_calls), 0) > 0 
              THEN (SUM(total_response_time) / SUM(total_calls)) 
              ELSE 0 END as overall_avg_response_time_ms
       FROM api_stats 
       WHERE app_id = ?`,
      [appId]
    );
    
    res.json({
      app_id: appId,
      summary: summary,
      hourly_stats: stats
    });
  } catch (err) {
    console.error(err);
    res.status(500).json({ error: 'Internal server error' });
  }
});

router.post('/record', async (req: Request, res: Response) => {
  try {
    const { devId, appId } = req.params;
    const { success, response_time_ms } = req.body;
    
    if (typeof success !== 'boolean') {
      return res.status(400).json({ error: 'success (boolean) is required' });
    }
    
    const responseTime = typeof response_time_ms === 'number' ? response_time_ms : 0;
    
    const db = getDatabase();
    
    const app = await db.get<any>(
      `SELECT * FROM apps WHERE id = ? AND developer_id = ?`,
      [appId, devId]
    );
    
    if (!app) {
      return res.status(404).json({ error: 'App not found' });
    }
    
    if (app.status === 'disabled') {
      return res.status(403).json({ error: 'App is disabled' });
    }
    
    const hourStart = getHourStart();
    const hourStartStr = hourStart.toISOString();
    
    const existing = await db.get<any>(
      `SELECT id FROM api_stats WHERE app_id = ? AND hour_start = ?`,
      [appId, hourStartStr]
    );
    
    if (existing) {
      await db.run(
        `UPDATE api_stats 
         SET total_calls = total_calls + 1,
             success_calls = success_calls + ?,
             total_response_time = total_response_time + ?
         WHERE app_id = ? AND hour_start = ?`,
        [success ? 1 : 0, responseTime, appId, hourStartStr]
      );
    } else {
      await db.run(
        `INSERT INTO api_stats (id, app_id, hour_start, total_calls, success_calls, total_response_time)
         VALUES (?, ?, ?, ?, ?, ?)`,
        [generateId(), appId, hourStartStr, 1, success ? 1 : 0, responseTime]
      );
    }
    
    await checkAndDisableAppIfNeeded(devId, appId, db);
    
    res.json({ recorded: true });
  } catch (err) {
    console.error(err);
    res.status(500).json({ error: 'Internal server error' });
  }
});

async function checkAndDisableAppIfNeeded(devId: string, appId: string, db: DatabaseWrapper): Promise<void> {
  const checkStart = addHours(new Date(), -FAILURE_CHECK_HOURS);
  
  const stats = await db.get<any>(
    `SELECT 
       COALESCE(SUM(total_calls), 0) as total_calls,
       COALESCE(SUM(success_calls), 0) as success_calls
     FROM api_stats 
     WHERE app_id = ? AND hour_start >= ?`,
    [appId, checkStart.toISOString()]
  );
  
  if (!stats || stats.total_calls === 0) {
    return;
  }
  
  const failureRate = (stats.total_calls - stats.success_calls) / stats.total_calls;
  
  if (failureRate > FAILURE_RATE_THRESHOLD) {
    await db.beginTransaction();
    
    try {
      const app = await db.get<any>(
        `SELECT * FROM apps WHERE id = ?`,
        [appId]
      );
      
      if (app.status === 'disabled') {
        await db.rollback();
        return;
      }
      
      await db.run(
        `UPDATE apps SET status = 'disabled', updated_at = ? WHERE id = ?`,
        [new Date().toISOString(), appId]
      );
      
      await db.run(
        `UPDATE api_keys SET status = 'disabled' WHERE app_id = ?`,
        [appId]
      );
      
      const failureRatePercent = (failureRate * 100).toFixed(2);
      const message = `App disabled due to high failure rate (${failureRatePercent}%) in the last ${FAILURE_CHECK_HOURS} hour(s). Total calls: ${stats.total_calls}, Failures: ${stats.total_calls - stats.success_calls}`;
      
      await db.run(
        `INSERT INTO notifications (id, app_id, type, message) VALUES (?, ?, ?, ?)`,
        [generateId(), appId, 'app_disabled', message]
      );
      
      await logAudit(devId, appId, 'app_auto_disabled', {
        failure_rate: failureRate,
        total_calls: stats.total_calls,
        success_calls: stats.success_calls,
        failure_calls: stats.total_calls - stats.success_calls,
        window_hours: FAILURE_CHECK_HOURS
      });
      
      await db.commit();
    } catch (err) {
      await db.rollback();
      throw err;
    }
  }
}

export default router;
