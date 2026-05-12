import { Router, Request, Response } from 'express';
import { getDatabase } from '../database';
import { generateId } from '../utils';
import { logAudit } from '../audit';

const router = Router();

interface Developer {
  id: string;
  name: string;
  email: string;
  company: string | null;
  created_at: string;
  updated_at: string;
}

router.get('/', async (req: Request, res: Response) => {
  try {
    const db = getDatabase();
    const developers = await db.all<Developer>(
      `SELECT id, name, email, company, created_at, updated_at FROM developers ORDER BY created_at DESC`
    );
    res.json(developers);
  } catch (err) {
    console.error(err);
    res.status(500).json({ error: 'Internal server error' });
  }
});

router.get('/:id', async (req: Request, res: Response) => {
  try {
    const { id } = req.params;
    const db = getDatabase();
    const developer = await db.get<Developer>(
      `SELECT id, name, email, company, created_at, updated_at FROM developers WHERE id = ?`,
      [id]
    );
    
    if (!developer) {
      return res.status(404).json({ error: 'Developer not found' });
    }
    
    res.json(developer);
  } catch (err) {
    console.error(err);
    res.status(500).json({ error: 'Internal server error' });
  }
});

router.post('/', async (req: Request, res: Response) => {
  try {
    const { name, email, company } = req.body;
    
    if (!name || !email) {
      return res.status(400).json({ error: 'Name and email are required' });
    }
    
    const db = getDatabase();
    const id = generateId('dev');
    const now = new Date().toISOString();
    
    await db.run(
      `INSERT INTO developers (id, name, email, company, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
      [id, name, email, company || null, now, now]
    );
    
    await logAudit(id, null, 'developer_created', { name, email });
    
    const developer = await db.get<Developer>(
      `SELECT id, name, email, company, created_at, updated_at FROM developers WHERE id = ?`,
      [id]
    );
    
    res.status(201).json(developer);
  } catch (err: any) {
    if (err.message && err.message.includes('UNIQUE constraint failed')) {
      return res.status(409).json({ error: 'Email already exists' });
    }
    console.error(err);
    res.status(500).json({ error: 'Internal server error' });
  }
});

router.put('/:id', async (req: Request, res: Response) => {
  try {
    const { id } = req.params;
    const { name, email, company } = req.body;
    
    const db = getDatabase();
    const existing = await db.get<Developer>(
      `SELECT * FROM developers WHERE id = ?`,
      [id]
    );
    
    if (!existing) {
      return res.status(404).json({ error: 'Developer not found' });
    }
    
    const now = new Date().toISOString();
    const updates: any[] = [];
    const fields: string[] = [];
    
    if (name !== undefined) {
      fields.push('name = ?');
      updates.push(name);
    }
    if (email !== undefined) {
      fields.push('email = ?');
      updates.push(email);
    }
    if (company !== undefined) {
      fields.push('company = ?');
      updates.push(company);
    }
    
    if (fields.length === 0) {
      return res.status(400).json({ error: 'No fields to update' });
    }
    
    fields.push('updated_at = ?');
    updates.push(now, id);
    
    await db.run(
      `UPDATE developers SET ${fields.join(', ')} WHERE id = ?`,
      updates
    );
    
    await logAudit(id, null, 'developer_updated', { 
      old: existing, 
      new: { name, email, company } 
    });
    
    const developer = await db.get<Developer>(
      `SELECT id, name, email, company, created_at, updated_at FROM developers WHERE id = ?`,
      [id]
    );
    
    res.json(developer);
  } catch (err: any) {
    if (err.message && err.message.includes('UNIQUE constraint failed')) {
      return res.status(409).json({ error: 'Email already exists' });
    }
    console.error(err);
    res.status(500).json({ error: 'Internal server error' });
  }
});

router.delete('/:id', async (req: Request, res: Response) => {
  try {
    const { id } = req.params;
    const db = getDatabase();
    
    const developer = await db.get<Developer>(
      `SELECT * FROM developers WHERE id = ?`,
      [id]
    );
    
    if (!developer) {
      return res.status(404).json({ error: 'Developer not found' });
    }
    
    const activeApps = await db.get<any>(
      `SELECT COUNT(*) as count FROM apps WHERE developer_id = ? AND status != 'disabled'`,
      [id]
    );
    
    if (activeApps && activeApps.count > 0) {
      return res.status(409).json({ 
        error: 'Cannot delete developer with active apps',
        activeAppsCount: activeApps.count 
      });
    }
    
    await db.run(`DELETE FROM developers WHERE id = ?`, [id]);
    await logAudit(id, null, 'developer_deleted', {});
    
    res.status(204).send();
  } catch (err) {
    console.error(err);
    res.status(500).json({ error: 'Internal server error' });
  }
});

export default router;
