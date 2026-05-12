import fs from 'fs';
import path from 'path';
import cron from 'node-cron';
import db from '../database';
import { BackupTaskService } from './BackupTaskService';
import { SnapshotStateService } from './SnapshotStateService';
import { Snapshot, SnapshotStatus } from '../types';
import { generateId, getCurrentTime, calculateSHA256 } from '../utils';

const MAX_SNAPSHOTS = 50;
const BACKUP_DIR = path.join(process.cwd(), 'backups');

if (!fs.existsSync(BACKUP_DIR)) {
  fs.mkdirSync(BACKUP_DIR, { recursive: true });
}

export class BackupService {
  private static instance: BackupService;
  private scheduledTasks: Map<string, cron.ScheduledTask> = new Map();
  
  static getInstance(): BackupService {
    if (!this.instance) {
      this.instance = new BackupService();
    }
    return this.instance;
  }

  private constructor() {
    this.initializeScheduledTasks();
  }

  private initializeScheduledTasks(): void {
    const tasks = BackupTaskService.getInstance().getAllTasks();
    tasks.forEach(task => {
      if (task.enabled) {
        this.scheduleTask(task.id, task.cronExpression);
      }
    });
  }

  private scheduleTask(taskId: string, cronExpression: string): void {
    if (this.scheduledTasks.has(taskId)) {
      this.scheduledTasks.get(taskId)?.stop();
    }

    const task = cron.schedule(cronExpression, async () => {
      const currentTask = BackupTaskService.getInstance().getTask(taskId);
      if (currentTask && currentTask.enabled) {
        await this.executeBackup(currentTask.dataRange);
      }
    });

    this.scheduledTasks.set(taskId, task);
  }

  private unscheduleTask(taskId: string): void {
    const task = this.scheduledTasks.get(taskId);
    if (task) {
      task.stop();
      this.scheduledTasks.delete(taskId);
    }
  }

  onTaskEnabled(taskId: string, cronExpression: string): void {
    this.scheduleTask(taskId, cronExpression);
  }

  onTaskDisabled(taskId: string): void {
    this.unscheduleTask(taskId);
  }

  onTaskCronUpdated(taskId: string, cronExpression: string): void {
    const task = BackupTaskService.getInstance().getTask(taskId);
    if (task && task.enabled) {
      this.scheduleTask(taskId, cronExpression);
    }
  }

  async triggerManualBackup(taskId: string): Promise<{ snapshot: Snapshot | null; error?: string }> {
    const task = BackupTaskService.getInstance().getTask(taskId);
    if (!task) {
      return { snapshot: null, error: 'Task not found' };
    }

    if (!task.enabled) {
      return { snapshot: null, error: 'Task is disabled, cannot trigger manual backup' };
    }

    return await this.executeBackup(task.dataRange);
  }

  private async executeBackup(dataRange: string): Promise<{ snapshot: Snapshot | null; error?: string }> {
    const snapshotId = generateId();
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
    const fileName = `snapshot-${snapshotId}-${timestamp}.json`;
    const filePath = path.join(BACKUP_DIR, fileName);

    const insertStmt = db.prepare(`
      INSERT INTO snapshots (id, created_at, data_range, file_size, status, sha256, file_path, error_reason)
      VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `);

    insertStmt.run(
      snapshotId,
      getCurrentTime(),
      dataRange,
      0,
      SnapshotStatus.CREATING,
      null,
      filePath,
      null
    );

    try {
      const backupData = await this.collectBackupData(dataRange);
      const jsonContent = JSON.stringify(backupData, null, 2);
      
      fs.writeFileSync(filePath, jsonContent, 'utf8');
      const fileSize = fs.statSync(filePath).size;

      const updateStmt = db.prepare('UPDATE snapshots SET file_size = ? WHERE id = ?');
      updateStmt.run(fileSize, snapshotId);

      const stateService = SnapshotStateService.getInstance();
      const transitionToValidating = stateService.transitionSnapshot(snapshotId, SnapshotStatus.VALIDATING);
      if (!transitionToValidating.success) {
        return { snapshot: null, error: transitionToValidating.error };
      }

      const validation = await this.validateSnapshot(snapshotId, filePath);
      if (!validation.success) {
        stateService.markValidationFailed(snapshotId, validation.error || 'Validation failed');
        return { snapshot: null, error: validation.error };
      }

      const sha256 = await calculateSHA256(filePath);
      stateService.updateSHA256(snapshotId, sha256);

      const transitionToAvailable = stateService.transitionSnapshot(snapshotId, SnapshotStatus.AVAILABLE);
      if (!transitionToAvailable.success) {
        return { snapshot: null, error: transitionToAvailable.error };
      }

      await this.cleanupOldSnapshots();

      const finalSnapshot = stateService.getSnapshot(snapshotId);
      return { snapshot: finalSnapshot };
    } catch (error) {
      const errorMsg = error instanceof Error ? error.message : 'Backup failed';
      const updateStmt = db.prepare('UPDATE snapshots SET error_reason = ? WHERE id = ?');
      updateStmt.run(errorMsg, snapshotId);
      
      return { snapshot: null, error: errorMsg };
    }
  }

  private async collectBackupData(dataRange: string): Promise<any> {
    return {
      dataRange,
      timestamp: getCurrentTime(),
      tables: {
        backup_tasks: db.prepare('SELECT * FROM backup_tasks').all(),
        snapshots: db.prepare('SELECT * FROM snapshots').all(),
        restore_operations: db.prepare('SELECT * FROM restore_operations').all()
      },
      metadata: {
        dataRange,
        backupTime: getCurrentTime()
      }
    };
  }

  private async validateSnapshot(snapshotId: string, filePath: string): Promise<{ success: boolean; error?: string }> {
    if (!fs.existsSync(filePath)) {
      return { success: false, error: 'Snapshot file does not exist' };
    }

    try {
      const content = fs.readFileSync(filePath, 'utf8');
      JSON.parse(content);
      return { success: true };
    } catch (error) {
      return { success: false, error: error instanceof Error ? error.message : 'Invalid JSON format' };
    }
  }

  private async cleanupOldSnapshots(): Promise<void> {
    const snapshots = db.prepare(`
      SELECT * FROM snapshots 
      ORDER BY created_at ASC
    `).all() as any[];

    if (snapshots.length <= MAX_SNAPSHOTS) {
      return;
    }

    const toDelete = snapshots.length - MAX_SNAPSHOTS;
    let deleted = 0;

    for (const snapshot of snapshots) {
      if (deleted >= toDelete) break;

      const hasActiveRestore = db.prepare(`
        SELECT COUNT(*) as count 
        FROM restore_operations 
        WHERE snapshot_id = ? AND status IN ('pending', 'in_progress')
      `).get(snapshot.id) as { count: number };

      if (hasActiveRestore.count > 0) {
        continue;
      }

      try {
        if (fs.existsSync(snapshot.file_path)) {
          fs.unlinkSync(snapshot.file_path);
        }
        
        const stateService = SnapshotStateService.getInstance();
        const currentSnapshot = stateService.getSnapshot(snapshot.id);
        if (currentSnapshot && currentSnapshot.status === SnapshotStatus.AVAILABLE) {
          stateService.transitionSnapshot(snapshot.id, SnapshotStatus.EXPIRED);
        }

        db.prepare('DELETE FROM snapshots WHERE id = ?').run(snapshot.id);
        deleted++;
      } catch (error) {
        console.error(`Failed to delete snapshot ${snapshot.id}:`, error);
      }
    }
  }

  getAllSnapshots(): Snapshot[] {
    const rows = db.prepare('SELECT * FROM snapshots ORDER BY created_at DESC').all() as any[];
    return rows.map(row => ({
      id: row.id,
      createdAt: row.created_at,
      dataRange: row.data_range,
      fileSize: row.file_size,
      status: row.status as SnapshotStatus,
      sha256: row.sha256,
      filePath: row.file_path,
      errorReason: row.error_reason
    }));
  }
}
