import type { SwitchConditions, EvaluationContext } from '../types';
import crypto from 'crypto';

export function evaluateConditions(conditions: SwitchConditions, context: EvaluationContext): boolean {
  const result = evaluateAllAnd(conditions, context);
  return result;
}

function evaluateAllAnd(conditions: SwitchConditions, context: EvaluationContext): boolean {
  if (conditions.whitelist && conditions.whitelist.enabled) {
    if (!context.userId || !conditions.whitelist.userIds.includes(context.userId)) {
      return false;
    }
  }
  if (conditions.percentage && conditions.percentage.enabled) {
    if (!context.userId) return false;
    const hash = crypto.createHash('sha256').update(context.userId).digest();
    const value = hash.readUInt32BE(0) % 100;
    if (value >= conditions.percentage.value) {
      return false;
    }
  }
  if (conditions.attributes && conditions.attributes.length > 0) {
    for (const attr of conditions.attributes) {
      if (!attr.enabled) continue;
      const attrValue = context.attributes?.[attr.key];
      if (attrValue === undefined) return false;
      if (!compareValues(attrValue, attr.value, attr.operator)) {
        return false;
      }
    }
  }
  if (conditions.timeWindow && conditions.timeWindow.enabled) {
    const now = Date.now();
    const start = Date.parse(conditions.timeWindow.startTime);
    const end = Date.parse(conditions.timeWindow.endTime);
    if (now < start || now > end) {
      return false;
    }
  }
  return true;
}

function compareValues(actual: string | number | boolean, expected: string | number | boolean, op: string): boolean {
  switch (op) {
    case 'equals':
      return actual === expected;
    case 'contains':
      if (typeof actual !== 'string' || typeof expected !== 'string') return false;
      return actual.toLowerCase().includes(expected.toLowerCase());
    case 'greaterThan':
      if (typeof actual !== 'number' || typeof expected !== 'number') return false;
      return actual > expected;
    case 'lessThan':
      if (typeof actual !== 'number' || typeof expected !== 'number') return false;
      return actual < expected;
    default:
      return false;
  }
}

export function hasEnabledConditions(conditions: SwitchConditions): boolean {
  if (conditions.whitelist?.enabled) return true;
  if (conditions.percentage?.enabled) return true;
  if (conditions.attributes?.some(a => a.enabled)) return true;
  if (conditions.timeWindow?.enabled) return true;
  return false;
}
