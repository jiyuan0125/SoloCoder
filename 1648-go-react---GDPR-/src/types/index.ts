export type LegalBasis = 'consent' | 'contract' | 'legal_obligation' | 'legitimate_interest';

export type RequestType = 'access' | 'correction' | 'deletion';

export type RequestStatus = 'pending' | 'processing' | 'completed' | 'overdue';

export interface DataCategory {
  id: string;
  name: string;
  purpose: string;
  legalBasis: LegalBasis;
  retentionDays: number;
  createdAt: Date;
}

export interface UserDataRecord {
  id: string;
  userId: string;
  categoryId: string;
  value: string;
  createdAt: Date;
  status: 'active' | 'pending_cleanup' | 'anonymized';
  anonymizedValue?: string;
}

export interface DataSubjectRequest {
  id: string;
  type: RequestType;
  userId: string;
  status: RequestStatus;
  createdAt: Date;
  deadline: Date;
  details?: Record<string, unknown>;
  result?: Record<string, unknown>;
  completedAt?: Date;
}

export interface Consent {
  id: string;
  userId: string;
  categoryId: string;
  granted: boolean;
  grantedAt: Date;
  withdrawnAt?: Date;
}
