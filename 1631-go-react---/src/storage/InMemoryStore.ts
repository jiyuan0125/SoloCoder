import { Span, TraceTree, SpanTreeNode, ServiceStats, TraceQueryParams, PaginatedResult } from '../types';

const SEVEN_DAYS_MS = 7 * 24 * 60 * 60 * 1000;

export class InMemoryStore {
  private spans: Map<string, Map<string, Span>> = new Map();
  private writeLock = false;
  private pendingWrites: Span[] = [];

  private getTraceMap(traceId: string): Map<string, Span> {
    let traceMap = this.spans.get(traceId);
    if (!traceMap) {
      traceMap = new Map();
      this.spans.set(traceId, traceMap);
    }
    return traceMap;
  }

  private processPendingWrites(): void {
    while (this.pendingWrites.length > 0) {
      const span = this.pendingWrites.shift()!;
      this.doWrite(span);
    }
  }

  private doWrite(span: Span): void {
    const traceMap = this.getTraceMap(span.traceId);
    traceMap.set(span.spanId, span);
  }

  addSpan(span: Span): void {
    if (this.writeLock) {
      this.pendingWrites.push(span);
    } else {
      this.doWrite(span);
    }
  }

  getSpansByTraceId(traceId: string): Span[] {
    const traceMap = this.spans.get(traceId);
    if (!traceMap) {
      return [];
    }
    return Array.from(traceMap.values());
  }

  getTraceTree(traceId: string): TraceTree | null {
    const spans = this.getSpansByTraceId(traceId);
    if (spans.length === 0) {
      return null;
    }

    const spanMap = new Map<string, SpanTreeNode>();
    const allNodes: SpanTreeNode[] = spans.map(span => ({
      ...span,
      children: []
    }));

    allNodes.forEach(node => {
      spanMap.set(node.spanId, node);
    });

    const rootSpans: SpanTreeNode[] = [];
    allNodes.forEach(node => {
      if (node.parentSpanId) {
        const parent = spanMap.get(node.parentSpanId);
        if (parent) {
          parent.children.push(node);
        } else {
          rootSpans.push(node);
        }
      } else {
        rootSpans.push(node);
      }
    });

    const sortNode = (node: SpanTreeNode): void => {
      node.children.sort((a, b) => a.startTime - b.startTime);
      node.children.forEach(sortNode);
    };

    rootSpans.sort((a, b) => a.startTime - b.startTime);
    rootSpans.forEach(sortNode);

    let minStartTime = Infinity;
    let maxEndTime = -Infinity;
    spans.forEach(span => {
      if (span.startTime < minStartTime) {
        minStartTime = span.startTime;
      }
      if (span.endTime > maxEndTime) {
        maxEndTime = span.endTime;
      }
    });

    return {
      traceId,
      rootSpans,
      totalDuration: maxEndTime - minStartTime,
      spanCount: spans.length
    };
  }

  queryTraces(params: TraceQueryParams): PaginatedResult<string> {
    const {
      serviceName,
      operationName,
      startTime,
      endTime,
      page = 1,
      pageSize = 50
    } = params;

    const traceIds: string[] = [];

    for (const [traceId, traceMap] of this.spans.entries()) {
      const spans = Array.from(traceMap.values());
      const match = spans.some(span => {
        if (serviceName && !span.serviceName.includes(serviceName)) {
          return false;
        }
        if (operationName && !span.operationName.includes(operationName)) {
          return false;
        }
        if (startTime !== undefined && span.endTime < startTime) {
          return false;
        }
        if (endTime !== undefined && span.startTime > endTime) {
          return false;
        }
        return true;
      });

      if (match) {
        traceIds.push(traceId);
      }
    }

    traceIds.sort((a, b) => {
      const spansA = this.spans.get(a);
      const spansB = this.spans.get(b);
      const minStartA = Array.from(spansA!.values()).reduce((min, s) => Math.min(min, s.startTime), Infinity);
      const minStartB = Array.from(spansB!.values()).reduce((min, s) => Math.min(min, s.startTime), Infinity);
      return minStartB - minStartA;
    });

    const total = traceIds.length;
    const totalPages = Math.ceil(total / pageSize);
    
    if (page < 1 || page > totalPages) {
      return {
        data: [],
        page,
        pageSize,
        total,
        totalPages
      };
    }

    const start = (page - 1) * pageSize;
    const data = traceIds.slice(start, start + pageSize);

    return {
      data,
      page,
      pageSize,
      total,
      totalPages
    };
  }

  getSpansPaginated(traceId: string, page: number = 1, pageSize: number = 50): PaginatedResult<Span> {
    const spans = this.getSpansByTraceId(traceId);
    spans.sort((a, b) => a.startTime - b.startTime);

    const total = spans.length;
    const totalPages = Math.ceil(total / pageSize);

    if (page < 1 || page > totalPages) {
      return {
        data: [],
        page,
        pageSize,
        total,
        totalPages
      };
    }

    const start = (page - 1) * pageSize;
    const data = spans.slice(start, start + pageSize);

    return {
      data,
      page,
      pageSize,
      total,
      totalPages
    };
  }

  getServiceStats(): ServiceStats[] {
    const statsMap = new Map<string, {
      totalSpans: number;
      totalDuration: number;
      errorCount: number;
    }>();

    for (const traceMap of this.spans.values()) {
      for (const span of traceMap.values()) {
        const stats = statsMap.get(span.serviceName) || {
          totalSpans: 0,
          totalDuration: 0,
          errorCount: 0
        };

        stats.totalSpans++;
        stats.totalDuration += (span.endTime - span.startTime);

        if (span.tags?.error === true) {
          stats.errorCount++;
        }

        statsMap.set(span.serviceName, stats);
      }
    }

    return Array.from(statsMap.entries()).map(([serviceName, stats]) => ({
      serviceName,
      totalSpans: stats.totalSpans,
      avgResponseTime: stats.totalSpans > 0 ? stats.totalDuration / stats.totalSpans : 0,
      errorCount: stats.errorCount,
      errorRate: stats.totalSpans > 0 ? stats.errorCount / stats.totalSpans : 0
    }));
  }

  getSlowTraces(threshold: number): TraceTree[] {
    const slowTraces: TraceTree[] = [];

    for (const traceId of this.spans.keys()) {
      const tree = this.getTraceTree(traceId);
      if (tree && tree.totalDuration >= threshold) {
        slowTraces.push(tree);
      }
    }

    slowTraces.sort((a, b) => b.totalDuration - a.totalDuration);
    return slowTraces;
  }

  cleanupExpiredData(cutoffTime: number): number {
    this.writeLock = true;
    
    try {
      let removedCount = 0;
      const expiredTraceIds: string[] = [];

      for (const [traceId, traceMap] of this.spans.entries()) {
        const spans = Array.from(traceMap.values());
        const maxEndTime = Math.max(...spans.map(s => s.endTime));

        if (maxEndTime < cutoffTime - SEVEN_DAYS_MS) {
          expiredTraceIds.push(traceId);
          removedCount += spans.length;
        }
      }

      expiredTraceIds.forEach(traceId => {
        this.spans.delete(traceId);
      });

      return removedCount;
    } finally {
      this.writeLock = false;
      this.processPendingWrites();
    }
  }

  hasTrace(traceId: string): boolean {
    return this.spans.has(traceId);
  }
}

export const store = new InMemoryStore();
