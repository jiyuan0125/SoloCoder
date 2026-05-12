import { Router, Request, Response } from 'express';
import { getDatabase } from '../database';
import { generateId, hashSecret, generateAppSecret, isValidUrl, addMinutes, getMinutesUntil } from '../utils';
import { logAudit } from '../audit';

const router = Router({ mergeParams: true });

const MAX_APPS_PER_DEVELOPER = 10;
const KEY_RESET_COOLDOWN_HOURS = 24;
const TRANSITION_PERIOD_MINUTES = 5;

interface App {
  id: string;
  developer_id: string;
  name: string;
  description: string | null;
  callback_url: string;
  status: string;
  created_at: string;
  updated_at: string;
}

router.get('/', async (req: Request, res: Response) => {
  try {
    const { devId } = req.params;
    const db = getDatabase();
    
    const developer = await db.get<any>(`SELECT id FROM developers WHERE id = ?`, [devId]);
    if (!developer) {
      return res.status(404).json({ error: 'Developer not found' });
    }
    
    const apps = await db.all<App>(
      `SELECT id, developer_id, name, description, callback_url, status, created_at, updated_at 
       FROM apps WHERE developer_id = ? ORDER BY created_at DESC`,
      [devId]
    );
    
    res.json(apps);
  } catch (err) {
    console.error(err);
    res.status(500).json({ error: 'Internal server error' });
  }
});

router.get('/:appId', async (req: Request, res: Response) => {
  try {
    const { devId, appId } = req.params;
    const db = getDatabase();
    
    const app = await db.get<App>(
      `SELECT id, developer_id, name, description, callback_url, status, created_at, updated_at 
       FROM apps WHERE id = ? AND developer_id = ?`,
      [appId, devId]
    );
    
    if (!app) {
      return res.status(404).json({ error: 'App not found' });
    }
    
    res.json(app);
  } catch (err) {
    console.error(err);
    res.status(500).json({ error: 'Internal server error' });
  }
});

router.post('/', async (req: Request, res: Response) => {
  try {
    const { devId } = req.params;
    const { name, description, callback_url } = req.body;
    
    if (!name || !callback_url) {
      return res.status(400).json({ error: 'Name and callback_url are required' });
    }
    
    if (!isValidUrl(callback_url)) {
      return res.status(400).json({ error: 'Invalid callback URL format' });
    }
    
    const db = getDatabase();
    
    const developer = await db.get<any>(`SELECT id FROM developers WHERE id = ?`, [devId]);
    if (!developer) {
      return res.status(404).json({ error: 'Developer not found' });
    }
    
    const appCount = await db.get<any>(
      `SELECT COUNT(*) as count FROM apps WHERE developer_id = ?`,
      [devId]
    );
    
    if (appCount && appCount.count >= MAX_APPS_PER_DEVELOPER) {
      return res.status(403).json({ 
        error: 'Maximum number of apps reached',
        max: MAX_APPS_PER_DEVELOPER 
      });
    }
    
    await db.beginTransaction();
    
    try {
      const appId = generateId('app');
      const now = new Date().toISOString();
      
      await db.run(
        `INSERT INTO apps (id, developer_id, name, description, callback_url, created_at, updated_at) 
         VALUES (?, ?, ?, ?, ?, ?, ?)`,
        [appId, devId, name, description || null, callback_url, now, now]
      );
      
      const appSecret = generateAppSecret();
      const keyHash = hashSecret(appSecret);
      const keyId = generateId('key');
      
      await db.run(
        `INSERT INTO api_keys (id, app_id, key_hash, version, status) VALUES (?, ?, ?, ?, ?)`,
        [keyId, appId, keyHash, 1, 'active']
      );
      
      await logAudit(devId, appId, 'app_created', { name, callback_url });
      await logAudit(devId, appId, 'key_created', { version: 1 });
      
      await db.commit();
      
      const app = await db.get<App>(
        `SELECT id, developer_id, name, description, callback_url, status, created_at, updated_at 
         FROM apps WHERE id = ?`,
        [appId]
      );
      
      res.status(201).json({
        ...app,
        app_secret: appSecret
      });
    } catch (err) {
      await db.rollback();
      throw err;
    }
  } catch (err) {
    console.error(err);
    res.status(500).json({ error: 'Internal server error' });
  }
});

router.put('/:appId', async (req: Request, res: Response) => {
  try {
    const { devId, appId } = req.params;
    const { name, description, callback_url } = req.body;
    
    const db = getDatabase();
    
    const app = await db.get<App>(
      `SELECT * FROM apps WHERE id = ? AND developer_id = ?`,
      [appId, devId]
    );
    
    if (!app) {
      return res.status(404).json({ error: 'App not found' });
    }
    
    if (callback_url !== undefined && !isValidUrl(callback_url)) {
      return res.status(400).json({ error: 'Invalid callback URL format' });
    }
    
    const now = new Date().toISOString();
    const updates: any[] = [];
    const fields: string[] = [];
    
    if (name !== undefined) {
      fields.push('name = ?');
      updates.push(name);
    }
    if (description !== undefined) {
      fields.push('description = ?');
      updates.push(description);
    }
    if (callback_url !== undefined) {
      fields.push('callback_url = ?');
      updates.push(callback_url);
    }
    
    if (fields.length === 0) {
      return res.status(400).json({ error: 'No fields to update' });
    }
    
    fields.push('updated_at = ?');
    updates.push(now, appId, devId);
    
    await db.run(
      `UPDATE apps SET ${fields.join(', ')} WHERE id = ? AND developer_id = ?`,
      updates
    );
    
    await logAudit(devId, appId, 'app_updated', {
      old: app,
      new: { name, description, callback_url }
    });
    
    const updatedApp = await db.get<App>(
      `SELECT id, developer_id, name, description, callback_url, status, created_at, updated_at 
       FROM apps WHERE id = ?`,
      [appId]
    );
    
    res.json(updatedApp);
  } catch (err) {
    console.error(err);
    res.status(500).json({ error: 'Internal server error' });
  }
});

router.delete('/:appId', async (req: Request, res: Response) => {
  try {
    const { devId, appId } = req.params;
    const db = getDatabase();
    
    const app = await db.get<App>(
      `SELECT * FROM apps WHERE id = ? AND developer_id = ?`,
      [appId, devId]
    );
    
    if (!app) {
      return res.status(404).json({ error: 'App not found' });
    }
    
    await db.run(`DELETE FROM api_keys WHERE app_id = ?`, [appId]);
    await db.run(`DELETE FROM key_resets WHERE app_id = ?`, [appId]);
    await db.run(`DELETE FROM api_stats WHERE app_id = ?`, [appId]);
    await db.run(`DELETE FROM notifications WHERE app_id = ?`, [appId]);
    await db.run(`DELETE FROM apps WHERE id = ? AND developer_id = ?`, [appId, devId]);
    
    await logAudit(devId, appId, 'app_deleted', {});
    
    res.status(204).send();
  } catch (err) {
    console.error(err);
    res.status(500).json({ error: 'Internal server error' });
  }
});

router.post('/:appId/reset-secret', async (req: Request, res: Response) => {
  try {
    const { devId, appId } = req.params;
    const db = getDatabase();
    
    const app = await db.get<App>(
      `SELECT * FROM apps WHERE id = ? AND developer_id = ?`,
      [appId, devId]
    );
    
    if (!app) {
      return res.status(404).json({ error: 'App not found' });
    }
    
    if (app.status === 'disabled') {
      return res.status(403).json({ error: 'App is disabled' });
    }
    
    const lastReset = await db.get<any>(
      `SELECT reset_at FROM key_resets WHERE app_id = ? ORDER BY reset_at DESC LIMIT 1`,
      [appId]
    );
    
    if (lastReset) {
      const resetTime = new Date(lastReset.reset_at);
      const cooldownEnd = new Date(resetTime.getTime() + KEY_RESET_COOLDOWN_HOURS * 60 * 60 * 1000);
      
      if (new Date() < cooldownEnd) {
        const remainingMinutes = getMinutesUntil(cooldownEnd);
        return res.status(429).json({
          error: 'Key reset cooldown active',
          remaining_minutes: remainingMinutes,
          cooldown_hours: KEY_RESET_COOLDOWN_HOURS
        });
      }
    }
    
    await db.beginTransaction();
    
    try {
      const currentKey = await db.get<any>(
        `SELECT * FROM api_keys WHERE app_id = ? AND status = 'active' ORDER BY version DESC LIMIT 1`,
        [appId]
      );
      
      if (!currentKey) {
        await db.rollback();
        return res.status(500).json({ error: 'No active key found' });
      }
      
      const transitionUntil = addMinutes(new Date(), TRANSITION_PERIOD_MINUTES);
      
      await db.run(
        `UPDATE api_keys SET status = 'transitioning', transition_until = ? WHERE id = ?`,
        [transitionUntil.toISOString(), currentKey.id]
      );
      
      const newSecret = generateAppSecret();
      const newHash = hashSecret(newSecret);
      const newKeyId = generateId('key');
      const newVersion = currentKey.version + 1;
      
      await db.run(
        `INSERT INTO api_keys (id, app_id, key_hash, version, status) VALUES (?, ?, ?, ?, ?)`,
        [newKeyId, appId, newHash, newVersion, 'active']
      );
      
      await db.run(
        `INSERT INTO key_resets (id, app_id) VALUES (?, ?)`,
        [generateId(), appId]
      );
      
      await logAudit(devId, appId, 'key_reset', {
        old_version: currentKey.version,
        new_version: newVersion,
        transition_until: transitionUntil.toISOString()
      });
      
      await db.commit();
      
      res.json({
        app_id: appId,
        app_secret: newSecret,
        version: newVersion,
        transition_until: transitionUntil.toISOString(),
        transition_minutes: TRANSITION_PERIOD_MINUTES
      });
    } catch (err) {
      await db.rollback();
      throw err;
    }
  } catch (err) {
    console.error(err);
    res.status(500).json({ error: 'Internal server error' });
  }
});

export default router;
