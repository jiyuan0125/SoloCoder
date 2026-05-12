import { ValueType, Environment } from './types';

export const VALID_ENVIRONMENTS: Environment[] = ['development', 'testing', 'production'];
export const VALID_VALUE_TYPES: ValueType[] = ['string', 'number', 'boolean', 'json'];

export interface ValidationResult {
  valid: boolean;
  error?: string;
  errorPosition?: number;
}

export function validateEnvironment(env: string): ValidationResult {
  if (!VALID_ENVIRONMENTS.includes(env as Environment)) {
    return {
      valid: false,
      error: `Invalid environment: ${env}. Must be one of: ${VALID_ENVIRONMENTS.join(', ')}`
    };
  }
  return { valid: true };
}

export function validateValueType(type: string): ValidationResult {
  if (!VALID_VALUE_TYPES.includes(type as ValueType)) {
    return {
      valid: false,
      error: `Invalid value type: ${type}. Must be one of: ${VALID_VALUE_TYPES.join(', ')}`
    };
  }
  return { valid: true };
}

export function validateValue(value: string, valueType: ValueType): ValidationResult {
  switch (valueType) {
    case 'string':
      return { valid: true };
    case 'number':
      if (value.trim() === '') {
        return { valid: false, error: 'Number value cannot be empty' };
      }
      const parsed = Number(value);
      if (isNaN(parsed)) {
        return { valid: false, error: `Invalid number: ${value}` };
      }
      return { valid: true };
    case 'boolean':
      if (value !== 'true' && value !== 'false') {
        return { valid: false, error: `Invalid boolean: ${value}. Must be 'true' or 'false'` };
      }
      return { valid: true };
    case 'json':
      try {
        JSON.parse(value);
        return { valid: true };
      } catch (e) {
        const error = e as SyntaxError;
        const match = error.message.match(/position (\d+)/);
        return {
          valid: false,
          error: `Invalid JSON: ${error.message}`,
          errorPosition: match ? parseInt(match[1]) : 0
        };
      }
    default:
      return { valid: false, error: `Unknown value type: ${valueType}` };
  }
}

export function calculateMaxVersion(versions: number[]): number {
  if (versions.length === 0) return 0;
  return Math.max(...versions);
}
