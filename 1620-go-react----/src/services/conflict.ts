import {
  DataRecord,
  ConflictStrategy,
  ConflictItem,
} from '../types';

export function detectConflict(
  source: DataRecord,
  target: DataRecord | undefined,
  strategy: ConflictStrategy
): { isConflict: boolean; item?: ConflictItem; resolvedBy?: ConflictStrategy } {
  if (!target) {
    return { isConflict: false };
  }

  if (target.updatedAt > source.updatedAt) {
    const conflictItem: ConflictItem = {
      dataId: source.id,
      sourceUpdatedAt: source.updatedAt,
      targetUpdatedAt: target.updatedAt,
      resolved: strategy !== ConflictStrategy.MANUAL_RESOLVE,
      resolvedBy:
        strategy !== ConflictStrategy.MANUAL_RESOLVE ? strategy : undefined,
    };

    return { isConflict: true, item: conflictItem, resolvedBy: conflictItem.resolvedBy };
  }

  return { isConflict: false };
}

export function resolveConflict(
  source: DataRecord,
  target: DataRecord,
  strategy: ConflictStrategy
): { shouldWrite: boolean; recordToUse: DataRecord } {
  switch (strategy) {
    case ConflictStrategy.SOURCE_PRIORITY:
      return { shouldWrite: true, recordToUse: source };

    case ConflictStrategy.TARGET_PRIORITY:
      return { shouldWrite: false, recordToUse: target };

    case ConflictStrategy.MANUAL_RESOLVE:
    default:
      return { shouldWrite: false, recordToUse: target };
  }
}
