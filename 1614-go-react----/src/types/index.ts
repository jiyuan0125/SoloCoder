export enum PlanStatus {
  PENDING = 'pending',
  GRAYSCALE = 'grayscale',
  FULL_RELEASE = 'full_release',
  PAUSED = 'paused',
  COMPLETED = 'completed',
  ROLLED_BACK = 'rolled_back'
}

export enum GrayStrategyType {
  RATIO = 'ratio',
  WHITELIST = 'whitelist',
  RULE = 'rule'
}

export interface GrayStrategy {
  type: GrayStrategyType;
  ratio?: number;
  whitelist?: string[];
  rules?: RuleCondition[];
}

export interface RuleCondition {
  field: string;
  operator: '==' | '!=' | '>' | '<' | '>=' | '<=' | 'contains' | 'startsWith' | 'endsWith';
  value: string | number;
}

export interface ErrorRateConfig {
  threshold: number;
  windowMinutes: number;
}

export interface CreatePlanRequest {
  name: string;
  appId: string;
  oldVersion: string;
  newVersion: string;
  strategy: GrayStrategy;
  errorRateConfig?: ErrorRateConfig;
}

export interface UpdatePlanRequest {
  name?: string;
  strategy?: GrayStrategy;
  errorRateConfig?: ErrorRateConfig;
}

export interface Plan {
  id: string;
  name: string;
  appId: string;
  oldVersion: string;
  newVersion: string;
  strategy: GrayStrategy;
  status: PlanStatus;
  errorRateConfig: ErrorRateConfig;
  createdAt: number;
  updatedAt: number;
}

export interface ErrorRecord {
  id: string;
  planId: string;
  timestamp: number;
  isError: boolean;
}

export interface StatusTransitionError {
  status: number;
  message: string;
  allowedActions: string[];
}
