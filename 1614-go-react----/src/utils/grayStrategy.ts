import { createHash } from 'crypto';
import { GrayStrategy, GrayStrategyType, RuleCondition } from '../types';

export function hashString(input: string): number {
  const hash = createHash('sha256').update(input).digest('hex');
  const hashNum = parseInt(hash.substring(0, 8), 16);
  return hashNum % 100;
}

export function shouldUseNewVersion(
  planId: string,
  userId: string,
  strategy: GrayStrategy,
  userAttributes: Record<string, any> = {}
): boolean {
  switch (strategy.type) {
    case GrayStrategyType.RATIO:
      return checkRatioStrategy(planId, userId, strategy.ratio || 0);
    case GrayStrategyType.WHITELIST:
      return checkWhitelistStrategy(userId, strategy.whitelist || []);
    case GrayStrategyType.RULE:
      return checkRuleStrategy(userAttributes, strategy.rules || []);
    default:
      return false;
  }
}

function checkRatioStrategy(planId: string, userId: string, ratio: number): boolean {
  const combined = `${planId}-${userId}`;
  const hashValue = hashString(combined);
  return hashValue < ratio;
}

function checkWhitelistStrategy(userId: string, whitelist: string[]): boolean {
  return whitelist.includes(userId);
}

function checkRuleStrategy(
  userAttributes: Record<string, any>,
  rules: RuleCondition[]
): boolean {
  if (rules.length === 0) return false;
  return rules.every(rule => evaluateRule(userAttributes, rule));
}

function evaluateRule(
  userAttributes: Record<string, any>,
  rule: RuleCondition
): boolean {
  const fieldValue = userAttributes[rule.field];
  if (fieldValue === undefined) return false;

  switch (rule.operator) {
    case '==':
      return fieldValue == rule.value;
    case '!=':
      return fieldValue != rule.value;
    case '>':
      return fieldValue > rule.value;
    case '<':
      return fieldValue < rule.value;
    case '>=':
      return fieldValue >= rule.value;
    case '<=':
      return fieldValue <= rule.value;
    case 'contains':
      return typeof fieldValue === 'string' && fieldValue.includes(String(rule.value));
    case 'startsWith':
      return typeof fieldValue === 'string' && fieldValue.startsWith(String(rule.value));
    case 'endsWith':
      return typeof fieldValue === 'string' && fieldValue.endsWith(String(rule.value));
    default:
      return false;
  }
}

const VALID_STRATEGY_TYPES = ['ratio', 'whitelist', 'rule'];

export function validateStrategy(strategy: GrayStrategy): { valid: boolean; error?: string } {
  if (!strategy.type) {
    return { valid: false, error: '灰度策略类型不能为空' };
  }

  if (!VALID_STRATEGY_TYPES.includes(strategy.type)) {
    return { valid: false, error: '无效的灰度策略类型' };
  }

  switch (strategy.type) {
    case 'ratio':
      if (strategy.ratio === undefined) {
        return { valid: false, error: '比例灰度必须指定比例值' };
      }
      if (strategy.ratio < 0 || strategy.ratio > 100) {
        return { valid: false, error: '灰度比例必须在 0-100 之间' };
      }
      break;
    case 'whitelist':
      if (!strategy.whitelist || !Array.isArray(strategy.whitelist)) {
        return { valid: false, error: '白名单灰度必须指定白名单列表' };
      }
      break;
    case 'rule':
      if (!strategy.rules || !Array.isArray(strategy.rules)) {
        return { valid: false, error: '规则灰度必须指定规则列表' };
      }
      break;
  }

  return { valid: true };
}
