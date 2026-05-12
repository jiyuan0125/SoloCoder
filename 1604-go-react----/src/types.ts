export interface Event {
  id: string;
  userId: string;
  eventName: string;
  occurredAt: Date;
  pagePath: string;
  properties: Record<string, unknown>;
}

export interface Funnel {
  id: string;
  name: string;
  steps: FunnelStep[];
  createdAt: Date;
  updatedAt: Date;
  version: number;
}

export interface FunnelStep {
  order: number;
  eventName: string;
  displayName?: string;
}

export interface FunnelAnalysisQuery {
  funnelId: string;
  startTime?: Date;
  endTime?: Date;
  timeWindowDays?: number;
}

export interface FunnelAnalysisResult {
  funnelId: string;
  funnelVersion: number;
  queryTime: Date;
  timeRange: { start: Date; end: Date };
  steps: FunnelStepResult[];
  overallConversionRate: number;
}

export interface FunnelStepResult {
  step: number;
  eventName: string;
  uniqueUsers: number;
  conversionRate: number;
}

export interface RetentionAnalysisQuery {
  baselineEvent: string;
  retentionEvents: string[];
  startTime?: Date;
  endTime?: Date;
  days: number[];
}

export interface RetentionAnalysisResult {
  baselineEvent: string;
  baselineUsers: number;
  queryTime: Date;
  timeRange: { start: Date; end: Date };
  retention: RetentionDayResult[];
}

export interface RetentionDayResult {
  day: number;
  retainedUsers: number;
  retentionRate: number;
}

export interface PathAnalysisQuery {
  userId?: string;
  startTime?: Date;
  endTime?: Date;
  maxDepth?: number;
}

export interface PathAnalysisResult {
  paths: PathResult[];
  totalPaths: number;
  queryTime: Date;
  timeRange: { start: Date; end: Date };
}

export interface PathResult {
  path: string[];
  loopCounts: number[];
  userCount: number;
}
