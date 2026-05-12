import {
  SyncTask,
  SyncResult,
  SyncMode,
  DataRecord,
  ConflictStrategy,
  ConflictItem,
  TaskStatus,
} from '../types';
import { getDataSource, DataSource } from './dataSource';
import { detectConflict, resolveConflict } from './conflict';
import {
  addToRetryQueue,
  getPendingRetryItems,
  removeFromRetryQueueByDataId,
  getManualInterventionItems,
} from '../models/retry';
import { updateTaskStatus, updateTaskWatermark } from '../models/task';
import { createReport } from '../models/report';

interface SyncContext {
  task: SyncTask;
  source: DataSource;
  target: DataSource;
  result: SyncResult;
}

function createEmptyResult(): SyncResult {
  return {
    insertCount: 0,
    updateCount: 0,
    deleteCount: 0,
    conflictCount: 0,
    manualInterventionItems: [],
    conflicts: [],
  };
}

async function processRecord(
  context: SyncContext,
  sourceRecord: DataRecord
): Promise<void> {
  const { task, source, target, result } = context;

  try {
    const targetRecord = await target.getRecordById(sourceRecord.id);

    const conflictCheck = detectConflict(
      sourceRecord,
      targetRecord,
      task.conflictStrategy
    );

    if (conflictCheck.isConflict && conflictCheck.item) {
      result.conflictCount++;
      result.conflicts.push(conflictCheck.item);

      if (task.conflictStrategy === ConflictStrategy.MANUAL_RESOLVE) {
        return;
      }

      const resolution = resolveConflict(
        sourceRecord,
        targetRecord!,
        task.conflictStrategy
      );

      if (!resolution.shouldWrite) {
        return;
      }
    }

    await target.upsertRecord(sourceRecord);
    removeFromRetryQueueByDataId(task.id, sourceRecord.id);

    if (targetRecord) {
      result.updateCount++;
    } else {
      result.insertCount++;
    }
  } catch (error) {
    const errMsg = error instanceof Error ? error.message : String(error);
    const retryItem = addToRetryQueue(task.id, sourceRecord, errMsg);

    if (retryItem.status === 'needs_manual') {
      if (!result.manualInterventionItems.includes(sourceRecord.id)) {
        result.manualInterventionItems.push(sourceRecord.id);
      }
    }
  }
}

async function processRetryQueue(context: SyncContext): Promise<void> {
  const { task } = context;
  const retryItems = getPendingRetryItems(task.id);

  for (const item of retryItems) {
    try {
      const record = JSON.parse(item.data) as DataRecord;
      await processRecord(context, record);
    } catch (error) {
      const errMsg = error instanceof Error ? error.message : String(error);
      addToRetryQueue(task.id, JSON.parse(item.data), errMsg);
    }
  }
}

async function performFullSync(context: SyncContext): Promise<SyncResult> {
  const { task, source } = context;

  await processRetryQueue(context);

  const sourceRecords = await source.getAllRecords();

  for (const record of sourceRecords) {
    await processRecord(context, record);
  }

  return context.result;
}

async function performIncrementalSync(
  context: SyncContext
): Promise<SyncResult> {
  const { task, source } = context;

  await processRetryQueue(context);

  const sourceRecords = await source.getRecordsAfterWatermark(task.watermark);

  let maxTimestamp = task.watermark;
  for (const record of sourceRecords) {
    if (record.updatedAt > maxTimestamp) {
      maxTimestamp = record.updatedAt;
    }
    await processRecord(context, record);
  }

  context.result = { ...context.result, _newWatermark: maxTimestamp } as any;
  return context.result;
}

export async function executeSync(task: SyncTask): Promise<{
  success: boolean;
  report?: any;
  error?: string;
  manualInterventionItems?: string[];
}> {
  const statusUpdate = updateTaskStatus(task.id, TaskStatus.SYNCING);
  if (!statusUpdate.success) {
    return {
      success: false,
      error: statusUpdate.reason,
    };
  }

  try {
    const source = getDataSource(task.sourceSystem);
    const target = getDataSource(task.targetSystem);

    const context: SyncContext = {
      task,
      source,
      target,
      result: createEmptyResult(),
    };

    let result: SyncResult;

    if (task.mode === SyncMode.FULL) {
      result = await performFullSync(context);
    } else {
      result = await performIncrementalSync(context);
    }

    const pendingManual = getManualInterventionItems(task.id);
    const pendingIds = pendingManual.map((item) => item.dataId);
    for (const id of pendingIds) {
      if (!result.manualInterventionItems.includes(id)) {
        result.manualInterventionItems.push(id);
      }
    }

    if (task.mode === SyncMode.INCREMENTAL) {
      const incrementalResult = result as any;
      if (incrementalResult._newWatermark !== undefined) {
        updateTaskWatermark(task.id, incrementalResult._newWatermark);
      }
    }

    const report = createReport(task.id, result);

    updateTaskStatus(task.id, TaskStatus.SUCCESS);

    if (result.manualInterventionItems.length > 0) {
      return {
        success: true,
        report,
        manualInterventionItems: result.manualInterventionItems,
        error: `Some data items require manual intervention: ${result.manualInterventionItems.join(', ')}`,
      };
    }

    return {
      success: true,
      report,
    };
  } catch (error) {
    updateTaskStatus(task.id, TaskStatus.FAILED);
    return {
      success: false,
      error: error instanceof Error ? error.message : String(error),
    };
  }
}
