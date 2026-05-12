export type ChangeType = 'release' | 'config' | 'scale' | 'migrate';

export type ChangeStatus = 'pending' | 'approved' | 'executing' | 'completed' | 'rolled_back';

export interface Change {
  id: string;
  title: string;
  type: ChangeType;
  affectedScope: string;
  planDescription: string;
  rollbackPlan: string;
  scheduledAt: string;
  status: ChangeStatus;
  createdAt: string;
  updatedAt: string;
  executedAt?: string;
  completedAt?: string;
  rolledBackAt?: string;
  rollbackReason?: string;
}

export interface Approval {
  id: string;
  changeId: string;
  approver: string;
  approved: boolean;
  comment?: string;
  createdAt: string;
}

export interface Fault {
  id: string;
  changeId: string;
  description: string;
  isFalsePositive: boolean;
  createdAt: string;
}

export interface CreateChangeRequest {
  title: string;
  type: ChangeType;
  affectedScope: string;
  planDescription: string;
  rollbackPlan: string;
  scheduledAt: string;
}

export interface UpdateChangeRequest {
  title?: string;
  type?: ChangeType;
  affectedScope?: string;
  planDescription?: string;
  rollbackPlan?: string;
  scheduledAt?: string;
}

export interface ApproveRequest {
  approver: string;
  comment?: string;
}

export interface RejectRequest {
  approver: string;
  comment?: string;
}

export interface RollbackRequest {
  reason: string;
}

export interface ObserveStatus {
  status: ChangeStatus;
  inObservation: boolean;
  observationEndsAt?: string;
  boundFaults: number;
  actualFaults: number;
}
