import { EnvironmentType } from '../types';

export const ENVIRONMENT_QUOTAS: Record<EnvironmentType, number> = {
  dev: 3,
  test: 3,
  staging: 2,
  prod: 1,
};

export const STATE_TRANSITIONS: Record<string, string[]> = {
  creating: ['initializing'],
  initializing: ['running'],
  running: ['maintenance'],
  maintenance: ['running', 'destroyed'],
  destroyed: [],
};

export const VALID_ENVIRONMENT_TYPES: EnvironmentType[] = ['dev', 'test', 'staging', 'prod'];

export const VALID_ENVIRONMENT_STATUSES: string[] = [
  'creating',
  'initializing',
  'running',
  'maintenance',
  'destroyed',
];
