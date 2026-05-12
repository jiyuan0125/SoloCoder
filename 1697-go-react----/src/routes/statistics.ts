import { Router, Request, Response } from 'express';
import { generateStatistics } from '../services/statisticsService';

const router = Router();

router.get('/:type', (req: Request, res: Response) => {
  const { type } = req.params;

  if (type !== 'daily' && type !== 'weekly' && type !== 'monthly') {
    return res.status(400).json({ error: '无效的统计类型，请使用 daily, weekly 或 monthly' });
  }

  try {
    const stats = generateStatistics(type);
    return res.json(stats);
  } catch (error) {
    console.error('生成统计报表失败:', error);
    return res.status(500).json({ error: '服务器内部错误' });
  }
});

export default router;
