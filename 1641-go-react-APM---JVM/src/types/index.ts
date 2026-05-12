export interface Application {
  id: string;
  name: string;
  description: string;
  createdAt: number;
  lastReportTime: number | null;
}

export interface HttpMetrics {
  totalRequests: number;
  statusCodes: Record<string, number>;
  avgResponseTime: number;
}

export interface DatabaseMetrics {
  queryCount: number;
  slowQueryCount: number;
  avgQueryTime: number;
}

export interface JvmMetrics {
  memoryUsage: number;
  gcCount: number;
  threadCount: number;
}

export interface MetricsPayload {
  http: HttpMetrics;
  database: DatabaseMetrics;
  jvm: JvmMetrics;
  custom?: Record<string, number>;
  traces?: TraceData[];
}

export interface TraceData {
  traceId: string;
  spanId: string;
  parentSpanId?: string;
  serviceName: string;
  operationName: string;
  startTime: number;
  duration: number;
  tags?: Record<string, string>;
}

export interface AggregatedMetrics {
  id: string;
  applicationId: string;
  timestamp: number;
  http: HttpMetrics;
  database: DatabaseMetrics;
  jvm: JvmMetrics;
  custom: Record<string, number>;
}

export interface AlertRule {
  id: string;
  name: string;
  description: string;
  metric: string;
  operator: '>' | '<' | '>=' | '<=' | '==' | '!=';
  threshold: number;
  notificationType: 'email' | 'webhook' | 'sms';
  notificationTarget: string;
  createdAt: number;
}

export interface Alert {
  id: string;
  applicationId: string;
  ruleId: string;
  ruleName: string;
  message: string;
  severity: 'critical' | 'warning' | 'info';
  status: 'active' | 'resolved';
  triggeredAt: number;
  resolvedAt?: number;
}

export interface TopologyNode {
  id: string;
  name: string;
  healthStatus: 'healthy' | 'degraded' | 'critical';
}

export interface TopologyEdge {
  source: string;
  target: string;
  count: number;
}

export interface Topology {
  nodes: TopologyNode[];
  edges: TopologyEdge[];
}
