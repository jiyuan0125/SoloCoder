import { Router, Request, Response } from 'express';
import { listUsers, getUserByUsername } from '../services/userService';
import { sendError } from '../utils';

const router = Router();

router.get('/', (req: Request, res: Response) => {
  try {
    const users = listUsers();
    res.json(users);
  } catch (error) {
    sendError(res, { code: 500, message: '获取用户列表失败' });
  }
});

router.get('/:username', (req: Request, res: Response) => {
  try {
    const user = getUserByUsername(req.params.username);
    if (!user) {
      return sendError(res, { code: 404, message: '用户不存在' });
    }
    res.json(user);
  } catch (error) {
    sendError(res, { code: 500, message: '获取用户失败' });
  }
});

export { router as usersRouter };
