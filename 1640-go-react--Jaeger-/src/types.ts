export interface Span {
  traceId: string;
  spanId: string;
  parentSpanId?: string;
  operationName: string;
  serviceName: string;
  startTime: number;
  duration: number;
  tags?: Record<string, string>;
}

export interface Trace {
  traceId: string;
  spans: Span[];
  startTime: number;
  duration: number;
}

export interface SearchParams {
  serviceName?: string;
  operationName?: string;
  startTimeMin?: number;
  startTimeMax?: number;
  minDuration?: number;
  page?: number;
  pageSize?: number;
}

export interface Dependency {
  from: string;
  to: string;
  callCount: number;
  avgDuration: number;
}
