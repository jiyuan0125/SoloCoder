import { Request, Response } from 'express';
import { Span } from './types';
import * as db from './database';

const MAX_SPANS_PER_BATCH = 500;
const DEFAULT_PAGE_SIZE = 20;

function postSpansBatch(req: Request, res: Response): void {
  const spans = req.body as Span[];

  if (!Array.isArray(spans)) {
    res.status(400).json({ error: 'Request body must be an array of spans' });
    return;
  }

  if (spans.length > MAX_SPANS_PER_BATCH) {
    res.status(400).json({ error: `Batch size exceeds limit of ${MAX_SPANS_PER_BATCH} spans` });
    return;
  }

  for (let i = 0; i < spans.length; i++) {
    const span = spans[i];
    if (!span.traceId) {
      res.status(400).json({ error: `Span at index ${i} is missing required field: traceId` });
      return;
    }
    if (!span.spanId) {
      res.status(400).json({ error: `Span at index ${i} is missing required field: spanId` });
      return;
    }
  }

  try {
    db.insertSpansBatch(spans);
    res.status(202).json({ message: 'Spans accepted', count: spans.length });
  } catch (err) {
    res.status(500).json({ error: 'Failed to store spans' });
  }
}

function getTraceById(req: Request, res: Response): void {
  const { traceId } = req.params;
  const spans = db.getSpansByTraceId(traceId);

  if (spans.length === 0) {
    res.status(404).json({ error: 'Trace not found' });
    return;
  }

  const startTime = Math.min(...spans.map(s => s.startTime));
  const endTime = Math.max(...spans.map(s => s.startTime + s.duration));

  res.json({
    traceId,
    spans,
    startTime,
    duration: endTime - startTime,
  });
}

function searchTraces(req: Request, res: Response): void {
  const {
    serviceName,
    operationName,
    startTimeMin,
    startTimeMax,
    minDuration,
    page: pageStr,
    pageSize: pageSizeStr,
  } = req.query;

  const page = pageStr ? parseInt(pageStr as string, 10) : 1;
  const pageSize = pageSizeStr ? parseInt(pageSizeStr as string, 10) : DEFAULT_PAGE_SIZE;

  const result = db.searchTraces({
    serviceName: serviceName as string | undefined,
    operationName: operationName as string | undefined,
    startTimeMin: startTimeMin ? parseInt(startTimeMin as string, 10) : undefined,
    startTimeMax: startTimeMax ? parseInt(startTimeMax as string, 10) : undefined,
    minDuration: minDuration ? parseInt(minDuration as string, 10) : undefined,
    page,
    pageSize,
  });

  res.json({
    traces: result.traces,
    total: result.total,
    page,
    pageSize,
    totalPages: Math.ceil(result.total / pageSize),
  });
}

function getSpansByTraceId(req: Request, res: Response): void {
  const { traceId } = req.params;
  const spans = db.getSpansByTraceId(traceId);

  if (spans.length === 0) {
    res.status(404).json({ error: 'Trace not found' });
    return;
  }

  res.json(spans);
}

function getSpanById(req: Request, res: Response): void {
  const { traceId, spanId } = req.params;
  const span = db.getSpanById(traceId, spanId);

  if (!span) {
    res.status(404).json({ error: 'Span not found' });
    return;
  }

  res.json(span);
}

function getDependencies(req: Request, res: Response): void {
  const deps = db.getDependencies();
  res.json({ dependencies: deps });
}

export {
  postSpansBatch,
  getTraceById,
  searchTraces,
  getSpansByTraceId,
  getSpanById,
  getDependencies,
};
