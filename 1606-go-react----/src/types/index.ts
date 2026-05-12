export enum SensitiveWordLevel {
  LEVEL_1 = 1,
  LEVEL_2 = 2,
  LEVEL_3 = 3,
}

export interface SensitiveWord {
  id: string;
  word: string;
  level: SensitiveWordLevel;
  createdAt: Date;
  updatedAt: Date;
}

export enum ContentStatus {
  SUBMITTED = 'submitted',
  AUTO_REVIEWED = 'auto_reviewed',
  MANUAL_REVIEWED = 'manual_reviewed',
  FINALIZED = 'finalized',
}

export enum ReviewResult {
  PASS = 'pass',
  REJECT = 'reject',
  PENDING = 'pending',
}

export interface Condition {
  field: string;
  operator: string;
  value: string | string[] | number;
}

export interface Rule {
  id: string;
  name: string;
  priority: number;
  conditionLogic: 'AND' | 'OR';
  conditions: Condition[];
  action: 'REJECT' | 'PENDING' | 'PASS' | 'LOG';
  sensitiveWordIds: string[];
  createdAt: Date;
  updatedAt: Date;
}

export interface Content {
  id: string;
  text?: string;
  imageUrl?: string;
  status: ContentStatus;
  autoReviewResult?: ReviewResult;
  manualReviewResult?: ReviewResult;
  finalResult?: ReviewResult;
  submittedAt: Date;
  autoReviewedAt?: Date;
  manualReviewedAt?: Date;
  finalizedAt?: Date;
}

export interface ReviewLog {
  id: string;
  contentId: string;
  action: string;
  details: Record<string, unknown>;
  timestamp: Date;
}
