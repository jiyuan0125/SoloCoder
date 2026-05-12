import { Router, Request, Response } from 'express';
import { assignUser } from '../services/assignmentService';

const router = Router();

router.post('/', (req: Request, res: Response): void => {
  try {
    const { experimentId, userId, deviceId, region } = req.body;

    if (!experimentId) {
      res.status(400).json({ error: 'experimentId is required' });
      return;
    }

    const result = assignUser({
      experimentId,
      userId,
      deviceId,
      region
    });

    res.json(result);
  } catch (error: any) {
    res.status(400).json({ error: error.message });
  }
});

router.post('/batch', (req: Request, res: Response): void => {
  try {
    const { experimentIds, userId, deviceId, region } = req.body;

    if (!experimentIds || !Array.isArray(experimentIds)) {
      res.status(400).json({ error: 'experimentIds array is required' });
      return;
    }

    const results = experimentIds.map((experimentId: string) => {
      try {
        return assignUser({
          experimentId,
          userId,
          deviceId,
          region
        });
      } catch {
        return null;
      }
    });

    res.json(results);
  } catch (error: any) {
    res.status(400).json({ error: error.message });
  }
});

export default router;
