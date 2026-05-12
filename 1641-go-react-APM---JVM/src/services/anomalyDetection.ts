import {
  getApplicationById,
  getMetricsInRange,
  getAllAlertRules,
  createAlert,
} from '../storage';
import { AggregatedMetrics } from '../types';

function getMetricValue(metrics: AggregatedMetrics, path: string): number | null {
  const parts = path.split('.');
  let current: any = metrics;
  
  for (const part of parts) {
    if (current === undefined || current === null) {
      return null;
    }
    current = current[part];
  }
  
  return typeof current === 'number' ? current : null;
}

function mean(values: number[]): number {
  if (values.length === 0) return 0;
  return values.reduce((a, b) => a + b, 0) / values.length;
}

function stdDev(values: number[], meanVal: number): number {
  if (values.length === 0) return 0;
  const squaredDiffs = values.map(v => Math.pow(v - meanVal, 2));
  return Math.sqrt(mean(squaredDiffs));
}

function extractMetricValues(
  metricsList: AggregatedMetrics[],
  metricPath: string
): number[] {
  const values: number[] = [];
  for (const m of metricsList) {
    const val = getMetricValue(m, metricPath);
    if (val !== null) {
      values.push(val);
    }
  }
  return values;
}

const MONITORED_METRICS = [
  'http.totalRequests',
  'http.avgResponseTime',
  'database.queryCount',
  'database.slowQueryCount',
  'database.avgQueryTime',
  'jvm.memoryUsage',
  'jvm.gcCount',
  'jvm.threadCount',
];

export function runAnomalyDetection(applicationId: string): void {
  const app = getApplicationById(applicationId);
  if (!app) return;
  
  const now = Date.now();
  
  if (app.lastReportTime && now - app.lastReportTime > 10 * 60 * 1000) {
    return;
  }
  
  const fiveMinutesAgo = now - 5 * 60 * 1000;
  const oneHourAgo = now - 60 * 60 * 1000;
  const fiveMinutesBeforeThat = oneHourAgo - 5 * 60 * 1000;
  
  const recentMetrics = getMetricsInRange(applicationId, fiveMinutesAgo, now);
  const historicalMetrics = getMetricsInRange(applicationId, fiveMinutesBeforeThat, oneHourAgo);
  
  if (historicalMetrics.length === 0) {
    return;
  }
  
  for (const metricPath of MONITORED_METRICS) {
    const recentValues = extractMetricValues(recentMetrics, metricPath);
    const historicalValues = extractMetricValues(historicalMetrics, metricPath);
    
    if (recentValues.length === 0 || historicalValues.length === 0) {
      continue;
    }
    
    const recentMean = mean(recentValues);
    const historicalMean = mean(historicalValues);
    const historicalStd = stdDev(historicalValues, historicalMean);
    
    if (historicalStd === 0) {
      continue;
    }
    
    const deviations = Math.abs(recentMean - historicalMean) / historicalStd;
    
    if (deviations > 2) {
      console.log(`[Anomaly Detected] app=${app.name} metric=${metricPath} deviations=${deviations.toFixed(2)}`);
      
      const rules = getAllAlertRules().filter(r => r.metric === metricPath);
      for (const rule of rules) {
        const shouldTrigger = checkThreshold(recentMean, rule.operator, rule.threshold);
        if (shouldTrigger) {
          createAlert(
            applicationId,
            rule.id,
            rule.name,
            `Anomaly detected for ${metricPath}: recent mean=${recentMean.toFixed(2)}, historical mean=${historicalMean.toFixed(2)}, deviations=${deviations.toFixed(2)}`,
            'warning'
          );
        }
      }
    }
  }
}

function checkThreshold(
  value: number,
  operator: '>' | '<' | '>=' | '<=' | '==' | '!=',
  threshold: number
): boolean {
  switch (operator) {
    case '>': return value > threshold;
    case '<': return value < threshold;
    case '>=': return value >= threshold;
    case '<=': return value <= threshold;
    case '==': return value === threshold;
    case '!=': return value !== threshold;
    default: return false;
  }
}

export function checkAlertRules(applicationId: string, metrics: AggregatedMetrics): void {
  const rules = getAllAlertRules();
  
  for (const rule of rules) {
    const value = getMetricValue(metrics, rule.metric);
    if (value === null) continue;
    
    if (checkThreshold(value, rule.operator, rule.threshold)) {
      const result = createAlert(
        applicationId,
        rule.id,
        rule.name,
        `Metric ${rule.metric} value ${value} ${rule.operator} ${rule.threshold}`,
        'critical'
      );
      
      if (result) {
        console.log(`[Alert Triggered] rule=${rule.name} value=${value}`);
      }
    }
  }
}
