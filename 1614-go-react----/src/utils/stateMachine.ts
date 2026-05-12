import { PlanStatus } from '../types';

export enum ActionType {
  START_GRAYSCALE = 'start_grayscale',
  FULL_RELEASE = 'full_release',
  COMPLETE = 'complete',
  PAUSE = 'pause',
  RESUME = 'resume',
  ROLLBACK = 'rollback'
}

interface StateTransition {
  from: PlanStatus;
  to: PlanStatus;
  action: ActionType;
}

const validTransitions: StateTransition[] = [
  { from: PlanStatus.PENDING, to: PlanStatus.GRAYSCALE, action: ActionType.START_GRAYSCALE },
  { from: PlanStatus.GRAYSCALE, to: PlanStatus.FULL_RELEASE, action: ActionType.FULL_RELEASE },
  { from: PlanStatus.FULL_RELEASE, to: PlanStatus.COMPLETED, action: ActionType.COMPLETE },
  { from: PlanStatus.GRAYSCALE, to: PlanStatus.PAUSED, action: ActionType.PAUSE },
  { from: PlanStatus.PAUSED, to: PlanStatus.GRAYSCALE, action: ActionType.RESUME }
];

export function getAllowedActions(currentStatus: PlanStatus): ActionType[] {
  return validTransitions
    .filter(t => t.from === currentStatus)
    .map(t => t.action);
}

export function canTransition(from: PlanStatus, action: ActionType): boolean {
  return validTransitions.some(t => t.from === from && t.action === action);
}

export function getNextStatus(from: PlanStatus, action: ActionType): PlanStatus | null {
  const transition = validTransitions.find(t => t.from === from && t.action === action);
  return transition ? transition.to : null;
}

export function getActionDescription(action: ActionType): string {
  const descriptions: Record<ActionType, string> = {
    [ActionType.START_GRAYSCALE]: '开始灰度发布',
    [ActionType.FULL_RELEASE]: '全量发布',
    [ActionType.COMPLETE]: '完成发布',
    [ActionType.PAUSE]: '暂停灰度',
    [ActionType.RESUME]: '恢复灰度',
    [ActionType.ROLLBACK]: '回滚发布'
  };
  return descriptions[action];
}

export function isValidActionForStatus(status: PlanStatus, action: ActionType): boolean {
  return validTransitions.some(t => t.from === status && t.action === action);
}
