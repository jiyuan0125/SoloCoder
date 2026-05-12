export interface Rule {
  id: string;
  name: string;
  condition: string;
  weight: number;
  priority: number;
  enabled: boolean;
  createdAt: number;
  updatedAt: number;
}

export interface RuleTreeNode {
  type: 'AND' | 'OR' | 'RULE';
  ruleId?: string;
  children?: RuleTreeNode[];
}

export interface Transaction {
  id: string;
  amount: number;
  deviceFingerprint: string;
  userId: string;
  merchantId: string;
  [key: string]: string | number | boolean | null;
}

export interface RiskAssessment {
  id: string;
  transactionId: string;
  score: number;
  status: 'APPROVED' | 'FLAGGED' | 'BLOCKED' | 'PENDING_REVIEW';
  matchedRuleIds: string[];
  assessedAt: number;
  accountFrozen: boolean;
  frozenAttempted: boolean;
  freezeFailed: boolean;
}

export interface ReviewRecord {
  id: string;
  assessmentId: string;
  transactionId: string;
  reviewerId?: string;
  decision: 'NORMAL' | 'FRAUD' | 'PENDING';
  decidedAt?: number;
  createdAt: number;
}

export interface Account {
  id: string;
  userId: string;
  frozen: boolean;
  frozenAt?: number;
}

export interface ConditionParseError extends Error {
  code: string;
}
