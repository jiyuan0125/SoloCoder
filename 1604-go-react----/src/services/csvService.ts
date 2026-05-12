import { Writable } from 'stream';
import { 
  FunnelAnalysisResult, 
  RetentionAnalysisResult, 
  PathAnalysisResult 
} from '../types';

const BATCH_SIZE = 10000;

function escapeCSV(value: unknown): string {
  if (value === null || value === undefined) {
    return '';
  }
  const str = String(value);
  if (str.includes(',') || str.includes('"') || str.includes('\n')) {
    return `"${str.replace(/"/g, '""')}"`;
  }
  return str;
}

function formatNumber(num: number): string {
  return num.toFixed(2);
}

interface CSVWriter {
  write(rows: string[][]): void;
  end(): void;
  getContent(): string;
}

class MemoryCSVWriter implements CSVWriter {
  private content: string = '';
  private batchBuffer: string[] = [];
  private rowCount: number = 0;

  constructor(private headers: string[]) {
    this.write([headers]);
  }

  write(rows: string[][]): void {
    for (const row of rows) {
      this.batchBuffer.push(row.map(escapeCSV).join(','));
      this.rowCount++;
      
      if (this.rowCount % BATCH_SIZE === 0) {
        this.flushBatch();
      }
    }
  }

  private flushBatch(): void {
    if (this.batchBuffer.length > 0) {
      this.content += this.batchBuffer.join('\n') + '\n';
      this.batchBuffer = [];
    }
  }

  end(): void {
    this.flushBatch();
    if (this.content.endsWith('\n')) {
      this.content = this.content.slice(0, -1);
    }
  }

  getContent(): string {
    return this.content;
  }
}

export function exportFunnelToCSV(result: FunnelAnalysisResult): string {
  const headers = [
    'step',
    'event_name',
    'unique_users',
    'conversion_rate(%)',
    'funnel_id',
    'funnel_version',
    'time_range_start',
    'time_range_end',
    'query_time',
  ];

  const writer = new MemoryCSVWriter(headers);

  for (const step of result.steps) {
    writer.write([[
      String(step.step),
      step.eventName,
      String(step.uniqueUsers),
      formatNumber(step.conversionRate),
      result.funnelId,
      String(result.funnelVersion),
      result.timeRange.start.toISOString(),
      result.timeRange.end.toISOString(),
      result.queryTime.toISOString(),
    ]]);
  }

  writer.end();
  return writer.getContent();
}

export function exportRetentionToCSV(result: RetentionAnalysisResult): string {
  const headers = [
    'day',
    'retained_users',
    'retention_rate(%)',
    'baseline_event',
    'baseline_users',
    'time_range_start',
    'time_range_end',
    'query_time',
  ];

  const writer = new MemoryCSVWriter(headers);

  for (const day of result.retention) {
    writer.write([[
      String(day.day),
      String(day.retainedUsers),
      formatNumber(day.retentionRate),
      result.baselineEvent,
      String(result.baselineUsers),
      result.timeRange.start.toISOString(),
      result.timeRange.end.toISOString(),
      result.queryTime.toISOString(),
    ]]);
  }

  writer.end();
  return writer.getContent();
}

export function exportPathsToCSV(result: PathAnalysisResult): string {
  const headers = [
    'path',
    'loop_counts',
    'user_count',
    'time_range_start',
    'time_range_end',
    'query_time',
  ];

  const writer = new MemoryCSVWriter(headers);

  for (const path of result.paths) {
    writer.write([[
      path.path.join(' > '),
      path.loopCounts.join(','),
      String(path.userCount),
      result.timeRange.start.toISOString(),
      result.timeRange.end.toISOString(),
      result.queryTime.toISOString(),
    ]]);
  }

  writer.end();
  return writer.getContent();
}
