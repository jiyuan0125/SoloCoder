import { Stage } from './types';

export const STAGE_ORDER: Stage[] = [
  '初次接触',
  '需求调研',
  '方案演示',
  '商务报价',
  '合同谈判',
  '赢单',
];

export const STAGE_INDEX: Record<Stage, number> = STAGE_ORDER.reduce(
  (acc, stage, index) => ({ ...acc, [stage]: index }),
  {} as Record<Stage, number>
);

export function getStageIndex(stage: Stage): number {
  return STAGE_INDEX[stage] ?? -1;
}

export function isStageAdvancement(from: Stage, to: Stage): boolean {
  if (to === '输单') return false;
  return getStageIndex(to) > getStageIndex(from);
}

export function isValidStageTransition(from: Stage, to: Stage): { valid: boolean; message?: string } {
  if (to === '输单') {
    return { valid: true };
  }

  const fromIndex = getStageIndex(from);
  const toIndex = getStageIndex(to);

  if (toIndex === -1) {
    return { valid: false, message: '无效的阶段' };
  }

  if (toIndex > fromIndex) {
    if (toIndex - fromIndex > 1) {
      if (from === '初次接触' && toIndex >= getStageIndex('方案演示')) {
        return { valid: false, message: '必须经过需求调研阶段' };
      }
      return { valid: false, message: '阶段只能向前推进一次' };
    }
  }

  return { valid: true };
}
