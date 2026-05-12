export type RuleType = 'whitelist' | 'blacklist';

export interface Rule {
  id: string;
  type: RuleType;
  ipPattern: string;
  createdAt: number;
  expiresAt?: number;
  enabled: boolean;
}

export interface RuleCondition {
  id: string;
  ruleId: string;
  key: string;
  value: string;
  operator: 'equals' | 'contains' | 'regex';
}

export interface RequestLog {
  id: string;
  ip: string;
  path: string;
  method: string;
  timestamp: number;
}

export interface RiskAssessment {
  score: number;
  factors: RiskFactor[];
  allowed: boolean;
}

export interface RiskFactor {
  type: string;
  description: string;
  score: number;
}

export interface RateLimitResult {
  allowed: boolean;
  remaining: number;
  reset: number;
  currentCount: number;
}

export interface ParsedIP {
  ip: string;
  isTrustedProxy: boolean;
}
