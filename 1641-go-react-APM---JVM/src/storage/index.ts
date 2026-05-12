import { v4 as uuidv4 } from 'uuid';
import {
  Application,
  AggregatedMetrics,
  AlertRule,
  Alert,
  TraceData,
} from '../types';

const applications: Map<string, Application> = new Map();
const applicationNames: Set<string> = new Set();
const metrics: Map<string, AggregatedMetrics[]> = new Map();
const alertRules: Map<string, AlertRule> = new Map();
const alerts: Map<string, Alert[]> = new Map();
const traces: TraceData[] = [];

export function createApplication(name: string, description: string): Application {
  const id = uuidv4();
  const app: Application = {
    id,
    name,
    description,
    createdAt: Date.now(),
    lastReportTime: null,
  };
  applications.set(id, app);
  applicationNames.add(name);
  metrics.set(id, []);
  alerts.set(id, []);
  return app;
}

export function applicationExistsByName(name: string): boolean {
  return applicationNames.has(name);
}

export function getApplicationById(id: string): Application | undefined {
  return applications.get(id);
}

export function getAllApplications(): Application[] {
  return Array.from(applications.values());
}

export function updateApplicationLastReport(id: string): void {
  const app = applications.get(id);
  if (app) {
    app.lastReportTime = Date.now();
  }
}

export function getMinuteBucket(timestamp: number): number {
  return Math.floor(timestamp / 60000) * 60000;
}

export function addMetrics(
  applicationId: string,
  timestamp: number,
  http: any,
  database: any,
  jvm: any,
  custom: Record<string, number>
): AggregatedMetrics {
  const bucket = getMinuteBucket(timestamp);
  const appMetrics = metrics.get(applicationId) || [];
  
  const existing = appMetrics.find(m => m.timestamp === bucket);
  
  if (existing) {
    existing.http.totalRequests += http.totalRequests;
    Object.keys(http.statusCodes).forEach(code => {
      existing.http.statusCodes[code] = (existing.http.statusCodes[code] || 0) + http.statusCodes[code];
    });
    const totalSamples = existing.http.totalRequests;
    existing.http.avgResponseTime = (
      existing.http.avgResponseTime * (totalSamples - http.totalRequests) +
      http.avgResponseTime * http.totalRequests
    ) / totalSamples;
    
    existing.database.queryCount += database.queryCount;
    existing.database.slowQueryCount += database.slowQueryCount;
    const dbTotal = existing.database.queryCount;
    existing.database.avgQueryTime = (
      existing.database.avgQueryTime * (dbTotal - database.queryCount) +
      database.avgQueryTime * database.queryCount
    ) / dbTotal;
    
    existing.jvm.memoryUsage = Math.max(existing.jvm.memoryUsage, jvm.memoryUsage);
    existing.jvm.gcCount += jvm.gcCount;
    existing.jvm.threadCount = Math.max(existing.jvm.threadCount, jvm.threadCount);
    
    Object.keys(custom).forEach(key => {
      existing.custom[key] = (existing.custom[key] || 0) + custom[key];
    });
    
    return existing;
  } else {
    const newMetrics: AggregatedMetrics = {
      id: uuidv4(),
      applicationId,
      timestamp: bucket,
      http: {
        totalRequests: http.totalRequests,
        statusCodes: { ...http.statusCodes },
        avgResponseTime: http.avgResponseTime,
      },
      database: {
        queryCount: database.queryCount,
        slowQueryCount: database.slowQueryCount,
        avgQueryTime: database.avgQueryTime,
      },
      jvm: {
        memoryUsage: jvm.memoryUsage,
        gcCount: jvm.gcCount,
        threadCount: jvm.threadCount,
      },
      custom: { ...custom },
    };
    
    appMetrics.push(newMetrics);
    metrics.set(applicationId, appMetrics);
    return newMetrics;
  }
}

export function getMetricsInRange(
  applicationId: string,
  startTime: number,
  endTime: number
): AggregatedMetrics[] {
  const appMetrics = metrics.get(applicationId) || [];
  return appMetrics.filter(m => m.timestamp >= startTime && m.timestamp <= endTime);
}

export function getLatestMetrics(applicationId: string): AggregatedMetrics | undefined {
  const appMetrics = metrics.get(applicationId) || [];
  return appMetrics.length > 0 ? appMetrics[appMetrics.length - 1] : undefined;
}

export function createAlertRule(
  name: string,
  description: string,
  metric: string,
  operator: AlertRule['operator'],
  threshold: number,
  notificationType: AlertRule['notificationType'],
  notificationTarget: string
): AlertRule {
  const rule: AlertRule = {
    id: uuidv4(),
    name,
    description,
    metric,
    operator,
    threshold,
    notificationType,
    notificationTarget,
    createdAt: Date.now(),
  };
  alertRules.set(rule.id, rule);
  return rule;
}

export function getAllAlertRules(): AlertRule[] {
  return Array.from(alertRules.values());
}

export function getAlertRuleById(id: string): AlertRule | undefined {
  return alertRules.get(id);
}

export function createAlert(
  applicationId: string,
  ruleId: string,
  ruleName: string,
  message: string,
  severity: Alert['severity']
): Alert | null {
  const appAlerts = alerts.get(applicationId) || [];
  const existingActive = appAlerts.find(
    a => a.ruleId === ruleId && a.status === 'active'
  );
  
  if (existingActive) {
    return null;
  }
  
  const alert: Alert = {
    id: uuidv4(),
    applicationId,
    ruleId,
    ruleName,
    message,
    severity,
    status: 'active',
    triggeredAt: Date.now(),
  };
  
  appAlerts.push(alert);
  alerts.set(applicationId, appAlerts);
  return alert;
}

export function resolveAlert(alertId: string, applicationId: string): void {
  const appAlerts = alerts.get(applicationId) || [];
  const alert = appAlerts.find(a => a.id === alertId);
  if (alert && alert.status === 'active') {
    alert.status = 'resolved';
    alert.resolvedAt = Date.now();
  }
}

export function getAlertsByApplication(applicationId: string): Alert[] {
  return alerts.get(applicationId) || [];
}

export function addTraces(newTraces: TraceData[]): void {
  traces.push(...newTraces);
}

export function getAllTraces(): TraceData[] {
  return traces;
}

export { applications, metrics, alertRules, alerts };
