import { v4 as uuidv4 } from 'uuid';
import {
  Change,
  ChangeStatus,
  CreateChangeRequest,
  UpdateChangeRequest,
  ApproveRequest,
  RejectRequest,
  RollbackRequest,
  ObserveStatus
} from './types';
import {
  insertChange,
  updateChange,
  getChangeById,
  getAllChanges,
  insertApproval,
  getApprovalsByChangeId,
  getExecutingChange,
  getFaultsByChangeId
} from './database';

const VALID_CHANGE_TYPES = ['release', 'config', 'scale', 'migrate'] as const;
const VALID_STATUSES: ChangeStatus[] = ['pending', 'approved', 'executing', 'completed', 'rolled_back'];
const MIN_APPROVALS = 2;
const CONFLICT_WINDOW_HOURS = 2;
const OBSERVATION_PERIOD_HOURS = 24;

const nowISO = () => new Date().toISOString();

const formatISO = (date: Date) => date.toISOString();

const addHours = (iso: string, hours: number) => {
  const d = new Date(iso);
  d.setHours(d.getHours() + hours);
  return formatISO(d);
};

export const createChange = (data: CreateChangeRequest): Change => {
  const now = nowISO();
  const change: Change = {
    id: uuidv4(),
    title: data.title,
    type: data.type,
    affectedScope: data.affectedScope,
    planDescription: data.planDescription,
    rollbackPlan: data.rollbackPlan,
    scheduledAt: data.scheduledAt,
    status: 'pending',
    createdAt: now,
    updatedAt: now
  };
  insertChange(change);
  return change;
};

export const listChanges = (): Change[] => {
  return getAllChanges();
};

export const updateChangeById = (id: string, data: UpdateChangeRequest): Change | { error: string; code: number } => {
  const change = getChangeById(id);
  if (!change) return { error: 'change not found', code: 404 };

  if (change.status !== 'pending') {
    return { error: 'cannot update change in non-pending status', code: 400 };
  }

  const updates: Partial<Change> = { updatedAt: nowISO() };

  if (data.title !== undefined) updates.title = data.title;
  if (data.type !== undefined) updates.type = data.type;
  if (data.affectedScope !== undefined) updates.affectedScope = data.affectedScope;
  if (data.planDescription !== undefined) updates.planDescription = data.planDescription;
  if (data.rollbackPlan !== undefined) updates.rollbackPlan = data.rollbackPlan;
  if (data.scheduledAt !== undefined) updates.scheduledAt = data.scheduledAt;

  updateChange(id, updates);
  const updated = getChangeById(id)!;
  return updated;
};

export const approveChange = (id: string, data: ApproveRequest): Change | { error: string; code: number } => {
  const change = getChangeById(id);
  if (!change) return { error: 'change not found', code: 404 };

  if (change.status !== 'pending') {
    return { error: `invalid status transition from ${change.status} to approved`, code: 400 };
  }

  const now = nowISO();
  insertApproval({
    id: uuidv4(),
    changeId: id,
    approver: data.approver,
    approved: true,
    comment: data.comment,
    createdAt: now
  });

  const approvals = getApprovalsByChangeId(id);
  const positiveCount = approvals.filter(a => a.approved).length;

  if (positiveCount >= MIN_APPROVALS) {
    updateChange(id, { status: 'approved', updatedAt: now });
    const updated = getChangeById(id)!;
    return updated;
  }

  return change;
};

export const rejectChange = (id: string, data: RejectRequest): Change | { error: string; code: number } => {
  const change = getChangeById(id);
  if (!change) return { error: 'change not found', code: 404 };

  if (change.status !== 'pending') {
    return { error: `invalid status transition from ${change.status} to rejected`, code: 400 };
  }

  const now = nowISO();
  insertApproval({
    id: uuidv4(),
    changeId: id,
    approver: data.approver,
    approved: false,
    comment: data.comment,
    createdAt: now
  });

  updateChange(id, { status: 'rolled_back', updatedAt: now });
  const updated = getChangeById(id)!;
  return updated;
};

export const executeChange = (id: string): Change | { error: string; code: number; conflictEndsAt?: string } => {
  const change = getChangeById(id);
  if (!change) return { error: 'change not found', code: 404 };

  if (change.status !== 'approved') {
    return { error: `invalid status transition from ${change.status} to executing`, code: 400 };
  }

  const now = nowISO();
  if (new Date(now) < new Date(change.scheduledAt)) {
    return { error: 'cannot execute before scheduled time', code: 400 };
  }

  const executing = getExecutingChange(now);
  if (executing && executing.id !== id) {
    const conflictEndsAt = addHours(executing.executedAt!, CONFLICT_WINDOW_HOURS);
    return {
      error: `another change is currently executing, conflict window ends at ${conflictEndsAt}`,
      code: 409,
      conflictEndsAt
    };
  }

  updateChange(id, { status: 'executing', executedAt: now, updatedAt: now });
  const updated = getChangeById(id)!;
  return updated;
};

export const rollbackChange = (id: string, data: RollbackRequest): Change | { error: string; code: number } => {
  const change = getChangeById(id);
  if (!change) return { error: 'change not found', code: 404 };

  if (change.status !== 'executing') {
    return { error: `invalid status transition from ${change.status} to rolled_back`, code: 400 };
  }

  const now = nowISO();
  updateChange(id, {
    status: 'rolled_back',
    rolledBackAt: now,
    rollbackReason: data.reason,
    updatedAt: now
  });
  const updated = getChangeById(id)!;
  return updated;
};

export const confirmChange = (id: string): Change | { error: string; code: number } => {
  const change = getChangeById(id);
  if (!change) return { error: 'change not found', code: 404 };

  if (change.status !== 'executing') {
    return { error: `invalid status transition from ${change.status} to completed`, code: 400 };
  }

  const now = nowISO();
  updateChange(id, { status: 'completed', completedAt: now, updatedAt: now });
  const updated = getChangeById(id)!;
  return updated;
};

export const getObserveStatus = (id: string): ObserveStatus | { error: string; code: number } => {
  const change = getChangeById(id);
  if (!change) return { error: 'change not found', code: 404 };

  const faults = getFaultsByChangeId(id);
  const now = new Date();

  let inObservation = false;
  let observationEndsAt: string | undefined;

  if (change.status === 'completed' && change.completedAt) {
    observationEndsAt = addHours(change.completedAt, OBSERVATION_PERIOD_HOURS);
    inObservation = now < new Date(observationEndsAt);
  }

  return {
    status: change.status,
    inObservation,
    observationEndsAt: inObservation ? observationEndsAt : undefined,
    boundFaults: faults.length,
    actualFaults: faults.filter(f => !f.isFalsePositive).length
  };
};

export const validateCreateRequest = (data: any): { valid: boolean; errors: string[] } => {
  const errors: string[] = [];

  if (!data || typeof data !== 'object') {
    return { valid: false, errors: ['request body must be an object'] };
  }

  if (typeof data.title !== 'string' || data.title.trim().length === 0) {
    errors.push('title is required and must be a non-empty string');
  }

  if (!VALID_CHANGE_TYPES.includes(data.type)) {
    errors.push(`type must be one of: ${VALID_CHANGE_TYPES.join(', ')}`);
  }

  if (typeof data.affectedScope !== 'string' || data.affectedScope.trim().length === 0) {
    errors.push('affectedScope is required and must be a non-empty string');
  }

  if (typeof data.planDescription !== 'string' || data.planDescription.trim().length === 0) {
    errors.push('planDescription is required and must be a non-empty string');
  }

  if (typeof data.rollbackPlan !== 'string' || data.rollbackPlan.trim().length === 0) {
    errors.push('rollbackPlan is required and must be a non-empty string');
  }

  if (typeof data.scheduledAt !== 'string') {
    errors.push('scheduledAt is required and must be an ISO 8601 string');
  } else {
    const d = new Date(data.scheduledAt);
    if (isNaN(d.getTime())) {
      errors.push('scheduledAt must be a valid ISO 8601 string');
    }
  }

  return { valid: errors.length === 0, errors };
};

export const validateUpdateRequest = (data: any): { valid: boolean; errors: string[] } => {
  const errors: string[] = [];

  if (!data || typeof data !== 'object') {
    return { valid: false, errors: ['request body must be an object'] };
  }

  if (data.title !== undefined && (typeof data.title !== 'string' || data.title.trim().length === 0)) {
    errors.push('title must be a non-empty string');
  }

  if (data.type !== undefined && !VALID_CHANGE_TYPES.includes(data.type)) {
    errors.push(`type must be one of: ${VALID_CHANGE_TYPES.join(', ')}`);
  }

  if (data.affectedScope !== undefined && (typeof data.affectedScope !== 'string' || data.affectedScope.trim().length === 0)) {
    errors.push('affectedScope must be a non-empty string');
  }

  if (data.planDescription !== undefined && (typeof data.planDescription !== 'string' || data.planDescription.trim().length === 0)) {
    errors.push('planDescription must be a non-empty string');
  }

  if (data.rollbackPlan !== undefined && (typeof data.rollbackPlan !== 'string' || data.rollbackPlan.trim().length === 0)) {
    errors.push('rollbackPlan must be a non-empty string');
  }

  if (data.scheduledAt !== undefined) {
    if (typeof data.scheduledAt !== 'string') {
      errors.push('scheduledAt must be an ISO 8601 string');
    } else {
      const d = new Date(data.scheduledAt);
      if (isNaN(d.getTime())) {
        errors.push('scheduledAt must be a valid ISO 8601 string');
      }
    }
  }

  return { valid: errors.length === 0, errors };
};
