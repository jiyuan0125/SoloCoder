import { Router, Request, Response } from 'express';
import { generateId, getCurrentTime } from '../utils/id';
import { runSQL, getOne, getAll } from '../db';
import { User } from '../types';

const router = Router();

router.post('/', async (req: Request, res: Response): Promise<void> => {
  try {
    const { name, building } = req.body;
    
    if (!name || !building) {
      res.status(400).json({ error: '用户名和楼栋不能为空' });
      return;
    }

    const id = generateId();
    const now = getCurrentTime();
    
    await runSQL(
      'INSERT INTO users (id, name, building, created_at) VALUES (?, ?, ?, ?)',
      [id, name, building, now]
    );

    const user = await getOne<UserRow>(
      'SELECT * FROM users WHERE id = ?',
      [id]
    );

    res.status(201).json(mapToUser(user!));
  } catch (err) {
    console.error(err);
    res.status(500).json({ error: '内部服务器错误' });
  }
});

router.get('/:id', async (req: Request, res: Response): Promise<void> => {
  try {
    const user = await getOne<UserRow>(
      'SELECT * FROM users WHERE id = ?',
      [req.params.id]
    );

    if (!user) {
      res.status(404).json({ error: '用户不存在' });
      return;
    }

    res.json(mapToUser(user));
  } catch (err) {
    console.error(err);
    res.status(500).json({ error: '内部服务器错误' });
  }
});

router.get('/', async (_req: Request, res: Response): Promise<void> => {
  try {
    const users = await getAll<UserRow>('SELECT * FROM users');
    res.json(users.map(mapToUser));
  } catch (err) {
    console.error(err);
    res.status(500).json({ error: '内部服务器错误' });
  }
});

interface UserRow {
  id: string;
  name: string;
  building: string;
  avg_rating: number;
  rating_count: number;
  cool_down_until?: number;
  created_at: number;
}

function mapToUser(row: UserRow): User {
  return {
    id: row.id,
    name: row.name,
    building: row.building,
    avgRating: row.avg_rating || 0,
    ratingCount: row.rating_count || 0,
    coolDownUntil: row.cool_down_until,
    createdAt: row.created_at,
  };
}

export { router as usersRouter };
