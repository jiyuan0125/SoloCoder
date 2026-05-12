export type ResourceType = 'cpu' | 'memory' | 'disk' | 'network';

export interface Server {
  id: string;
  name: string;
  createdAt: Date;
}

export interface Metric {
  serverId: string;
  timestamp: Date;
  cpu: number;
  memory: number;
  disk: number;
  network: number;
}

export interface Alert {
  id: string;
  serverId: string;
  resourceType: ResourceType;
  startTime: Date;
  endTime?: Date;
  threshold: number;
  active: boolean;
}

export interface MetricQueryParams {
  timeRange: string;
  resourceType?: ResourceType;
}

export interface AlertConfig {
  threshold: number;
  durationMinutes: number;
}

export interface PredictionResult {
  resourceType: ResourceType;
  predictedValues: { day: number; value: number }[];
  willReach90Percent: boolean;
  daysToReach90?: number;
}
