import { Metric, ResourceType } from './types';

export function parseTimeRange(timeRange: string): { startTime: Date; endTime: Date; granularity: string } {
  const endTime = new Date();
  let startTime: Date;
  let granularity: string;

  switch (timeRange) {
    case '1h':
      startTime = new Date(endTime.getTime() - 60 * 60 * 1000);
      granularity = 'minute';
      break;
    case '1d':
      startTime = new Date(endTime.getTime() - 24 * 60 * 60 * 1000);
      granularity = '5min';
      break;
    case '7d':
      startTime = new Date(endTime.getTime() - 7 * 24 * 60 * 60 * 1000);
      granularity = 'hour';
      break;
    case '30d':
    case 'all':
    default:
      startTime = new Date(endTime.getTime() - 30 * 24 * 60 * 60 * 1000);
      granularity = 'day';
      break;
  }

  return { startTime, endTime, granularity };
}

export function aggregateMetrics(metrics: Metric[], granularity: string, resourceType?: ResourceType): any[] {
  if (metrics.length === 0) return [];

  const grouped: Map<string, Metric[]> = new Map();

  for (const metric of metrics) {
    const key = getAggregationKey(metric.timestamp, granularity);
    if (!grouped.has(key)) {
      grouped.set(key, []);
    }
    grouped.get(key)!.push(metric);
  }

  const results: any[] = [];
  for (const [timestampKey, groupMetrics] of grouped) {
    const timestamp = new Date(timestampKey);
    if (resourceType) {
      const values = groupMetrics.map(m => m[resourceType]);
      const avgValue = values.reduce((a, b) => a + b, 0) / values.length;
      results.push({ timestamp, [resourceType]: avgValue });
    } else {
      const cpuAvg = groupMetrics.map(m => m.cpu).reduce((a, b) => a + b, 0) / groupMetrics.length;
      const memoryAvg = groupMetrics.map(m => m.memory).reduce((a, b) => a + b, 0) / groupMetrics.length;
      const diskAvg = groupMetrics.map(m => m.disk).reduce((a, b) => a + b, 0) / groupMetrics.length;
      const networkAvg = groupMetrics.map(m => m.network).reduce((a, b) => a + b, 0) / groupMetrics.length;
      results.push({ timestamp, cpu: cpuAvg, memory: memoryAvg, disk: diskAvg, network: networkAvg });
    }
  }

  return results.sort((a, b) => a.timestamp.getTime() - b.timestamp.getTime());
}

function getAggregationKey(timestamp: Date, granularity: string): string {
  const d = new Date(timestamp);
  
  switch (granularity) {
    case 'minute':
      return new Date(d.getFullYear(), d.getMonth(), d.getDate(), d.getHours(), d.getMinutes()).toISOString();
    case '5min':
      const fiveMinuteBlock = Math.floor(d.getMinutes() / 5) * 5;
      return new Date(d.getFullYear(), d.getMonth(), d.getDate(), d.getHours(), fiveMinuteBlock).toISOString();
    case 'hour':
      return new Date(d.getFullYear(), d.getMonth(), d.getDate(), d.getHours()).toISOString();
    case 'day':
    default:
      return new Date(d.getFullYear(), d.getMonth(), d.getDate()).toISOString();
  }
}

export function checkThresholdDuration(metrics: Metric[], resourceType: ResourceType, threshold: number, durationMinutes: number): boolean {
  if (metrics.length === 0) return false;

  const now = metrics[metrics.length - 1].timestamp;
  const checkStartTime = new Date(now.getTime() - durationMinutes * 60 * 1000);

  const recentMetrics = metrics.filter(m => m.timestamp >= checkStartTime);
  if (recentMetrics.length === 0) return false;

  return recentMetrics.every(m => m[resourceType] >= threshold);
}

export function linearRegression(x: number[], y: number[]): { slope: number; intercept: number } {
  if (x.length === 0 || x.length !== y.length) {
    return { slope: 0, intercept: 0 };
  }

  const n = x.length;
  const sumX = x.reduce((a, b) => a + b, 0);
  const sumY = y.reduce((a, b) => a + b, 0);
  const sumXY = x.reduce((acc, xi, i) => acc + xi * y[i], 0);
  const sumX2 = x.reduce((a, b) => a + b * b, 0);

  const slope = (n * sumXY - sumX * sumY) / (n * sumX2 - sumX * sumX);
  const intercept = (sumY - slope * sumX) / n;

  return { slope, intercept };
}

export function validateUsage(value: number): boolean {
  return value >= 0 && value <= 100;
}
