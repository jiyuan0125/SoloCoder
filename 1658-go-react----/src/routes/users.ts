import { Router, Request, Response } from 'express';
import {
  getUserById,
  createUser,
  updateUser,
  listUsers
} from '../services/userService';

const router = Router();

router.post('/', (req: Request, res: Response) => {
  try {
    const { name, email, phone } = req.body;
    if (!name || !email || !phone) {
      return res.status(400).json({ error: 'Name, email, and phone are required' });
    }
    const user = createUser({ name, email, phone });
    res.status(201).json(user);
  } catch (error: any) {
    res.status(400).json({ error: error.message });
  }
});

router.get('/', (_req: Request, res: Response) => {
  const users = listUsers();
  res.json(users);
});

router.get('/:id', (req: Request, res: Response) => {
  const user = getUserById(req.params.id);
  if (!user) {
    return res.status(404).json({ error: 'User not found' });
  }
  res.json(user);
});

router.put('/:id', (req: Request, res: Response) => {
  const { name, phone } = req.body;
  const user = updateUser(req.params.id, { name, phone });
  if (!user) {
    return res.status(404).json({ error: 'User not found' });
  }
  res.json(user);
});

export default router;
