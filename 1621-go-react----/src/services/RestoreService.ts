import fs from 'fs';
import db from '../database';
import { SnapshotStateService } from './SnapshotStateService';
import { RestoreOperation, SnapshotStatus, Snapshot } from '../types';
import { generateId, getCurrentTime, calculateSHA256 } from '../utils';

export class RestoreService {
  private static instance: RestoreService;
  private restoreStates: Map<string, any> = new Map();
  
  static getInstance(): RestoreService {
    if (!this.instance) {
      this.instance = new RestoreService();
    }
    return this.instance;
  }

  private constructor() {}

  initiateRestore(snapshotId: string): { operation: RestoreOperation | null; error?: string } {
    const stateService = SnapshotStateService.getInstance();
    const snapshot = stateService.getSnapshot(snapshotId);
    
    if (!snapshot) {
      return { operation: null, error: 'Snapshot not found' };
    }

    if (snapshot.status !== SnapshotStatus.AVAILABLE) {
      const stateName = stateService.getStatusName(snapshot.status);
      return { operation: null, error: `Snapshot is not available for restore. Current status: ${stateName}` };
    }

    const validation = this.preValidateSnapshot(snapshot);
    if (!validation.success) {
      return { operation: null, error: validation.error };
    }

    const operationId = generateId();
    const restoreState = {
      snapshotId,
      originalData: this.captureCurrentState()
    };
    
    this.restoreStates.set(operationId, restoreState);

    const operation: RestoreOperation = {
      id: operationId,
      snapshotId,
      startTime: getCurrentTime(),
      endTime: null,
      status: 'pending',
      errorReason: null
    };

    const stmt = db.prepare(`
      INSERT INTO restore_operations (id, snapshot_id, start_time, end_time, status, error_reason)
      VALUES (?, ?, ?, ?, ?, ?)
    `);
    stmt.run(operation.id, operation.snapshotId, operation.startTime, null, 'pending', null);

    return { operation };
  }

  confirmRestore(operationId: string, confirmation: { 
    confirm: boolean; 
    acknowledgeDataLoss: boolean 
  }): { success: boolean; error?: string } {
    if (!confirmation.confirm || !confirmation.acknowledgeDataLoss) {
      return { 
        success: false, 
        error: 'Restore requires secondary confirmation: both confirm and acknowledgeDataLoss must be true' 
      };
    }

    const restoreState = this.restoreStates.get(operationId);
    if (!restoreState) {
      return { success: false, error: 'Restore operation not found or expired' };
    }

    const operation = this.getOperation(operationId);
    if (!operation) {
      return { success: false, error: 'Restore operation not found' };
    }

    if (operation.status !== 'pending') {
      return { success: false, error: `Invalid operation status: ${operation.status}` };
    }

    const updateStatus = db.prepare('UPDATE restore_operations SET status = ? WHERE id = ?');
    updateStatus.run('in_progress', operationId);

    try {
      const stateService = SnapshotStateService.getInstance();
      const snapshot = stateService.getSnapshot(restoreState.snapshotId);
      if (!snapshot) {
        throw new Error('Snapshot not found');
      }

      const integrityCheck = this.performIntegrityCheck(snapshot);
      if (!integrityCheck.success) {
        throw new Error(integrityCheck.error);
      }

      const restoreResult = this.executeRestore(snapshot);
      if (!restoreResult.success) {
        throw new Error(restoreResult.error);
      }

      const finalUpdate = db.prepare(`
        UPDATE restore_operations 
        SET status = ?, end_time = ? 
        WHERE id = ?
      `);
      finalUpdate.run('completed', getCurrentTime(), operationId);
      
      this.restoreStates.delete(operationId);
      return { success: true };
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : 'Restore failed';
      
      const rollbackResult = this.rollbackRestore(operationId, restoreState);
      
      const finalStatus = rollbackResult.success ? 'rolled_back' : 'failed';
      const finalError = rollbackResult.success 
        ? `Restore failed, but successfully rolled back. Error: ${errorMsg}`
        : `Restore failed and rollback also failed. Restore error: ${errorMsg}. Rollback error: ${rollbackResult.error}`;

      const finalUpdate = db.prepare(`
        UPDATE restore_operations 
        SET status = ?, end_time = ?, error_reason = ? 
        WHERE id = ?
      `);
      finalUpdate.run(finalStatus, getCurrentTime(), finalError, operationId);
      
      this.restoreStates.delete(operationId);
      return { success: false, error: finalError };
    }
  }

  private preValidateSnapshot(snapshot: Snapshot): { success: boolean; error?: string } {
    if (!fs.existsSync(snapshot.filePath)) {
      return { success: false, error: 'Snapshot file not found on disk' };
    }

    if (!snapshot.sha256) {
      return { success: false, error: 'Snapshot missing SHA256 checksum' };
    }

    return { success: true };
  }

  private performIntegrityCheck(snapshot: Snapshot): { success: boolean; error?: string } {
    try {
      const currentSHA256 = this.calculateFileSyncSHA256(snapshot.filePath);
      
      if (currentSHA256 !== snapshot.sha256) {
        return { 
          success: false, 
          error: `Integrity check failed: SHA256 checksum mismatch. Expected ${snapshot.sha256}, got ${currentSHA256}` 
        };
      }

      const content = fs.readFileSync(snapshot.filePath, 'utf8');
      const data = JSON.parse(content);
      
      if (!data || !data.metadata) {
        return { success: false, error: 'Integrity check failed: Invalid snapshot structure' };
      }

      return { success: true };
    } catch (error) {
      return { 
        success: false, 
        error: error instanceof Error 
          ? `Integrity check failed: ${error.message}` 
          : 'Integrity check failed with unknown error' 
      };
    }
  }

  private calculateFileSyncSHA256(filePath: string): string {
    const content = fs.readFileSync(filePath);
    return require('crypto').createHash('sha256').update(content).digest('hex');
  }

  private captureCurrentState(): any {
    return {
      backup_tasks: db.prepare('SELECT * FROM backup_tasks').all(),
      snapshots: db.prepare('SELECT * FROM snapshots').all(),
      restore_operations: db.prepare('SELECT * FROM restore_operations').all()
    };
  }

  private executeRestore(snapshot: Snapshot): { success: boolean; error?: string } {
    try {
      const content = fs.readFileSync(snapshot.filePath, 'utf8');
      const data = JSON.parse(content);
      
      if (!data.tables) {
        return { success: false, error: 'Invalid snapshot data structure' };
      }

      db.exec('BEGIN TRANSACTION');
      
      try {
        if (data.tables.backup_tasks) {
          db.exec('DELETE FROM backup_tasks');
          const insertTask = db.prepare(`
            INSERT INTO backup_tasks (id, name, cron_expression, enabled, data_range)
            VALUES (?, ?, ?, ?, ?)
          `);
          for (const task of data.tables.backup_tasks) {
            insertTask.run(task.id, task.name, task.cron_expression, task.enabled, task.data_range);
          }
        }

        db.exec('COMMIT');
        return { success: true };
      } catch (error) {
        db.exec('ROLLBACK');
        return { success: false, error: error instanceof Error ? error.message : 'Database transaction failed' };
      }
    } catch (error) {
      return { success: false, error: error instanceof Error ? error.message : 'Restore execution failed' };
    }
  }

  private rollbackRestore(operationId: string, restoreState: any): { success: boolean; error?: string } {
    try {
      db.exec('BEGIN TRANSACTION');
      
      try {
        db.exec('DELETE FROM backup_tasks');
        if (restoreState.originalData.backup_tasks) {
          const insertTask = db.prepare(`
            INSERT INTO backup_tasks (id, name, cron_expression, enabled, data_range)
            VALUES (?, ?, ?, ?, ?)
          `);
          for (const task of restoreState.originalData.backup_tasks) {
            insertTask.run(task.id, task.name, task.cron_expression, task.enabled, task.data_range);
          }
        }

        db.exec('COMMIT');
        return { success: true };
      } catch (error) {
        db.exec('ROLLBACK');
        return { success: false, error: error instanceof Error ? error.message : 'Rollback transaction failed' };
      }
    } catch (error) {
      return { success: false, error: error instanceof Error ? error.message : 'Rollback failed' };
    }
  }

  getOperation(id: string): RestoreOperation | null {
    const row = db.prepare('SELECT * FROM restore_operations WHERE id = ?').get(id) as any;
    if (!row) return null;
    
    return {
      id: row.id,
      snapshotId: row.snapshot_id,
      startTime: row.start_time,
      endTime: row.end_time,
      status: row.status,
      errorReason: row.error_reason
    };
  }

  getAllOperations(): RestoreOperation[] {
    const rows = db.prepare('SELECT * FROM restore_operations ORDER BY start_time DESC').all() as any[];
    return rows.map(row => ({
      id: row.id,
      snapshotId: row.snapshot_id,
      startTime: row.start_time,
      endTime: row.end_time,
      status: row.status,
      errorReason: row.error_reason
    }));
  }

  isSnapshotInUse(snapshotId: string): boolean {
    const result = db.prepare(`
      SELECT COUNT(*) as count 
      FROM restore_operations 
      WHERE snapshot_id = ? AND status IN ('pending', 'in_progress')
    `).get(snapshotId) as { count: number };
    
    return result.count > 0;
  }
}
