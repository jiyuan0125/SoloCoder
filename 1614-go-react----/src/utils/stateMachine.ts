const STATUS_PENDING = 'pending';
const STATUS_GRAYSCALE = 'grayscale';
const STATUS_FULL_RELEASE = 'full_release';
const STATUS_PAUSED = 'paused';
const STATUS_COMPLETED = 'completed';
const STATUS_ROLLED_BACK = 'rolled_back';

export enum ActionType {
  START_GRAYSCALE = 'start_grayscale',
  FULL_RELEASE = 'full_release',
  COMPLETE = 'complete',
  PAUSE = 'pause',
  RESUME = 'resume',
  ROLLBACK = 'rollback'
}

interface StateTransition {
  from: string;
  to: string;
  action: ActionType;
}

const validTransitions: StateTransition[] = [
  { from: STATUS_PENDING, to: STATUS_GRAYSCALE, action: ActionType.START_GRAYSCALE },
  { from: STATUS_GRAYSCALE, to: STATUS_FULL_RELEASE, action: ActionType.FULL_RELEASE },
  { from: STATUS_FULL_RELEASE, to: STATUS_COMPLETED, action: ActionType.COMPLETE },
  { from: STATUS_GRAYSCALE, to: STATUS_PAUSED, action: ActionType.PAUSE },
  { from: STATUS_PAUSED, to: STATUS_GRAYSCALE, action: ActionType.RESUME }
];

export function getAllowedActions(currentStatus: string): ActionType[] {
  return validTransitions
    .filter(t => t.from === currentStatus)
    .map(t => t.action);
}

export function canTransition(from: string, action: ActionType): boolean {
  return validTransitions.some(t => t.from === from && t.action === action);
}

export function getNextStatus(from: string, action: ActionType): string | null {
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

export function isValidActionForStatus(status: string, action: ActionType): boolean {
  return validTransitions.some(t => t.from === status && t.action === action);
}
