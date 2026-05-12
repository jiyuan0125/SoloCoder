export type ValueType = 'boolean' | 'json';

export interface Condition {
  type: 'whitelist' | 'percentage' | 'attribute' | 'timeWindow';
  enabled: boolean;
}

export interface WhitelistCondition extends Condition {
  type: 'whitelist';
  userIds: string[];
}

export interface PercentageCondition extends Condition {
  type: 'percentage';
  value: number;
}

export interface AttributeCondition extends Condition {
  type: 'attribute';
  key: string;
  operator: 'equals' | 'contains' | 'greaterThan' | 'lessThan';
  value: string | number | boolean;
}

export interface TimeWindowCondition extends Condition {
  type: 'timeWindow';
  startTime: string;
  endTime: string;
}

export type SwitchConditions = {
  whitelist?: WhitelistCondition;
  percentage?: PercentageCondition;
  attributes?: AttributeCondition[];
  timeWindow?: TimeWindowCondition;
};

export interface FeatureSwitch {
  id: number;
  key: string;
  name: string;
  description: string;
  valueType: ValueType;
  booleanValue: boolean;
  jsonValue: string | null;
  conditions: SwitchConditions;
  isDeleted: boolean;
  createdBy: string;
  createdAt: number;
  updatedBy: string;
  updatedAt: number;
  version: number;
}

export interface ServiceReference {
  id: number;
  serviceName: string;
  switchKey: string;
  isActive: boolean;
  createdAt: number;
  lastPollAt: number;
}

export interface SwitchHistory {
  id: number;
  switchKey: string;
  operation: 'create' | 'update' | 'delete';
  oldValue: string | null;
  newValue: string | null;
  operator: string;
  timestamp: number;
  version: number;
}

export interface CreateSwitchRequest {
  key: string;
  name: string;
  description?: string;
  valueType: ValueType;
  value: boolean | object;
  conditions?: SwitchConditions;
  operator: string;
}

export interface UpdateSwitchRequest {
  name?: string;
  description?: string;
  value?: boolean | object;
  conditions?: SwitchConditions;
  operator: string;
}

export interface EvaluationContext {
  userId?: string;
  attributes?: Record<string, string | number | boolean>;
}

export interface SwitchResponse {
  key: string;
  name: string;
  description: string;
  valueType: ValueType;
  value: boolean | object | null;
  conditions: SwitchConditions;
  enabled: boolean;
  deleted: boolean;
  updatedAt: number;
  updatedBy: string;
  version: number;
}

export interface FullSwitchResponse extends SwitchResponse {
  id: number;
  createdBy: string;
  createdAt: number;
}

export interface BatchResponse {
  results: Record<string, SwitchResponse | null>;
  stale: boolean;
}

export interface IncrementalUpdate {
  version: number;
  switches: FeatureSwitch[];
  deletedKeys: string[];
}
