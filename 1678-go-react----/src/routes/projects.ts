import { Router, Request, Response } from 'express';
import { ProjectApplication, ProjectStatus } from '../types';
import {
  createProject,
  getAllProjects,
  getProjectById,
  updateProjectStatus,
  allocateFunds,
  getFundAllocations,
  processOverdueTasks
} from '../services/projectService';
import { getProjectReviews, finalizeReview, submitReview, selectRandomExperts } from '../services/reviewService';
import { getTodosByProject, completeTodo, getAllTodos } from '../services/todoService';
import { getBudgetAlerts, getAllBudgetAlerts } from '../services/budgetService';

const router = Router();

function getOperator(req: Request): string {
  return (req.headers['x-operator'] as string) || 'anonymous';
}

router.post('/', (req: Request, res: Response) => {
  const app = req.body as ProjectApplication;
  const operator = getOperator(req);
  
  const result = createProject(app, operator);
  
  if (!result.success) {
    const message = result.message;
    if (message.includes('信用代码') || message.includes('预算') || message === '已达最大提交次数') {
      return res.status(400).json({ error: message });
    }
    return res.status(400).json({ error: message });
  }
  
  res.status(201).json(result.project);
});

router.get('/', (req: Request, res: Response) => {
  const projects = getAllProjects();
  res.json(projects);
});

router.get('/:id', (req: Request, res: Response) => {
  const project = getProjectById(req.params.id);
  if (!project) {
    return res.status(404).json({ error: '项目不存在' });
  }
  res.json(project);
});

router.patch('/:id/status', (req: Request, res: Response) => {
  const { status, reason } = req.body;
  const operator = getOperator(req);
  
  const validStatuses: ProjectStatus[] = ['申请中', '初审', '评审', '批准', '整改', '不通过'];
  if (!validStatuses.includes(status)) {
    return res.status(400).json({ error: '无效的状态值' });
  }
  
  const result = updateProjectStatus(req.params.id, status, operator, reason);
  
  if (!result.success) {
    return res.status(400).json({ error: result.message });
  }
  
  res.json(result.project);
});

router.post('/:id/funds', (req: Request, res: Response) => {
  const { amount } = req.body;
  const operator = getOperator(req);
  
  if (typeof amount !== 'number' || amount <= 0) {
    return res.status(400).json({ error: '拨付金额必须为正数' });
  }
  
  const result = allocateFunds(req.params.id, amount, operator);
  
  if (!result.success) {
    return res.status(400).json({ error: result.message });
  }
  
  res.status(201).json(result.allocation);
});

router.get('/:id/funds', (req: Request, res: Response) => {
  const allocations = getFundAllocations(req.params.id);
  res.json(allocations);
});

router.get('/:id/reviews', (req: Request, res: Response) => {
  const reviews = getProjectReviews(req.params.id);
  res.json(reviews);
});

router.post('/:id/reviews', (req: Request, res: Response) => {
  const { expertId, score } = req.body;
  const operator = getOperator(req);
  
  const result = submitReview(req.params.id, expertId, score, operator);
  
  if (!result.success) {
    if (result.message.includes('分数')) {
      return res.status(400).json({ error: result.message });
    }
    return res.status(400).json({ error: result.message });
  }
  
  res.status(201).json(result.review);
});

router.post('/:id/reviews/finalize', (req: Request, res: Response) => {
  const operator = getOperator(req);
  const result = finalizeReview(req.params.id, operator);
  
  if (!result.success) {
    return res.status(400).json({ error: result.message });
  }
  
  res.json(result);
});

router.get('/:id/todos', (req: Request, res: Response) => {
  const todos = getTodosByProject(req.params.id);
  res.json(todos);
});

router.get('/:id/alerts', (req: Request, res: Response) => {
  const alerts = getBudgetAlerts(req.params.id);
  res.json(alerts);
});

router.post('/todos/process-overdue', (req: Request, res: Response) => {
  const count = processOverdueTasks();
  res.json({ processed: count });
});

router.get('/todos/all', (req: Request, res: Response) => {
  const todos = getAllTodos();
  res.json(todos);
});

router.patch('/todos/:id/complete', (req: Request, res: Response) => {
  const operator = getOperator(req);
  const result = completeTodo(req.params.id, operator);
  
  if (!result.success) {
    return res.status(400).json({ error: result.message });
  }
  
  res.json({ message: result.message });
});

router.get('/alerts/all', (req: Request, res: Response) => {
  const alerts = getAllBudgetAlerts();
  res.json(alerts);
});

router.get('/experts/random', (req: Request, res: Response) => {
  const min = parseInt(req.query.min as string) || 3;
  const max = parseInt(req.query.max as string) || 5;
  const experts = selectRandomExperts(min, max);
  res.json(experts);
});

export default router;
