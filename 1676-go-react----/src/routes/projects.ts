import express, { Request, Response } from 'express';
import {
  createProject,
  getProjectById,
  listProjects,
  getRewardTiersByProjectId,
  updateProjectStatus,
  supportProject,
  refundSupport,
  getSupportsByProjectId,
  getSupportsBySupporterAndProject,
  ProjectError,
} from '../services/projectService';
import {
  CreateProjectRequest,
  UpdateStatusRequest,
  SupportRequest,
  RefundRequest,
} from '../types';

const router = express.Router();

router.post('/', (req: Request, res: Response) => {
  try {
    const body = req.body as CreateProjectRequest;
    const result = createProject(body);
    res.status(201).json(result);
  } catch (err) {
    if (err instanceof ProjectError) {
      res.status(err.statusCode).json({ error: err.message });
    } else {
      res.status(500).json({ error: '服务器内部错误' });
    }
  }
});

router.get('/', (req: Request, res: Response) => {
  try {
    const projects = listProjects();
    res.json(projects);
  } catch (err) {
    res.status(500).json({ error: '服务器内部错误' });
  }
});

router.get('/:id', (req: Request, res: Response) => {
  try {
    const project = getProjectById(req.params.id);
    if (!project) {
      return res.status(404).json({ error: '项目不存在' });
    }
    res.json(project);
  } catch (err) {
    res.status(500).json({ error: '服务器内部错误' });
  }
});

router.get('/:id/rewards', (req: Request, res: Response) => {
  try {
    const project = getProjectById(req.params.id);
    if (!project) {
      return res.status(404).json({ error: '项目不存在' });
    }
    const rewards = getRewardTiersByProjectId(req.params.id);
    res.json(rewards);
  } catch (err) {
    res.status(500).json({ error: '服务器内部错误' });
  }
});

router.patch('/:id/status', (req: Request, res: Response) => {
  try {
    const body = req.body as UpdateStatusRequest;
    const updated = updateProjectStatus(req.params.id, body.status);
    res.json(updated);
  } catch (err) {
    if (err instanceof ProjectError) {
      res.status(err.statusCode).json({ error: err.message });
    } else {
      res.status(500).json({ error: '服务器内部错误' });
    }
  }
});

router.post('/:id/support', (req: Request, res: Response) => {
  try {
    const body = req.body as SupportRequest;
    const result = supportProject(body.rewardTierId, body.supporterId);
    res.status(201).json(result);
  } catch (err) {
    if (err instanceof ProjectError) {
      res.status(err.statusCode).json({ error: err.message });
    } else {
      res.status(500).json({ error: '服务器内部错误' });
    }
  }
});

router.post('/:id/support/:supportId/refund', (req: Request, res: Response) => {
  try {
    const result = refundSupport(req.params.supportId);
    res.status(200).json(result);
  } catch (err) {
    if (err instanceof ProjectError) {
      res.status(err.statusCode).json({ error: err.message });
    } else {
      res.status(500).json({ error: '服务器内部错误' });
    }
  }
});

router.get('/:id/supports', (req: Request, res: Response) => {
  try {
    const project = getProjectById(req.params.id);
    if (!project) {
      return res.status(404).json({ error: '项目不存在' });
    }
    const supports = getSupportsByProjectId(req.params.id);
    res.json(supports);
  } catch (err) {
    res.status(500).json({ error: '服务器内部错误' });
  }
});

router.get('/:id/supports/:supporterId', (req: Request, res: Response) => {
  try {
    const project = getProjectById(req.params.id);
    if (!project) {
      return res.status(404).json({ error: '项目不存在' });
    }
    const supports = getSupportsBySupporterAndProject(req.params.supporterId, req.params.id);
    res.json(supports);
  } catch (err) {
    res.status(500).json({ error: '服务器内部错误' });
  }
});

export default router;
