export type ErrorStatus = 'unhandled' | 'acknowledged' | 'resolved' | 'ignored';

export interface ErrorEventPayload {
  serviceName: string;
  errorType: string;
  errorMessage: string;
  stackTrace?: string;
  occurredAt?: string;
  context?: Record<string, unknown>;
}

export interface ErrorEvent {
  id: string;
  groupId: string;
  serviceName: string;
  errorType: string;
  errorMessage: string;
  stackTrace: string | null;
  topFrame: string | null;
  occurredAt: number;
  context: string | null;
  userId: string | null;
}

export interface ErrorGroup {
  id: string;
  serviceName: string;
  errorType: string;
  errorMessage: string;
  topFrame: string | null;
  status: ErrorStatus;
  occurrenceCount: number;
  lastOccurredAt: number;
  firstOccurredAt: number;
  isHighFrequency: boolean;
  affectedUsers: number;
  createdUserIdSet: string | null;
}

export interface Deployment {
  id: string;
  version: string;
  deployedAt: number;
}

export interface TrendPoint {
  hour: string;
  count: number;
}

export interface ErrorGroupFilters {
  status?: ErrorStatus;
  isHighFrequency?: boolean;
}
