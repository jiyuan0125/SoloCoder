import { Router, Request, Response, NextFunction } from 'express';
import {
  createProject,
  updateProject,
  processTownshipReview,
  processCountyReview,
  getAllProjects,
  getProjectById
} from '../services/projectService';

const router = Router();

router.post('/', (req: Request, res: Response, next: NextFunction) => {
  try {
    const { name, category, location, totalBudget, implementationPeriod, budgetBreakdown } = req.body;
    const projectId = createProject({
      name,
      category,
      location,
      totalBudget,
      implementationPeriod,
      budgetBreakdown
    });
    res.status(201).json({
      success: true,
      data: { id: projectId },
      message: '项目申报成功，进入乡镇初审'
    });
  } catch (err) {
    next(err);
  }
});

router.get('/', (req: Request, res: Response, next: NextFunction) => {
  try {
    const projects = getAllProjects();
    res.json({
      success: true,
      data: projects
    });
  } catch (err) {
    next(err);
  }
});

router.get('/:id', (req: Request, res: Response, next: NextFunction) => {
  try {
    const projectId = parseInt(req.params.id, 10);
    const project = getProjectById(projectId);
    res.json({
      success: true,
      data: project
    });
  } catch (err) {
    next(err);
  }
});

router.put('/:id', (req: Request, res: Response, next: NextFunction) => {
  try {
    const projectId = parseInt(req.params.id, 10);
    const { name, category, location, totalBudget, implementationPeriod, budgetBreakdown } = req.body;
    updateProject(projectId, {
      name,
      category,
      location,
      totalBudget,
      implementationPeriod,
      budgetBreakdown
    });
    res.json({
      success: true,
      message: '项目修改成功，已重新提交审核'
    });
  } catch (err) {
    next(err);
  }
});

router.post('/:id/township-review', (req: Request, res: Response, next: NextFunction) => {
  try {
    const projectId = parseInt(req.params.id, 10);
    const { result, comments, reviewer } = req.body;
    processTownshipReview(projectId, result, comments || null, reviewer);
    res.json({
      success: true,
      message: result === 'PASSED' ? '乡镇初审通过，进入县级复审' : '乡镇初审不通过'
    });
  } catch (err) {
    next(err);
  }
});

router.post('/:id/county-review', (req: Request, res: Response, next: NextFunction) => {
  try {
    const projectId = parseInt(req.params.id, 10);
    const { result, comments, reviewer } = req.body;
    processCountyReview(projectId, result, comments || null, reviewer);
    res.json({
      success: true,
      message: result === 'PASSED' ? '县级复审通过，项目已立项' : '县级复审不通过'
    });
  } catch (err) {
    next(err);
  }
});

export default router;
