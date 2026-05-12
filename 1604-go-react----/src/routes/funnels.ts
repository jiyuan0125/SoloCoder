import { Router, Request, Response } from 'express';
import { memoryStore } from '../storage/memoryStore';
import { analyzeFunnel } from '../services/funnelService';
import { exportFunnelToCSV } from '../services/csvService';

const router = Router();

router.post('/', (req: Request, res: Response) => {
  const { name, steps } = req.body;

  if (!name || !steps || !Array.isArray(steps) || steps.length === 0) {
    return res.status(400).json({
      error: '漏斗名称和步骤不能为空',
    });
  }

  try {
    const funnel = memoryStore.addFunnel({
      name: String(name),
      steps: steps.map((step: any, idx: number) => ({
        order: step.order !== undefined ? step.order : idx + 1,
        eventName: String(step.eventName),
        displayName: step.displayName ? String(step.displayName) : undefined,
      })),
    });
    res.status(201).json(funnel);
  } catch (error) {
    res.status(500).json({ error: '创建漏斗失败' });
  }
});

router.get('/', (req: Request, res: Response) => {
  res.json(memoryStore.getFunnels());
});

router.get('/:id', (req: Request, res: Response) => {
  const funnel = memoryStore.getFunnel(req.params.id);
  if (funnel) {
    res.json(funnel);
  } else {
    res.status(404).json({ error: '漏斗不存在' });
  }
});

router.put('/:id', (req: Request, res: Response) => {
  const { name, steps } = req.body;
  
  const updates: any = {};
  if (name) updates.name = String(name);
  if (steps && Array.isArray(steps)) {
    updates.steps = steps.map((step: any, idx: number) => ({
      order: step.order !== undefined ? step.order : idx + 1,
      eventName: String(step.eventName),
      displayName: step.displayName ? String(step.displayName) : undefined,
    }));
  }

  const updated = memoryStore.updateFunnel(req.params.id, updates);
  if (updated) {
    res.json(updated);
  } else {
    res.status(404).json({ error: '漏斗不存在' });
  }
});

router.delete('/:id', (req: Request, res: Response) => {
  const deleted = memoryStore.deleteFunnel(req.params.id);
  if (deleted) {
    res.json({ success: true });
  } else {
    res.status(404).json({ error: '漏斗不存在' });
  }
});

router.post('/:id/analyze', (req: Request, res: Response) => {
  const funnel = memoryStore.getFunnel(req.params.id);
  if (!funnel) {
    return res.status(404).json({ error: '漏斗不存在' });
  }

  if (!funnel.steps || funnel.steps.length === 0) {
    return res.status(400).json({ error: '漏斗步骤序列不能为空' });
  }

  try {
    const { startTime, endTime, timeWindowDays } = req.body;
    const result = analyzeFunnel(funnel, {
      funnelId: funnel.id,
      startTime: startTime ? new Date(String(startTime)) : undefined,
      endTime: endTime ? new Date(String(endTime)) : undefined,
      timeWindowDays: timeWindowDays ? Number(timeWindowDays) : undefined,
    });
    res.json(result);
  } catch (error) {
    res.status(500).json({ error: '漏斗分析失败' });
  }
});

router.post('/:id/export', (req: Request, res: Response) => {
  const funnel = memoryStore.getFunnel(req.params.id);
  if (!funnel) {
    return res.status(404).json({ error: '漏斗不存在' });
  }

  if (!funnel.steps || funnel.steps.length === 0) {
    return res.status(400).json({ error: '漏斗步骤序列不能为空' });
  }

  try {
    const { startTime, endTime, timeWindowDays } = req.body;
    const result = analyzeFunnel(funnel, {
      funnelId: funnel.id,
      startTime: startTime ? new Date(String(startTime)) : undefined,
      endTime: endTime ? new Date(String(endTime)) : undefined,
      timeWindowDays: timeWindowDays ? Number(timeWindowDays) : undefined,
    });

    const csv = exportFunnelToCSV(result);
    res.setHeader('Content-Type', 'text/csv; charset=utf-8');
    res.setHeader(
      'Content-Disposition',
      `attachment; filename="funnel_${funnel.id}.csv"`
    );
    res.send(csv);
  } catch (error) {
    res.status(500).json({ error: '导出失败' });
  }
});

export default router;
