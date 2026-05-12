import { AggregationType, Granularity, AggregatedDataPoint, MetricData } from '../types';

export const calculateGranularity = (start: Date, end: Date): Granularity => {
  const diffMs = end.getTime() - start.getTime();
  const diffDays = diffMs / (1000 * 60 * 60 * 24);

  if (diffDays <= 7) {
    return 'hour';
  } else if (diffDays <= 30) {
    return 'day';
  } else {
    return 'week';
  }
};

export const truncateToGranularity = (date: Date, granularity: Granularity): Date => {
  const d = new Date(date);
  d.setMilliseconds(0);
  d.setSeconds(0);
  d.setMinutes(0);

  if (granularity === 'hour') {
    return d;
  }

  d.setHours(0);

  if (granularity === 'day') {
    return d;
  }

  d.setDate(d.getDate() - d.getDay());
  return d;
};

const percentile = (sorted: number[], p: number): number => {
  if (sorted.length < 10) {
    return sorted[Math.floor(sorted.length / 2)];
  }

  const index = (sorted.length - 1) * p;
  const lower = Math.floor(index);
  const upper = Math.ceil(index);
  const weight = index - lower;

  if (upper >= sorted.length) {
    return sorted[sorted.length - 1];
  }

  if (lower === upper) {
    return sorted[lower];
  }

  return sorted[lower] * (1 - weight) + sorted[upper] * weight;
};

const aggregateGroup = (values: number[], type: AggregationType): number => {
  if (values.length === 0) {
    return 0;
  }

  switch (type) {
    case 'sum':
      return values.reduce((a, b) => a + b, 0);
    case 'avg':
      return values.reduce((a, b) => a + b, 0) / values.length;
    case 'max':
      return Math.max(...values);
    case 'min':
      return Math.min(...values);
    case 'count':
      return values.length;
    case 'p50':
      return percentile([...values].sort((a, b) => a - b), 0.5);
    case 'p90':
      return percentile([...values].sort((a, b) => a - b), 0.9);
    case 'p99':
      return percentile([...values].sort((a, b) => a - b), 0.99);
    default:
      return 0;
  }
};

const formatTimestamp = (date: Date, granularity: Granularity): string => {
  const d = new Date(date);
  const year = d.getFullYear();
  const month = String(d.getMonth() + 1).padStart(2, '0');
  const day = String(d.getDate()).padStart(2, '0');
  const hour = String(d.getHours()).padStart(2, '0');

  switch (granularity) {
    case 'hour':
      return `${year}-${month}-${day}T${hour}:00:00`;
    case 'day':
      return `${year}-${month}-${day}`;
    case 'week':
      return `${year}-${month}-${day}`;
  }
};

export const generateTimeBuckets = (
  start: Date,
  end: Date,
  granularity: Granularity
): Date[] => {
  const buckets: Date[] = [];
  let current = truncateToGranularity(new Date(start), granularity);
  const endBucket = truncateToGranularity(new Date(end), granularity);

  while (current <= endBucket) {
    buckets.push(new Date(current));

    switch (granularity) {
      case 'hour':
        current.setHours(current.getHours() + 1);
        break;
      case 'day':
        current.setDate(current.getDate() + 1);
        break;
      case 'week':
        current.setDate(current.getDate() + 7);
        break;
    }
  }

  return buckets;
};

const isValidDate = (d: Date): boolean => {
  return !isNaN(d.getTime());
};

const parseDate = (timestamp: string): Date | null => {
  const d = new Date(timestamp);
  if (isValidDate(d)) {
    return d;
  }

  const sqliteMatch = timestamp.match(/^(\d{4})-(\d{2})-(\d{2})\s+(\d{2}):(\d{2}):(\d{2})$/);
  if (sqliteMatch) {
    const [, y, m, d, h, mi, s] = sqliteMatch;
    const iso = `${y}-${m}-${d}T${h}:${mi}:${s}`;
    const parsed = new Date(iso);
    if (isValidDate(parsed)) {
      return parsed;
    }
  }

  const dashMatch = timestamp.match(/^(\d{4})-(\d{2})-(\d{2})$/);
  if (dashMatch) {
    const [, y, m, d] = dashMatch;
    const iso = `${y}-${m}-${d}T00:00:00`;
    const parsed = new Date(iso);
    if (isValidDate(parsed)) {
      return parsed;
    }
  }

  return null;
};

export const aggregateData = (
  data: MetricData[],
  aggregationType: AggregationType,
  start: Date,
  end: Date,
  fillZeros: boolean = false
): AggregatedDataPoint[] => {
  const granularity = calculateGranularity(start, end);

  const groups: Map<string, number[]> = new Map();
  const bucketDates: Map<string, Date> = new Map();

  for (const point of data) {
    const pointDate = parseDate(point.timestamp);
    if (!pointDate) {
      continue;
    }

    const bucket = truncateToGranularity(pointDate, granularity);
    if (!isValidDate(bucket)) {
      continue;
    }

    const key = bucket.getTime().toString();

    if (!groups.has(key)) {
      groups.set(key, []);
      bucketDates.set(key, new Date(bucket));
    }
    groups.get(key)!.push(point.value);
  }

  if (!fillZeros) {
    return Array.from(groups.entries())
      .map(([key, values]) => {
        const bucketDate = bucketDates.get(key)!;
        return {
          timestamp: formatTimestamp(bucketDate, granularity),
          value: aggregateGroup(values, aggregationType),
        };
      })
      .sort((a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime());
  }

  const buckets = generateTimeBuckets(start, end, granularity);

  return buckets.map((bucket) => {
    const key = bucket.getTime().toString();
    const values = groups.get(key) || [];
    return {
      timestamp: formatTimestamp(bucket, granularity),
      value: aggregateGroup(values, aggregationType),
    };
  });
};

export const normalizeHeatmapData = (data: AggregatedDataPoint[]): AggregatedDataPoint[] => {
  if (data.length === 0) {
    return data;
  }

  const values = data.map((d) => d.value);
  const min = Math.min(...values);
  const max = Math.max(...values);

  if (max === min) {
    return data.map((d) => ({ ...d, value: 0.5 }));
  }

  return data.map((d) => ({
    ...d,
    value: (d.value - min) / (max - min),
  }));
};
