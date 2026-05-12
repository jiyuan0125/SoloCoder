import type { CreateSwitchRequest, UpdateSwitchRequest, SwitchConditions, ValueType } from '../types';
import { BadRequestError, ConflictError } from './errors';

export function validateCreateRequest(req: Partial<CreateSwitchRequest>): CreateSwitchRequest {
  if (!req.key || typeof req.key !== 'string' || req.key.trim().length === 0) {
    throw new BadRequestError('Key is required and must be a non-empty string');
  }
  if (!req.name || typeof req.name !== 'string' || req.name.trim().length === 0) {
    throw new BadRequestError('Name is required and must be a non-empty string');
  }
  if (!req.valueType || (req.valueType !== 'boolean' && req.valueType !== 'json')) {
    throw new BadRequestError('valueType must be either "boolean" or "json"');
  }
  if (req.value === undefined) {
    throw new BadRequestError('value is required');
  }
  if (req.valueType === 'boolean' && typeof req.value !== 'boolean') {
    throw new BadRequestError('value must be a boolean when valueType is "boolean"');
  }
  if (req.valueType === 'json') {
    if (typeof req.value !== 'object' || req.value === null) {
      throw new BadRequestError('value must be an object when valueType is "json"');
    }
    try {
      JSON.stringify(req.value);
    } catch (e) {
      throw new BadRequestError('Invalid JSON value: ' + (e as Error).message);
    }
  }
  if (!req.operator || typeof req.operator !== 'string') {
    throw new BadRequestError('operator is required');
  }
  if (req.conditions) {
    validateConditions(req.conditions);
  }
  return req as CreateSwitchRequest;
}

export function validateUpdateRequest(req: Partial<UpdateSwitchRequest>, valueType: ValueType): UpdateSwitchRequest {
  if (req.name !== undefined && (typeof req.name !== 'string' || req.name.trim().length === 0)) {
    throw new BadRequestError('Name must be a non-empty string');
  }
  if (req.description !== undefined && typeof req.description !== 'string') {
    throw new BadRequestError('Description must be a string');
  }
  if (req.value !== undefined) {
    if (valueType === 'boolean' && typeof req.value !== 'boolean') {
      throw new BadRequestError('value must be a boolean');
    }
    if (valueType === 'json') {
      if (typeof req.value !== 'object' || req.value === null) {
        throw new BadRequestError('value must be an object');
      }
      try {
        JSON.stringify(req.value);
      } catch (e) {
        throw new BadRequestError('Invalid JSON value: ' + (e as Error).message);
      }
    }
  }
  if (!req.operator || typeof req.operator !== 'string') {
    throw new BadRequestError('operator is required');
  }
  if (req.conditions) {
    validateConditions(req.conditions);
  }
  return req as UpdateSwitchRequest;
}

export function validateConditions(conditions: SwitchConditions): void {
  if (conditions.whitelist) {
    if (conditions.whitelist.type !== 'whitelist') {
      throw new BadRequestError('Invalid whitelist condition type');
    }
    if (!Array.isArray(conditions.whitelist.userIds)) {
      throw new BadRequestError('userIds must be an array in whitelist condition');
    }
    const seen = new Set<string>();
    for (const uid of conditions.whitelist.userIds) {
      if (typeof uid !== 'string') {
        throw new BadRequestError('All userIds must be strings');
      }
      if (seen.has(uid)) {
        throw new ConflictError(`Duplicate userId in whitelist: ${uid}`, 'DUPLICATE_USER_ID');
      }
      seen.add(uid);
    }
  }
  if (conditions.percentage) {
    if (conditions.percentage.type !== 'percentage') {
      throw new BadRequestError('Invalid percentage condition type');
    }
    if (typeof conditions.percentage.value !== 'number') {
      throw new BadRequestError('Percentage value must be a number');
    }
    if (conditions.percentage.value < 0 || conditions.percentage.value > 100) {
      throw new BadRequestError('Percentage must be between 0 and 100 inclusive', 'INVALID_PERCENTAGE');
    }
  }
  if (conditions.attributes) {
    if (!Array.isArray(conditions.attributes)) {
      throw new BadRequestError('attributes must be an array');
    }
    const validOps = ['equals', 'contains', 'greaterThan', 'lessThan'];
    for (const attr of conditions.attributes) {
      if (attr.type !== 'attribute') {
        throw new BadRequestError('Invalid attribute condition type');
      }
      if (!attr.key || typeof attr.key !== 'string') {
        throw new BadRequestError('Attribute key is required');
      }
      if (!validOps.includes(attr.operator)) {
        throw new BadRequestError(`Invalid operator: ${attr.operator}. Must be one of: ${validOps.join(', ')}`);
      }
      if (attr.value === undefined) {
        throw new BadRequestError('Attribute value is required');
      }
    }
  }
  if (conditions.timeWindow) {
    if (conditions.timeWindow.type !== 'timeWindow') {
      throw new BadRequestError('Invalid timeWindow condition type');
    }
    if (!conditions.timeWindow.startTime || !conditions.timeWindow.endTime) {
      throw new BadRequestError('startTime and endTime are required for timeWindow condition');
    }
    const start = Date.parse(conditions.timeWindow.startTime);
    const end = Date.parse(conditions.timeWindow.endTime);
    if (isNaN(start) || isNaN(end)) {
      throw new BadRequestError('Invalid date format in timeWindow. Use ISO 8601 format');
    }
    if (start >= end) {
      throw new BadRequestError('startTime must be before endTime');
    }
  }
}

export function parseJsonValue(jsonStr: string): object {
  try {
    return JSON.parse(jsonStr);
  } catch (e) {
    throw new BadRequestError('Invalid JSON value: ' + (e as Error).message);
  }
}
