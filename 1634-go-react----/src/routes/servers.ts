import { Router, Request, Response } from 'express';
import { dataStore } from '../store';
import { ResourceType, Metric, PredictionResult } from '../types';
import { parseTimeRange, aggregateMetrics, checkThresholdDuration, linearRegression, validateUsage } from '../utils';

const router = Router();

router.post('/', (req: Request, res: Response) => {
  const { name } = req.body;
  if (!name || typeof name !== 'string') {
    return res.status(400).json({ error: 'server name is required' });
  }

  const server = dataStore.addServer(name);
  res.status(201).json(server);
});

router.post('/:id/metrics', (req: Request, res: Response) => {
  const serverId = req.params.id;
  const { cpu, memory, disk, network } = req.body;

  if (!dataStore.getServer(serverId)) {
    return res.status(404).json({ error: 'server not found' });
  }

  const values = [cpu, memory, disk, network];
  for (const value of values) {
    if (typeof value !== 'number' || !validateUsage(value)) {
      return res.status(400).json({ error: 'usage must be between 0 and 100' });
    }
  }

  const metric: Metric = {
    serverId,
    timestamp: new Date(),
    cpu,
    memory,
    disk,
    network
  };

  dataStore.addMetric(metric);

  const resourceTypes: ResourceType[] = ['cpu', 'memory', 'disk', 'network'];
  for (const resourceType of resourceTypes) {
    const activeAlert = dataStore.getActiveAlert(serverId, resourceType);
    if (activeAlert) {
      return res.status(409).json({ error: 'alert already active for this resource' });
    }

    const recentMetrics = dataStore.getMetrics(serverId, new Date(Date.now() - 10 * 60 * 1000), new Date());
    const shouldCreateAlert = checkThresholdDuration(
      recentMetrics,
      resourceType,
      dataStore.alertConfig.threshold,
      dataStore.alertConfig.durationMinutes
    );

    if (shouldCreateAlert) {
      dataStore.createAlert(serverId, resourceType);
    }
  }

  res.status(201).json({ success: true });
});

router.get('/:id/metrics', (req: Request, res: Response) => {
  const serverId = req.params.id;
  const { timeRange, resourceType } = req.query;

  if (!dataStore.getServer(serverId)) {
    return res.status(404).json({ error: 'server not found' });
  }

  if (!timeRange) {
    return res.status(400).json({ error: 'timeRange is required' });
  }

  const { startTime, endTime, granularity } = parseTimeRange(timeRange as string);
  const metrics = dataStore.getMetrics(serverId, startTime, endTime);
  const aggregated = aggregateMetrics(
    metrics,
    granularity,
    resourceType as ResourceType
  );

  res.json({
    timeRange: timeRange as string,
    granularity,
    metrics: aggregated
  });
});

router.get('/:id/alerts', (req: Request, res: Response) => {
  const serverId = req.params.id;

  if (!dataStore.getServer(serverId)) {
    return res.status(404).json({ error: 'server not found' });
  }

  const alerts = dataStore.getAlerts(serverId);
  res.json(alerts);
});

router.get('/:id/prediction', (req: Request, res: Response) => {
  const serverId = req.params.id;

  if (!dataStore.getServer(serverId)) {
    return res.status(404).json({ error: 'server not found' });
  }

  const sevenDaysAgo = new Date(Date.now() - 7 * 24 * 60 * 60 * 1000);
  const dailyMetrics = aggregateMetrics(
    dataStore.getMetrics(serverId, sevenDaysAgo, new Date()),
    'day'
  );

  const consecutiveMissing = calculateConsecutiveMissing(dailyMetrics);
  if (consecutiveMissing > 3) {
    return res.status(400).json({
      error: `insufficient data for prediction, ${consecutiveMissing} consecutive days missing`
    });
  }

  const filledMetrics = fillMissingDays(dailyMetrics);
  if (filledMetrics.length < 7) {
    return res.status(400).json({
      error: `insufficient data for prediction, ${7 - filledMetrics.length} consecutive days missing`
    });
  }

  const resourceTypes: ResourceType[] = ['cpu', 'memory', 'disk', 'network'];
  const predictions: PredictionResult[] = [];

  for (const resourceType of resourceTypes) {
    const x = filledMetrics.map((_, i) => i);
    const y = filledMetrics.map(m => m[resourceType]);
    const { slope, intercept } = linearRegression(x, y);

    const predictedValues: { day: number; value: number }[] = [];
    let willReach90 = false;
    let daysToReach90: number | undefined;

    for (let i = 1; i <= 7; i++) {
      const predictedValue = Math.max(0, Math.min(100, slope * (filledMetrics.length + i - 1) + intercept));
      predictedValues.push({ day: i, value: predictedValue });

      if (predictedValue >= 90 && !willReach90) {
        willReach90 = true;
        daysToReach90 = i;
      }
    }

    predictions.push({
      resourceType,
      predictedValues,
      willReach90Percent: willReach90,
      daysToReach90
    });
  }

  res.json(predictions);
});

function calculateConsecutiveMissing(dailyMetrics: any[]): number {
  if (dailyMetrics.length === 0) return 7;

  const dates = dailyMetrics.map(m => new Date(m.timestamp).setHours(0, 0, 0, 0));
  let maxConsecutive = 0;
  let currentConsecutive = 0;

  const today = new Date().setHours(0, 0, 0, 0);
  const sevenDaysAgo = today - 7 * 24 * 60 * 60 * 1000;

  for (let i = sevenDaysAgo; i < today; i += 24 * 60 * 60 * 1000) {
    if (!dates.includes(i)) {
      currentConsecutive++;
      maxConsecutive = Math.max(maxConsecutive, currentConsecutive);
    } else {
      currentConsecutive = 0;
    }
  }

  return maxConsecutive;
}

function fillMissingDays(dailyMetrics: any[]): any[] {
  if (dailyMetrics.length === 0) return [];

  const filled: any[] = [];
  const metricMap = new Map<string, any>();

  for (const m of dailyMetrics) {
    const dateStr = new Date(m.timestamp).toDateString();
    metricMap.set(dateStr, m);
  }

  const today = new Date();
  const sevenDaysAgo = new Date(today.getTime() - 7 * 24 * 60 * 60 * 1000);

  let lastMetric: any | null = null;
  for (let d = new Date(sevenDaysAgo); d <= today; d.setDate(d.getDate() + 1)) {
    const dateStr = d.toDateString();
    const metric = metricMap.get(dateStr);

    if (metric) {
      filled.push(metric);
      lastMetric = metric;
    } else if (lastMetric) {
      filled.push({ ...lastMetric, timestamp: new Date(d) });
    }
  }

  return filled;
}

export default router;
