export interface Service {
  id: string;
  name: string;
  description?: string;
  base_url?: string;
  created_at: string;
  updated_at: string;
  has_swagger: boolean;
}

export interface ApiEndpoint {
  id: string;
  service_id: string;
  path: string;
  method: string;
  summary?: string;
  description?: string;
  tags: string;
  parameters: string;
  request_body?: string;
  responses: string;
  created_at: string;
}

export interface DebugRecord {
  id: string;
  service_id: string;
  api_id: string;
  request: string;
  response: string;
  status_code: number;
  duration_ms: number;
  created_at: string;
}

export interface Config {
  proxy_timeout: number;
  max_request_body: number;
}

export interface SkippedApi {
  path: string;
  method: string;
  reason: string;
}

export interface ParameterInfo {
  name: string;
  in: string;
  required: boolean;
  type?: string;
  schema?: any;
  description?: string;
}

export interface RequestBodyInfo {
  description?: string;
  content?: Record<string, any>;
  required?: boolean;
}

export interface ResponseInfo {
  statusCode: string;
  description?: string;
  content?: Record<string, any>;
}

export interface AnalysisResult {
  tags: Record<string, number>;
  requiredParamsByType: Record<string, number>;
  statusCodes: Record<string, number>;
}
