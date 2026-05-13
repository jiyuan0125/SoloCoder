export type EnvironmentType = 'dev' | 'test' | 'staging' | 'prod';

export type EnvironmentStatus = 'creating' | 'initializing' | 'running' | 'maintenance' | 'destroyed';

export interface DatabaseConfig {
  host: string;
  port: number;
  name: string;
  username: string;
  password: string;
}

export interface MiddlewareConfig {
  [key: string]: string | number | boolean | null;
}

export interface EnvironmentResources {
  serviceInstanceCount: number;
  database: DatabaseConfig;
  middleware: MiddlewareConfig;
}

export interface EnvironmentConfig {
  [key: string]: string | number | boolean | null;
}

export interface ConfigVersion {
  version: number;
  config: EnvironmentConfig;
  timestamp: string;
}

export interface Environment {
  id: string;
  name: string;
  type: EnvironmentType;
  projectId: string;
  status: EnvironmentStatus;
  resources: EnvironmentResources;
  config: EnvironmentConfig;
  configHistory: ConfigVersion[];
  currentConfigVersion: number;
  isLocked: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface CreateEnvironmentRequest {
  name: string;
  type: EnvironmentType;
  projectId: string;
  resources: EnvironmentResources;
  config: EnvironmentConfig;
}

export interface UpdateEnvironmentRequest {
  name?: string;
  resources?: Partial<EnvironmentResources>;
  config?: EnvironmentConfig;
}

export interface UpdateStatusRequest {
  status: EnvironmentStatus;
}

export interface CloneEnvironmentRequest {
  name: string;
}

export interface RollbackRequest {
  version: number;
}

export interface RollbackResult {
  environment: Environment;
  warning?: string;
}

export class ApiError extends Error {
  statusCode: number;
  constructor(statusCode: number, message: string) {
    super(message);
    this.statusCode = statusCode;
    this.name = 'ApiError';
  }
}
