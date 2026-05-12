import { Router, Request, Response } from 'express';
import { createDeployment, listDeployments } from '../services/deploymentService';

const router = Router();

router.post('/deployments', (req: Request, res: Response) => {
  try {
    const { version, deployedAt } = req.body as { version: string; deployedAt?: string };
    const result = createDeployment(version, deployedAt);
    res.status(201).json(result);
  } catch (error: any) {
    const statusCode = error.statusCode || 500;
    res.status(statusCode).json({ error: error.message });
  }
});

router.get('/deployments', (_req: Request, res: Response) => {
  try {
    const deployments = listDeployments();
    res.status(200).json(deployments);
  } catch (error: any) {
    const statusCode = error.statusCode || 500;
    res.status(statusCode).json({ error: error.message });
  }
});

export default router;
