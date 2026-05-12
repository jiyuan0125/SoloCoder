import { Router, Request, Response } from 'express';
import { getDb } from '../db';

const router = Router();

router.post('/', async (req: Request, res: Response) => {
  const { name } = req.body;

  if (!name || typeof name !== 'string') {
    return res.status(400).json({ error: 'name is required' });
  }

  const db = getDb();
  const now = Date.now();

  try {
    await db.run(
      'INSERT INTO services (name, created_at) VALUES (?, ?)',
      [name, now]
    );
    res.status(201).json({ name, created_at: now });
  } catch (e: any) {
    if (e.code === 'SQLITE_CONSTRAINT') {
      res.status(409).json({ error: 'service already exists' });
    } else {
      res.status(500).json({ error: 'internal server error' });
    }
  }
});

export default router;
