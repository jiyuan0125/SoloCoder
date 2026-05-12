import { Router, Request, Response } from 'express';
import * as chartCardDao from '../dao/chartCardDao';
import * as dashboardDao from '../dao/dashboardDao';
import * as metricDao from '../dao/metricDao';
import { CreateChartCardRequest, UpdateChartCardRequest } from '../types';
import { aggregateData, normalizeHeatmapData } from '../services/aggregationService';

const router = Router();

const validateChartType = (type: string): type is 'line' | 'bar' | 'pie' | 'heatmap' => {
  return ['line', 'bar', 'pie', 'heatmap'].includes(type);
};

const validateAggregation = (type: string): boolean => {
  return ['sum', 'avg', 'max', 'min', 'count', 'p50', 'p90', 'p99'].includes(type);
};

router.post('/', (req: Request, res: Response) => {
  const body = req.body as CreateChartCardRequest;

  if (!body.dashboardId || typeof body.dashboardId !== 'number') {
    return res.status(400).json({ error: '缺少或无效的字段: dashboardId' });
  }

  if (!body.title || typeof body.title !== 'string') {
    return res.status(400).json({ error: '缺少或无效的字段: title' });
  }

  if (!validateChartType(body.chartType)) {
    return res.status(400).json({ error: '缺少或无效的字段: chartType (line, bar, pie, heatmap)' });
  }

  if (!body.metricName || typeof body.metricName !== 'string') {
    return res.status(400).json({ error: '缺少或无效的字段: metricName' });
  }

  if (!validateAggregation(body.aggregation)) {
    return res.status(400).json({
      error: '缺少或无效的字段: aggregation (sum, avg, max, min, count, p50, p90, p99)',
    });
  }

  try {
    const dashboard = dashboardDao.getDashboardById(body.dashboardId);
    if (!dashboard) {
      return res.status(404).json({ error: '仪表盘不存在' });
    }

    const metric = metricDao.getMetricByName(body.metricName);
    const dataStatus: 'available' | 'missing' = metric ? 'available' : 'missing';

    const card = chartCardDao.createChartCard(
      body.dashboardId,
      body.title,
      body.chartType,
      body.metricName,
      body.aggregation,
      body.position,
      dataStatus
    );

    return res.status(201).json(card);
  } catch (err) {
    return res.status(500).json({ error: '创建图表卡片失败' });
  }
});

router.get('/:id', (req: Request, res: Response) => {
  const id = parseInt(req.params.id);
  if (isNaN(id)) {
    return res.status(400).json({ error: '无效的图表卡片 ID' });
  }

  try {
    const card = chartCardDao.getChartCardById(id);
    if (!card) {
      return res.status(404).json({ error: '图表卡片不存在' });
    }
    return res.json(card);
  } catch (err) {
    return res.status(500).json({ error: '获取图表卡片失败' });
  }
});

router.get('/dashboard/:dashboardId', (req: Request, res: Response) => {
  const dashboardId = parseInt(req.params.dashboardId);
  if (isNaN(dashboardId)) {
    return res.status(400).json({ error: '无效的仪表盘 ID' });
  }

  try {
    const dashboard = dashboardDao.getDashboardById(dashboardId);
    if (!dashboard) {
      return res.status(404).json({ error: '仪表盘不存在' });
    }

    const cards = chartCardDao.getChartCardsByDashboardId(dashboardId);
    return res.json(cards);
  } catch (err) {
    return res.status(500).json({ error: '获取图表卡片列表失败' });
  }
});

router.put('/:id', (req: Request, res: Response) => {
  const id = parseInt(req.params.id);
  if (isNaN(id)) {
    return res.status(400).json({ error: '无效的图表卡片 ID' });
  }

  const body = req.body as UpdateChartCardRequest;

  if (body.chartType !== undefined && !validateChartType(body.chartType)) {
    return res.status(400).json({ error: '无效的 chartType' });
  }

  if (body.aggregation !== undefined && !validateAggregation(body.aggregation)) {
    return res.status(400).json({ error: '无效的 aggregation' });
  }

  try {
    const card = chartCardDao.getChartCardById(id);
    if (!card) {
      return res.status(404).json({ error: '图表卡片不存在' });
    }

    let dataStatus = card.dataStatus;
    let metricName = body.metricName ?? card.metricName;

    if (body.metricName) {
      const metric = metricDao.getMetricByName(body.metricName);
      dataStatus = metric ? 'available' : 'missing';
    }

    const updated = chartCardDao.updateChartCard(id, {
      title: body.title,
      chartType: body.chartType,
      metricName: body.metricName,
      aggregation: body.aggregation,
      position: body.position,
      dataStatus,
    });

    return res.json(updated);
  } catch (err) {
    return res.status(500).json({ error: '更新图表卡片失败' });
  }
});

router.delete('/:id', (req: Request, res: Response) => {
  const id = parseInt(req.params.id);
  if (isNaN(id)) {
    return res.status(400).json({ error: '无效的图表卡片 ID' });
  }

  try {
    const card = chartCardDao.getChartCardById(id);
    if (!card) {
      return res.status(404).json({ error: '图表卡片不存在' });
    }

    const success = chartCardDao.deleteChartCard(id);
    if (!success) {
      return res.status(404).json({ error: '图表卡片不存在' });
    }

    return res.status(204).send();
  } catch (err) {
    return res.status(500).json({ error: '删除图表卡片失败' });
  }
});

router.get('/:id/data', (req: Request, res: Response) => {
  const id = parseInt(req.params.id);
  if (isNaN(id)) {
    return res.status(400).json({ error: '无效的图表卡片 ID' });
  }

  const { start, end } = req.query;

  if (!start || typeof start !== 'string' || !end || typeof end !== 'string') {
    return res.status(400).json({ error: '缺少或无效的查询参数: start 和 end (ISO 日期字符串)' });
  }

  const startDate = new Date(start);
  const endDate = new Date(end);

  if (isNaN(startDate.getTime()) || isNaN(endDate.getTime())) {
    return res.status(400).json({ error: '无效的日期格式，请使用 ISO 日期字符串' });
  }

  try {
    const card = chartCardDao.getChartCardById(id);
    if (!card) {
      return res.status(404).json({ error: '图表卡片不存在' });
    }

    const metric = metricDao.getMetricByName(card.metricName);

    if (!metric) {
      chartCardDao.updateChartCard(id, { dataStatus: 'missing' });
      return res.json({
        chartType: card.chartType,
        dataStatus: 'missing',
        data: [],
      });
    }

    chartCardDao.updateChartCard(id, { dataStatus: 'available' });

    const rawData = metricDao.getMetricData(metric.id, start, end);

    const fillZeros = card.chartType === 'line';
    let aggregatedData = aggregateData(rawData, card.aggregation, startDate, endDate, fillZeros);

    if (card.chartType === 'heatmap') {
      aggregatedData = normalizeHeatmapData(aggregatedData);
    }

    return res.json({
      chartType: card.chartType,
      dataStatus: 'available',
      data: aggregatedData,
    });
  } catch (err) {
    return res.status(500).json({ error: '获取图表数据失败' });
  }
});

export default router;
