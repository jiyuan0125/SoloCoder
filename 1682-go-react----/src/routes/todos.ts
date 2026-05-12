import { Router, Request, Response } from 'express';
import { todoService } from '../services/todoService';
import { TodoStatus } from '../types';

const router = Router();

router.get('/', (req: Request, res: Response) => {
  const { assignee, status } = req.query;
  
  let todos;
  
  if (assignee) {
    todos = todoService.getTodosByAssignee(
      assignee as string,
      status as TodoStatus
    );
  } else {
    todos = todoService.getTodos();
  }
  
  res.json(todos);
});

router.get('/pending', (req: Request, res: Response) => {
  const todos = todoService.getPendingTodos();
  res.json(todos);
});

router.get('/:id', (req: Request, res: Response) => {
  const todo = todoService.getTodo(req.params.id);
  
  if (!todo) {
    return res.status(404).json({ error: 'Todo not found' });
  }
  
  res.json(todo);
});

router.post('/:id/complete', (req: Request, res: Response) => {
  const todo = todoService.getTodo(req.params.id);
  
  if (!todo) {
    return res.status(404).json({ error: 'Todo not found' });
  }
  
  if (todo.status !== TodoStatus.PENDING) {
    return res.status(409).json({ error: 'Todo has already been processed' });
  }
  
  const updated = todoService.completeTodo(req.params.id);
  res.json(updated);
});

router.post('/:id/escalate', (req: Request, res: Response) => {
  const todo = todoService.getTodo(req.params.id);
  
  if (!todo) {
    return res.status(404).json({ error: 'Todo not found' });
  }
  
  if (todo.status !== TodoStatus.PENDING) {
    return res.status(409).json({ error: 'Todo has already been processed' });
  }
  
  const updated = todoService.escalateTodo(req.params.id);
  res.json(updated);
});

router.post('/check-overdue', (req: Request, res: Response) => {
  const escalated = todoService.checkAndEscalateOverdueTodos();
  res.json({ 
    message: 'Checked overdue todos',
    escalatedCount: escalated.length,
    escalatedTodos: escalated
  });
});

export default router;
