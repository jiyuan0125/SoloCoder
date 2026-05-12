export type Environment = 'development' | 'testing' | 'production';

export type ValueType = 'string' | 'number' | 'boolean' | 'json';

export interface App {
  id: number;
  name: string;
  description: string;
  createdAt: string;
  updatedAt: string;
}

export interface Config {
  id: number;
  appId: number;
  environment: Environment;
  key: string;
  value: string;
  valueType: ValueType;
  description: string;
  version: number;
  createdAt: string;
  updatedAt: string;
}

export interface ConfigHistory {
  id: number;
  configId: number;
  version: number;
  key: string;
  value: string;
  valueType: ValueType;
  description: string;
  operator: string;
  modifiedAt: string;
  oldValue: string | null;
  newValue: string | null;
  operation: 'create' | 'update' | 'delete' | 'rollback';
}

export interface AppCreateRequest {
  name: string;
  description?: string;
}

export interface AppUpdateRequest {
  name?: string;
  description?: string;
}

export interface ConfigCreateRequest {
  environment: Environment;
  key: string;
  value: string;
  valueType: ValueType;
  description?: string;
  operator: string;
}

export interface ConfigUpdateRequest {
  value?: string;
  valueType?: ValueType;
  description?: string;
  operator: string;
}

export interface BatchConfigRequest {
  environment: Environment;
  items: BatchConfigItem[];
  operator: string;
}

export interface BatchConfigItem {
  key: string;
  value: string;
  valueType?: ValueType;
  description?: string;
}

export interface BatchPublishResponse {
  success: boolean;
  failedItems?: {
    index: number;
    key: string;
    error: string;
  }[];
}

export interface PullRequest {
  environment: Environment;
  version?: number;
}

export interface PullResponse {
  version: number;
  configs: {
    key: string;
    value: string;
    valueType: ValueType;
  }[];
}
