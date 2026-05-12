import { Router, Request, Response } from 'express';
import * as dashboardDao from '../dao/dashboardDao';
import * as chartCardDao from '../dao/chartCardDao';
import * as metricDao from '../dao/metricDao';
import { CreateDashboardRequest, DashboardExport } from '../types';

const router = Router();

const validateChartType = (type: string): type is 'line' | 'bar' | 'pie' | 'heatmap' => {
  return ['line', 'bar', 'pie', 'heatmap'].includes(type);
};

const validateAggregation = (type: string): boolean => {
  return ['sum', 'avg', 'max', 'min', 'count', 'p50', 'p90', 'p99'].includes(type);
};

router.post('/', (req: Request, res: Response) => {
  const body = req.body as CreateDashboardRequest;

  if (!body.name || typeof body.name !== 'string' || body.name.trim().length === 0) {
    return res.status(400).json({ error: '仪表盘名称不能为空' });
  }

  try {
    const dashboard = dashboardDao.createDashboard(
      body.name.trim(),
      body.description
    );
    return res.status(201).json(dashboard);
  } catch (err) {
    return res.status(500).json({ error: '创建仪表盘失败' });
  }
});

router.get('/', (req: Request, res: Response) => {
  try {
    const dashboards = dashboardDao.getAllDashboards();
    return res.json(dashboards);
  } catch (err) {
    return res.status(500).json({ error: '获取仪表盘列表失败' });
  }
});

router.get('/:id', (req: Request, res: Response) => {
  const id = parseInt(req.params.id);
  if (isNaN(id)) {
    return res.status(400).json({ error: '无效的仪表盘 ID' });
  }

  try {
    const dashboard = dashboardDao.getDashboardById(id);
    if (!dashboard) {
      return res.status(404).json({ error: '仪表盘不存在' });
    }
    return res.json(dashboard);
  } catch (err) {
    return res.status(500).json({ error: '获取仪表盘失败' });
  }
});

router.put('/:id', (req: Request, res: Response) => {
  const id = parseInt(req.params.id);
  if (isNaN(id)) {
    return res.status(400).json({ error: '无效的仪表盘 ID' });
  }

  const body = req.body as CreateDashboardRequest;

  if (body.name !== undefined && (typeof body.name !== 'string' || body.name.trim().length === 0)) {
    return res.status(400).json({ error: '仪表盘名称不能为空' });
  }

  try {
    const dashboard = dashboardDao.getDashboardById(id);
    if (!dashboard) {
      return res.status(404).json({ error: '仪表盘不存在' });
    }

    const updated = dashboardDao.updateDashboard(
      id,
      body.name?.trim(),
      body.description
    );
    return res.json(updated);
  } catch (err) {
    return res.status(500).json({ error: '更新仪表盘失败' });
  }
});

router.delete('/:id', (req: Request, res: Response) => {
  const id = parseInt(req.params.id);
  if (isNaN(id)) {
    return res.status(400).json({ error: '无效的仪表盘 ID' });
  }

  try {
    const dashboard = dashboardDao.getDashboardById(id);
    if (!dashboard) {
      return res.status(404).json({ error: '仪表盘不存在' });
    }

    const cardCount = dashboardDao.countDashboardCards(id);
    if (cardCount > 0) {
      return res.status(400).json({
        error: `仪表盘上仍有 ${cardCount} 个图表卡片，不能删除`,
      });
    }

    const success = dashboardDao.deleteDashboard(id);
    if (!success) {
      return res.status(404).json({ error: '仪表盘不存在' });
    }

    return res.status(204).send();
  } catch (err) {
    return res.status(500).json({ error: '删除仪表盘失败' });
  }
});

router.get('/:id/export', (req: Request, res: Response) => {
  const id = parseInt(req.params.id);
  if (isNaN(id)) {
    return res.status(400).json({ error: '无效的仪表盘 ID' });
  }

  try {
    const dashboard = dashboardDao.getDashboardById(id);
    if (!dashboard) {
      return res.status(404).json({ error: '仪表盘不存在' });
    }

    const cards = chartCardDao.getChartCardsByDashboardId(id);

    const exportData: DashboardExport = {
      version: '1.0',
      dashboard: {
        name: dashboard.name,
        description: dashboard.description,
      },
      charts: cards.map((card) => ({
        title: card.title,
        chartType: card.chartType,
        metricName: card.metricName,
        aggregation: card.aggregation,
        position: card.position,
      })),
    };

    return res.json(exportData);
  } catch (err) {
    return res.status(500).json({ error: '导出仪表盘失败' });
  }
});

router.post('/import', (req: Request, res: Response) => {
  const body = req.body;

  if (typeof body !== 'object' || body === null) {
    return res.status(400).json({ error: 'JSON 根节点必须是对象' });
  }

  const data = body as DashboardExport;

  if (typeof data.version !== 'string') {
    return res.status(400).json({ error: '缺少或无效的字段: version (应为字符串)' });
  }

  if (typeof data.dashboard !== 'object' || data.dashboard === null) {
    return res.status(400).json({ error: '缺少或无效的字段: dashboard (应为对象)' });
  }

  if (typeof data.dashboard.name !== 'string' || data.dashboard.name.trim().length === 0) {
    return res.status(400).json({ error: '缺少或无效的字段: dashboard.name (应为非空字符串)' });
  }

  if (!Array.isArray(data.charts)) {
    return res.status(400).json({ error: '缺少或无效的字段: charts (应为数组)' });
  }

  for (let i = 0; i < data.charts.length; i++) {
    const chart = data.charts[i];
    if (typeof chart !== 'object' || chart === null) {
      return res.status(400).json({ error: `charts[${i}] 应为对象` });
    }

    if (typeof chart.title !== 'string') {
      return res.status(400).json({ error: `charts[${i}].title 应为字符串` });
    }

    if (!validateChartType(chart.chartType)) {
      return res.status(400).json({
        error: `charts[${i}].chartType 无效，应为: line, bar, pie, heatmap 之一`,
      });
    }

    if (typeof chart.metricName !== 'string') {
      return res.status(400).json({ error: `charts[${i}].metricName 应为字符串` });
    }

    if (!validateAggregation(chart.aggregation)) {
      return res.status(400).json({
        error: `charts[${i}].aggregation 无效，应为: sum, avg, max, min, count, p50, p90, p99 之一`,
      });
    }
  }

  try {
    const dashboard = dashboardDao.createDashboard(
      data.dashboard.name.trim(),
      data.dashboard.description
    );

    for (const chart of data.charts) {
      const metric = metricDao.getMetricByName(chart.metricName);
      const dataStatus: 'available' | 'missing' = metric ? 'available' : 'missing';

      chartCardDao.createChartCard(
        dashboard.id,
        chart.title,
        chart.chartType,
        chart.metricName,
        chart.aggregation,
        chart.position,
        dataStatus
      );
    }

    return res.status(201).json(dashboard);
  } catch (err) {
    return res.status(500).json({ error: '导入仪表盘失败' });
  }
});

export default router;
