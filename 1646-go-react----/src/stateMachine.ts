import { TaskStatus } from './types';

const VALID_TRANSITIONS: Record<TaskStatus, TaskStatus[]> = {
  CREATED: ['WAITING_DEPENDENCIES', 'PENDING'],
  WAITING_DEPENDENCIES: ['PENDING', 'SKIPPED'],
  PENDING: ['RUNNING'],
  RUNNING: ['COMPLETED', 'TIMEOUT'],
  COMPLETED: ['PENDING'],
  TIMEOUT: ['PENDING'],
  SKIPPED: ['CREATED']
};

export function canTransition(from: TaskStatus, to: TaskStatus): boolean {
  return VALID_TRANSITIONS[from]?.includes(to) ?? false;
}
