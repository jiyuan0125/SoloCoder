import { Router, Request, Response } from 'express';
import {
  listTodosByUser,
  listOverdueTodos,
  markTodoCompleted,
  markTodoReminded
} from '../services/todoService';
import { authenticate, requireCommittee, requireExecutor } from '../middleware/auth';
import { sendError } from '../utils';

const router = Router();

router.get('/mine', authenticate, (req: Request, res: Response) => {
  try {
    const todos = listTodosByUser(req.user.id);
    res.json(todos);
  } catch (error) {
    sendError(res, { code: 500, message: '获取待办列表失败' });
  }
});

router.get('/overdue', authenticate, requireCommittee, (req: Request, res: Response) => {
  try {
    const todos = listOverdueTodos();
    for (const todo of todos) {
      if (!todo.isReminded) {
        markTodoReminded(todo.id);
      }
    }
    res.json(todos);
  } catch (error) {
    sendError(res, { code: 500, message: '获取超期待办失败' });
  }
});

router.post('/:id/complete', authenticate, (req: Request, res: Response) => {
  try {
    const id = parseInt(req.params.id, 10);
    const todo = markTodoCompleted(id);
    res.json(todo);
  } catch (error) {
    sendError(res, { code: 500, message: '完成待办失败' });
  }
});

export { router as todosRouter };
