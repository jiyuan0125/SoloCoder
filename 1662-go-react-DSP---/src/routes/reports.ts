import express, { Request, Response } from 'express';
import { getCampaignStatistics } from '../database';

const router = express.Router();

router.get('/campaigns', async (req: Request, res: Response) => {
  try {
    const stats = await getCampaignStatistics();
    res.json(stats);
  } catch (err) {
    res.status(500).json({ error: 'Internal server error' });
  }
});

export default router;
