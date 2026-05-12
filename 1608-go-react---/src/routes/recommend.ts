import { Router, Request, Response } from 'express';
import * as recommendService from '../services/recommendService';
import * as ratingService from '../services/ratingService';
import * as statsService from '../services/statsService';

const router = Router();

router.get('/user/:userId', async (req: Request, res: Response) => {
  try {
    const userId = parseInt(req.params.userId, 10);
    if (isNaN(userId)) {
      return res.status(400).json({ error: 'Invalid user ID' });
    }

    const hasRatings = await ratingService.userExists(userId);
    if (!hasRatings) {
      return res.status(404).json({ error: 'User not found' });
    }

    const limit = typeof req.query.limit === 'string' ? parseInt(req.query.limit, 10) : 20;
    const result = await recommendService.getRecommendations(userId, limit);
    res.json(result);
  } catch (err) {
    res.status(500).json({ error: 'Internal server error' });
  }
});

router.post('/interact', async (req: Request, res: Response) => {
  try {
    const { userId, productId } = req.body;
    if (typeof userId !== 'number' || typeof productId !== 'number') {
      return res.status(400).json({ error: 'userId and productId must be numbers' });
    }
    const success = await recommendService.recordInteraction(userId, productId);
    if (!success) {
      return res.status(404).json({ error: 'Exposure record not found' });
    }
    res.json({ success: true });
  } catch (err) {
    res.status(500).json({ error: 'Internal server error' });
  }
});

router.get('/stats/today', async (_req: Request, res: Response) => {
  try {
    const stats = await statsService.updateTodayStats();
    res.json(stats);
  } catch (err) {
    res.status(500).json({ error: 'Internal server error' });
  }
});

router.get('/stats/:date', async (req: Request, res: Response) => {
  try {
    const { date } = req.params;
    const stats = await statsService.calculateDayStats(date);
    res.json(stats);
  } catch (err) {
    res.status(500).json({ error: 'Internal server error' });
  }
});

router.get('/stats', async (req: Request, res: Response) => {
  try {
    const startDate = req.query.startDate as string;
    const endDate = req.query.endDate as string;
    if (!startDate || !endDate) {
      return res.status(400).json({ error: 'startDate and endDate are required' });
    }
    const stats = await statsService.getStatsRange(startDate, endDate);
    res.json(stats);
  } catch (err) {
    res.status(500).json({ error: 'Internal server error' });
  }
});

export default router;
