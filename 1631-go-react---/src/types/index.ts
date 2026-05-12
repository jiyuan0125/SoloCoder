export interface Span {
  traceId: string;
  spanId: string;
  parentSpanId?: string;
  serviceName: string;
  operationName: string;
  startTime: number;
  endTime: number;
  tags?: Record<string, string | number | boolean>;
}

export interface SpanTreeNode extends Span {
  children: SpanTreeNode[];
}

export interface TraceTree {
  traceId: string;
  rootSpans: SpanTreeNode[];
  totalDuration: number;
  spanCount: number;
}

export interface ServiceStats {
  serviceName: string;
  totalSpans: number;
  avgResponseTime: number;
  errorCount: number;
  errorRate: number;
}

export interface PaginatedResult<T> {
  data: T[];
  page: number;
  pageSize: number;
  total: number;
  totalPages: number;
}

export interface TraceQueryParams {
  serviceName?: string;
  operationName?: string;
  startTime?: number;
  endTime?: number;
  page?: number;
  pageSize?: number;
}
