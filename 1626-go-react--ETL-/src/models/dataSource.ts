import { DataSourceConfig } from './step';

export interface DataSource {
  id: string;
  name: string;
  description?: string;
  config: DataSourceConfig;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface CreateDataSourceRequest {
  name: string;
  description?: string;
  config: DataSourceConfig;
}

export interface UpdateDataSourceRequest {
  name?: string;
  description?: string;
  config?: DataSourceConfig;
  isActive?: boolean;
}
