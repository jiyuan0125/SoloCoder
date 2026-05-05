import { SalaryCalculationInput, SalaryCalculationResult, SalaryRecord, SalaryComparisonResult } from './types.js';
import { ApiError } from './errors.js';

export interface ApiResponse<T = unknown> {
  success: boolean;
  data?: T;
  error?: ApiError;
}

export interface CalculateSalaryRequest extends SalaryCalculationInput {}

export interface CalculateSalaryResponse extends ApiResponse<SalaryCalculationResult> {}

export interface SaveRecordRequest extends SalaryCalculationResult {}

export interface SaveRecordResponse extends ApiResponse<SalaryRecord> {}

export interface QueryRecordsRequest {
  employeeId?: string;
  month?: string;
}

export interface QueryRecordsResponse extends ApiResponse<SalaryRecord[]> {}

export interface GetComparisonRequest {
  employeeId: string;
  month: string;
}

export interface GetComparisonResponse extends ApiResponse<SalaryComparisonResult> {}

export interface HealthCheckResponse extends ApiResponse<{ status: string; timestamp: number }> {}

export const API_ENDPOINTS = {
  HEALTH: '/health',
  CALCULATE: '/api/salary/calculate',
  SAVE: '/api/salary/save',
  QUERY: '/api/salary/query',
  COMPARISON: '/api/salary/comparison',
} as const;
