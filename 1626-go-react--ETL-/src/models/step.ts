export type DataSourceType = 'http' | 'memory';

export interface DataSourceConfig {
  type: DataSourceType;
  httpConfig?: {
    url: string;
    method?: 'GET' | 'POST';
    headers?: Record<string, string>;
    body?: any;
  };
  memoryConfig?: {
    data: any[];
  };
}

export type TransformationType = 'mapping' | 'typeConversion' | 'filter' | 'aggregation';

export interface TransformationRule {
  type: TransformationType;
  mapping?: Record<string, string>;
  typeConversion?: Record<string, string>;
  filter?: {
    field: string;
    operator: 'equals' | 'notEquals' | 'greaterThan' | 'lessThan' | 'contains';
    value: any;
  };
  aggregation?: {
    groupBy: string[];
    aggregations: {
      field: string;
      function: 'sum' | 'avg' | 'count' | 'min' | 'max';
      alias: string;
    }[];
  };
}

export interface Step {
  id: string;
  pipelineId: string;
  stepNumber: number;
  name: string;
  dataSourceId: string;
  transformations: TransformationRule[];
  targetConfig: any;
  createdAt: string;
  updatedAt: string;
}

export interface CreateStepRequest {
  name: string;
  stepNumber: number;
  dataSourceId: string;
  transformations: TransformationRule[];
  targetConfig: any;
}

export interface UpdateStepRequest {
  name?: string;
  stepNumber?: number;
  dataSourceId?: string;
  transformations?: TransformationRule[];
  targetConfig?: any;
}
