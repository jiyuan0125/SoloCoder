export type ChartType = 'line' | 'bar' | 'pie' | 'heatmap';
export type AggregationType = 'sum' | 'avg' | 'max' | 'min' | 'count' | 'p50' | 'p90' | 'p99';
export type Granularity = 'hour' | 'day' | 'week';

export interface Metric {
  id: number;
  name: string;
  description?: string;
  createdAt: string;
}

export interface MetricData {
  id: number;
  metricId: number;
  value: number;
  timestamp: string;
}

export interface Dashboard {
  id: number;
  name: string;
  description?: string;
  createdAt: string;
}

export interface ChartCard {
  id: number;
  dashboardId: number;
  title: string;
  chartType: ChartType;
  metricName: string;
  aggregation: AggregationType;
  position?: string;
  dataStatus: 'available' | 'missing';
  createdAt: string;
}

export interface AggregatedDataPoint {
  timestamp: string;
  value: number;
}

export interface ChartDataResponse {
  chartType: ChartType;
  dataStatus: 'available' | 'missing';
  data: AggregatedDataPoint[] | PieDataPoint[] | HeatmapPoint;
}

export interface PieDataPoint {
  label: string;
  value: number;
}

export interface HeatmapPoint {
  x: string;
  y: string;
  value: number;
}

export interface DashboardExport {
  version: string;
  dashboard: {
    name: string;
    description?: string;
  };
  charts: Array<{
    title: string;
    chartType: ChartType;
    metricName: string;
    aggregation: AggregationType;
    position?: string;
  }>;
}

export interface CreateMetricRequest {
  name: string;
  description?: string;
  data: Array<{ value: number; timestamp: string }>;
}

export interface UpdateMetricRequest {
  name?: string;
  description?: string;
}

export interface CreateDashboardRequest {
  name: string;
  description?: string;
}

export interface CreateChartCardRequest {
  dashboardId: number;
  title: string;
  chartType: ChartType;
  metricName: string;
  aggregation: AggregationType;
  position?: string;
}

export interface UpdateChartCardRequest {
  title?: string;
  chartType?: ChartType;
  metricName?: string;
  aggregation?: AggregationType;
  position?: string;
}

export interface GetChartDataQuery {
  start: string;
  end: string;
}
