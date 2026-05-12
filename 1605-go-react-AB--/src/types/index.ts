export enum ExperimentStatus {
  CONFIGURING = 'configuring',
  RUNNING = 'running',
  PAUSED = 'paused',
  ENDED = 'ended'
}

export enum TrafficType {
  USER_ID = 'user_id',
  DEVICE_ID = 'device_id',
  REGION = 'region'
}

export interface ExperimentVariant {
  id: string;
  name: string;
  trafficPercentage: number;
  isControl: boolean;
}

export interface Experiment {
  id: string;
  name: string;
  status: ExperimentStatus;
  trafficType: TrafficType;
  variants: ExperimentVariant[];
  createdAt: Date;
  updatedAt: Date;
}

export interface Metric {
  id: string;
  experimentId: string;
  name: string;
  isCore: boolean;
}

export interface MetricDataPoint {
  experimentId: string;
  variantId: string;
  metricId: string;
  userKey: string;
  value: number;
}

export interface GrayscaleConfig {
  id: string;
  experimentId: string;
  steps: number[];
  currentStep: number;
  isActive: boolean;
}

export interface Assignment {
  id: string;
  experimentId: string;
  userKey: string;
  variantId: string;
  variantName: string;
  createdAt: Date;
}

export interface MetricComparison {
  metricId: string;
  metricName: string;
  control: {
    mean: number;
    sampleSize: number;
  };
  treatment: {
    mean: number;
    sampleSize: number;
    variantId: string;
    variantName: string;
  };
  absoluteDifference: number;
  relativeImprovement: number | null;
  pValue: number | null;
  isCore: boolean;
  sampleSizeWarning?: string;
}
