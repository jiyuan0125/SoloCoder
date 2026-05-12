import { Router, Request, Response } from 'express';
import * as metricDao from '../dao/metricDao';
import * as chartCardDao from '../dao/chartCardDao';
import { CreateMetricRequest, UpdateMetricRequest } from '../types';

const router = Router();

const isValidName = (name: string): boolean => {
  return typeof name === 'string' && name.trim().length > 0;
};

router.post('/', (req: Request, res: Response) => {
  const body = req.body as CreateMetricRequest;

  if (!body.name || !isValidName(body.name)) {
    return res.status(400).json({ error: '指标名称不能为空或纯空格' });
  }

  if (body.data !== undefined && !Array.isArray(body.data)) {
    return res.status(400).json({ error: 'data 字段必须是数组' });
  }

  if (Array.isArray(body.data)) {
    for (const d of body.data) {
      if (typeof d.value !== 'number' || typeof d.timestamp !== 'string') {
        return res.status(400).json({ error: '数据格式错误，每条数据需要 value(数字) 和 timestamp(ISO 字符串)' });
      }
    }
  }

  try {
    const existing = metricDao.getMetricByName(body.name.trim());
    if (existing) {
      return res.status(409).json({ error: '指标名称已存在' });
    }

    const metric = metricDao.createMetric(
      body.name.trim(),
      body.description,
      body.data || []
    );

    chartCardDao.markCardsAvailableByMetricName(body.name.trim());

    return res.status(201).json(metric);
  } catch (err) {
    return res.status(500).json({ error: '创建指标失败' });
  }
});

router.get('/', (req: Request, res: Response) => {
  try {
    const metrics = metricDao.getAllMetrics();
    return res.json(metrics);
  } catch (err) {
    return res.status(500).json({ error: '获取指标列表失败' });
  }
});

router.get('/:id', (req: Request, res: Response) => {
  const id = parseInt(req.params.id);
  if (isNaN(id)) {
    return res.status(400).json({ error: '无效的指标 ID' });
  }

  try {
    const metric = metricDao.getMetricById(id);
    if (!metric) {
      return res.status(404).json({ error: '指标不存在' });
    }
    return res.json(metric);
  } catch (err) {
    return res.status(500).json({ error: '获取指标失败' });
  }
});

router.put('/:id', (req: Request, res: Response) => {
  const id = parseInt(req.params.id);
  if (isNaN(id)) {
    return res.status(400).json({ error: '无效的指标 ID' });
  }

  const body = req.body as UpdateMetricRequest;

  if (body.name !== undefined && !isValidName(body.name)) {
    return res.status(400).json({ error: '指标名称不能为空或纯空格' });
  }

  try {
    const metric = metricDao.getMetricById(id);
    if (!metric) {
      return res.status(404).json({ error: '指标不存在' });
    }

    if (body.name) {
      const existing = metricDao.getMetricByName(body.name.trim());
      if (existing && existing.id !== id) {
        return res.status(409).json({ error: '指标名称已存在' });
      }

      if (body.name.trim() !== metric.name) {
        chartCardDao.markCardsMissingByMetricName(metric.name);
      }
    }

    const updated = metricDao.updateMetric(
      id,
      body.name?.trim(),
      body.description
    );

    if (body.name) {
      chartCardDao.markCardsAvailableByMetricName(body.name.trim());
    }

    return res.json(updated);
  } catch (err) {
    return res.status(500).json({ error: '更新指标失败' });
  }
});

router.delete('/:id', (req: Request, res: Response) => {
  const id = parseInt(req.params.id);
  if (isNaN(id)) {
    return res.status(400).json({ error: '指标不存在' });
  }

  try {
    const metric = metricDao.getMetricById(id);
    if (!metric) {
      return res.status(404).json({ error: '指标不存在' });
    }

    chartCardDao.markCardsMissingByMetricName(metric.name);

    const success = metricDao.deleteMetric(id);
    if (!success) {
      return res.status(404).json({ error: '指标不存在' });
    }

    return res.status(204).send();
  } catch (err) {
    return res.status(500).json({ error: '删除指标失败' });
  }
});

router.post('/:id/data', (req: Request, res: Response) => {
  const id = parseInt(req.params.id);
  if (isNaN(id)) {
    return res.status(400).json({ error: '无效的指标 ID' });
  }

  const body = req.body as { data: Array<{ value: number; timestamp: string }> };

  if (!Array.isArray(body.data)) {
    return res.status(400).json({ error: '需要 data 数组' });
  }

  for (const d of body.data) {
    if (typeof d.value !== 'number' || typeof d.timestamp !== 'string') {
      return res.status(400).json({ error: '数据格式错误' });
    }
  }

  try {
    const metric = metricDao.getMetricById(id);
    if (!metric) {
      return res.status(404).json({ error: '指标不存在' });
    }

    metricDao.addMetricData(id, body.data);
    return res.status(201).json({ success: true });
  } catch (err) {
    return res.status(500).json({ error: '添加数据失败' });
  }
});

export default router;
