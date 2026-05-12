import { Router, Request, Response } from 'express';
import { 
  getLevelDistribution, 
  getScaleAverageTrend, 
  getAlertTimeliness 
} from '../services/statisticsService';

const router = Router();

router.get('/level-distribution', async (req: Request, res: Response) => {
  try {
    const distribution = await getLevelDistribution();
    res.json(distribution);
  } catch (err: any) {
    res.status(500).json({ error: err.message });
  }
});

router.get('/scale-trends', async (req: Request, res: Response) => {
  try {
    const days = req.query.days ? parseInt(req.query.days as string, 10) : 30;
    const trends = await getScaleAverageTrend(days);
    res.json(trends);
  } catch (err: any) {
    res.status(500).json({ error: err.message });
  }
});

router.get('/alert-timeliness', async (req: Request, res: Response) => {
  try {
    const timeliness = await getAlertTimeliness();
    res.json(timeliness);
  } catch (err: any) {
    res.status(500).json({ error: err.message });
  }
});

export default router;
