export enum CircuitBreakerState {
  CLOSED = 'CLOSED',
  OPEN = 'OPEN',
  HALF_OPEN = 'HALF_OPEN',
}

export enum DegradationAction {
  DEFAULT_VALUE = 'DEFAULT_VALUE',
  RETURN_CACHE = 'RETURN_CACHE',
  REJECT = 'REJECT',
}

export enum TriggerConditionType {
  ERROR_RATE = 'ERROR_RATE',
  RESPONSE_TIME = 'RESPONSE_TIME',
}

export interface TriggerCondition {
  type: TriggerConditionType;
  threshold: number;
}

export interface CircuitBreakerConfig {
  id: string;
  name: string;
  endpoint: string;
  triggerConditions: TriggerCondition[];
  degradationAction: DegradationAction;
  defaultValue?: any;
  timeWindowSeconds: number;
  openDurationSeconds: number;
  halfOpenRequestLimit: number;
  createdAt: Date;
  updatedAt: Date;
}

export interface CircuitBreaker {
  id: string;
  configId: string;
  state: CircuitBreakerState;
  stateChangedAt: Date;
  lastOpenReason?: string;
}

export interface RequestStats {
  id?: string;
  circuitBreakerId: string;
  minute: string;
  successCount: number;
  failureCount: number;
  totalResponseTimeMs: number;
}

export interface StateTransitionLog {
  id?: string;
  circuitBreakerId: string;
  fromState: CircuitBreakerState;
  toState: CircuitBreakerState;
  reason: string;
  timestamp: Date;
}

export interface CreateCircuitBreakerRequest {
  name: string;
  endpoint: string;
  triggerConditions: TriggerCondition[];
  degradationAction: DegradationAction;
  defaultValue?: any;
  timeWindowSeconds?: number;
  openDurationSeconds?: number;
  halfOpenRequestLimit?: number;
}

export interface UpdateCircuitBreakerRequest {
  name?: string;
  endpoint?: string;
  triggerConditions?: TriggerCondition[];
  degradationAction?: DegradationAction;
  defaultValue?: any;
  timeWindowSeconds?: number;
  openDurationSeconds?: number;
  halfOpenRequestLimit?: number;
}
