export type LogLevel = 'DEBUG' | 'INFO' | 'WARN' | 'ERROR';

export const VALID_LOG_LEVELS: LogLevel[] = ['DEBUG', 'INFO', 'WARN', 'ERROR'];

export interface LogEntry {
  id?: number;
  timestamp: number;
  level: LogLevel;
  serviceName: string;
  traceId: string;
  message: string;
  tags: string[];
}

export interface LogQueryParams {
  startTime?: number;
  endTime?: number;
  level?: LogLevel;
  serviceName?: string;
  keyword?: string;
  page?: number;
  pageSize?: number;
}

export interface AggregationStats {
  timeRange: {
    startTime: number;
    endTime: number;
    granularity: string;
  };
  levelCounts: Record<LogLevel, number>;
  serviceErrorRates: Record<string, number>;
  topErrors: Array<{ message: string; count: number }>;
  timeline: Array<{
    time: string;
    level: LogLevel;
    count: number;
  }>;
}
