import { v4 as uuidv4 } from 'uuid';
import { ruleDAO, conditionDAO } from '../daos/ruleDAO';
import { Rule, RuleType, RuleCondition } from '../types';
import { validateCIDR, ipMatchesPattern } from '../utils/ipUtils';
import { isExpired } from '../utils/timeUtils';

export interface CreateRuleRequest {
  type: RuleType;
  ipPattern: string;
  expiresAt?: number;
}

export interface UpdateRuleRequest {
  type?: RuleType;
  ipPattern?: string;
  enabled?: boolean;
  expiresAt?: number;
}

export enum RuleValidationError {
  INVALID_CIDR = 'INVALID_CIDR',
  RULE_CONFLICT = 'RULE_CONFLICT'
}

export interface CreateRuleResult {
  success: boolean;
  rule?: Rule;
  error?: RuleValidationError;
  conflictRule?: Rule;
}

export interface UpdateRuleResult {
  success: boolean;
  rule?: Rule;
  notFound?: boolean;
  error?: RuleValidationError;
  conflictRule?: Rule;
}

export interface IPCheckResult {
  action: 'allow' | 'deny' | 'check';
  matchedRule?: Rule;
  ruleType?: RuleType;
}

export const ruleService = {
  createRule(request: CreateRuleRequest): CreateRuleResult {
    const cidrInfo = validateCIDR(request.ipPattern);
    if (!cidrInfo.isValid) {
      return { success: false, error: RuleValidationError.INVALID_CIDR };
    }

    const conflicting = ruleDAO.findConflictingRule(request.ipPattern);
    if (conflicting) {
      return { 
        success: false, 
        error: RuleValidationError.RULE_CONFLICT,
        conflictRule: conflicting
      };
    }

    const rule: Omit<Rule, 'createdAt'> = {
      id: uuidv4(),
      type: request.type,
      ipPattern: request.ipPattern,
      enabled: true,
      expiresAt: request.expiresAt
    };

    const created = ruleDAO.create(rule);
    return { success: true, rule: created };
  },

  updateRule(id: string, request: UpdateRuleRequest): UpdateRuleResult {
    const existing = ruleDAO.getById(id);
    if (!existing) {
      return { success: false, notFound: true };
    }

    if (request.ipPattern !== undefined) {
      const cidrInfo = validateCIDR(request.ipPattern);
      if (!cidrInfo.isValid) {
        return { success: false, error: RuleValidationError.INVALID_CIDR };
      }

      const conflicting = ruleDAO.findConflictingRule(request.ipPattern, id);
      if (conflicting) {
        return { 
          success: false, 
          error: RuleValidationError.RULE_CONFLICT,
          conflictRule: conflicting
        };
      }
    }

    const updated = ruleDAO.update(id, request);
    return { success: true, rule: updated! };
  },

  deleteRule(id: string): boolean {
    conditionDAO.deleteByRuleId(id);
    return ruleDAO.delete(id);
  },

  getRule(id: string): Rule | null {
    return ruleDAO.getById(id);
  },

  getAllRules(): Rule[] {
    return ruleDAO.getAll().filter(rule => !isExpired(rule.expiresAt));
  },

  getRulesByType(type: RuleType): Rule[] {
    return ruleDAO.getByType(type).filter(rule => !isExpired(rule.expiresAt));
  },

  checkIP(ip: string): IPCheckResult {
    const activeRules = this.getAllRules();
    const whitelistRules = activeRules.filter(r => r.type === 'whitelist');
    const blacklistRules = activeRules.filter(r => r.type === 'blacklist');

    for (const rule of whitelistRules) {
      if (ipMatchesPattern(ip, rule.ipPattern)) {
        return { action: 'allow', matchedRule: rule, ruleType: 'whitelist' };
      }
    }

    for (const rule of blacklistRules) {
      if (ipMatchesPattern(ip, rule.ipPattern)) {
        return { action: 'deny', matchedRule: rule, ruleType: 'blacklist' };
      }
    }

    return { action: 'check' };
  },

  addCondition(ruleId: string, key: string, value: string, operator: 'equals' | 'contains' | 'regex'): RuleCondition | null {
    const rule = ruleDAO.getById(ruleId);
    if (!rule) return null;

    const condition: RuleCondition = {
      id: uuidv4(),
      ruleId,
      key,
      value,
      operator
    };

    return conditionDAO.create(condition);
  },

  getCondition(id: string): RuleCondition | null {
    return conditionDAO.getById(id);
  },

  getConditions(ruleId: string): RuleCondition[] {
    return conditionDAO.getByRuleId(ruleId);
  },

  updateCondition(id: string, updates: Partial<Omit<RuleCondition, 'id' | 'ruleId'>>): RuleCondition | null {
    return conditionDAO.update(id, updates);
  },

  deleteCondition(id: string): boolean {
    return conditionDAO.delete(id);
  }
};
