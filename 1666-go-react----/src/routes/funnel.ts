import express, { Request, Response } from 'express';
import { calculateFunnel } from '../services/funnelService';

const router = express.Router();

router.get('/', (req: Request, res: Response) => {
  const { startDate, endDate, salespersonId, teamId } = req.query;

  if (!startDate || !endDate) {
    return res.status(400).json({ error: 'startDate 和 endDate 是必填项' });
  }

  const result = calculateFunnel({
    startDate: String(startDate),
    endDate: String(endDate),
    salespersonId: salespersonId ? String(salespersonId) : undefined,
    teamId: teamId ? String(teamId) : undefined,
  });

  if (result.success && result.data) {
    res.json(result.data);
  } else if (result.error) {
    res.status(result.error.status).json({ error: result.error.message });
  } else {
    res.status(500).json({ error: '查询失败' });
  }
});

export default router;
