import { Router, Request, Response } from 'express';
import { analyzeRetention } from '../services/retentionService';
import { analyzePaths } from '../services/pathService';
import { exportRetentionToCSV, exportPathsToCSV } from '../services/csvService';

const router = Router();

router.post('/retention', (req: Request, res: Response) => {
  const { baselineEvent, retentionEvents, startTime, endTime, days } = req.body;

  if (!baselineEvent) {
    return res.status(400).json({ error: '基准事件不能为空' });
  }

  if (!days || !Array.isArray(days) || days.length === 0) {
    return res.status(400).json({ error: '留存天数不能为空' });
  }

  try {
    const result = analyzeRetention({
      baselineEvent: String(baselineEvent),
      retentionEvents: Array.isArray(retentionEvents) 
        ? retentionEvents.map(String) 
        : [],
      startTime: startTime ? new Date(String(startTime)) : undefined,
      endTime: endTime ? new Date(String(endTime)) : undefined,
      days: days.map(Number),
    });
    res.json(result);
  } catch (error) {
    res.status(500).json({ error: '留存分析失败' });
  }
});

router.post('/retention/export', (req: Request, res: Response) => {
  const { baselineEvent, retentionEvents, startTime, endTime, days } = req.body;

  if (!baselineEvent) {
    return res.status(400).json({ error: '基准事件不能为空' });
  }

  if (!days || !Array.isArray(days) || days.length === 0) {
    return res.status(400).json({ error: '留存天数不能为空' });
  }

  try {
    const result = analyzeRetention({
      baselineEvent: String(baselineEvent),
      retentionEvents: Array.isArray(retentionEvents) 
        ? retentionEvents.map(String) 
        : [],
      startTime: startTime ? new Date(String(startTime)) : undefined,
      endTime: endTime ? new Date(String(endTime)) : undefined,
      days: days.map(Number),
    });

    const csv = exportRetentionToCSV(result);
    res.setHeader('Content-Type', 'text/csv; charset=utf-8');
    res.setHeader(
      'Content-Disposition',
      `attachment; filename="retention_${baselineEvent}.csv"`
    );
    res.send(csv);
  } catch (error) {
    res.status(500).json({ error: '导出失败' });
  }
});

router.post('/paths', (req: Request, res: Response) => {
  const { userId, startTime, endTime, maxDepth } = req.body;

  try {
    const result = analyzePaths({
      userId: userId ? String(userId) : undefined,
      startTime: startTime ? new Date(String(startTime)) : undefined,
      endTime: endTime ? new Date(String(endTime)) : undefined,
      maxDepth: maxDepth ? Number(maxDepth) : undefined,
    });
    res.json(result);
  } catch (error) {
    res.status(500).json({ error: '路径分析失败' });
  }
});

router.post('/paths/export', (req: Request, res: Response) => {
  const { userId, startTime, endTime, maxDepth } = req.body;

  try {
    const result = analyzePaths({
      userId: userId ? String(userId) : undefined,
      startTime: startTime ? new Date(String(startTime)) : undefined,
      endTime: endTime ? new Date(String(endTime)) : undefined,
      maxDepth: maxDepth ? Number(maxDepth) : undefined,
    });

    const csv = exportPathsToCSV(result);
    res.setHeader('Content-Type', 'text/csv; charset=utf-8');
    res.setHeader('Content-Disposition', 'attachment; filename="paths.csv"');
    res.send(csv);
  } catch (error) {
    res.status(500).json({ error: '导出失败' });
  }
});

export default router;
