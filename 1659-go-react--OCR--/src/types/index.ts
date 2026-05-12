export interface SensitiveWordMatch {
  word: string;
  start: number;
  end: number;
}

export interface TextNormalizeResult {
  normalized: string;
  mappings: { normalizedIndex: number; originalIndex: number }[];
}

export type ContentType = 'text' | 'image';

export type AuditStatus = 'pending' | 'approved' | 'rejected' | 'warned' | 'queue';

export type AuditResult = 'approved' | 'rejected' | 'warned';

export interface Content {
  id: string;
  type: ContentType;
  content: string;
  status: AuditStatus;
  createdAt: number;
}

export interface AuditRecord {
  id: string;
  contentId: string;
  status: AuditStatus;
  matches?: SensitiveWordMatch[];
  createdAt: number;
  decidedAt?: number;
  decision?: AuditResult;
}
