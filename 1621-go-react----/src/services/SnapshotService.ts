import fs from 'fs';
import db from '../database';
import { SnapshotStateService } from './SnapshotStateService';
import { RestoreService } from './RestoreService';
import { Snapshot } from '../types';

export class SnapshotService {
  private static instance: SnapshotService;
  
  static getInstance(): SnapshotService {
    if (!this.instance) {
      this.instance = new SnapshotService();
    }
    return this.instance;
  }

  private constructor() {}

  getSnapshot(id: string): Snapshot | null {
    return SnapshotStateService.getInstance().getSnapshot(id);
  }

  getAllSnapshots(): Snapshot[] {
    const rows = db.prepare('SELECT * FROM snapshots ORDER BY created_at DESC').all() as any[];
    return rows.map(row => ({
      id: row.id,
      createdAt: row.created_at,
      dataRange: row.data_range,
      fileSize: row.file_size,
      status: row.status,
      sha256: row.sha256,
      filePath: row.file_path,
      errorReason: row.error_reason
    }));
  }

  deleteSnapshot(id: string): { success: boolean; error?: string; conflict?: boolean } {
    const snapshot = this.getSnapshot(id);
    if (!snapshot) {
      return { success: false, error: 'Snapshot not found' };
    }

    const restoreService = RestoreService.getInstance();
    if (restoreService.isSnapshotInUse(id)) {
      return { 
        success: false, 
        error: '快照正在被恢复操作引用，无法删除',
        conflict: true
      };
    }

    try {
      if (fs.existsSync(snapshot.filePath)) {
        fs.unlinkSync(snapshot.filePath);
      }

      const stmt = db.prepare('DELETE FROM snapshots WHERE id = ?');
      stmt.run(id);
      
      return { success: true };
    } catch (error) {
      return { 
        success: false, 
        error: error instanceof Error ? error.message : 'Failed to delete snapshot' 
      };
    }
  }
}
