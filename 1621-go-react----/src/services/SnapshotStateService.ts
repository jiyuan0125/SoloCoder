import db from '../database';
import { Snapshot, SnapshotStatus } from '../types';

const STATUS_TRANSITIONS: Map<SnapshotStatus, SnapshotStatus> = new Map([
  [SnapshotStatus.CREATING, SnapshotStatus.VALIDATING],
  [SnapshotStatus.VALIDATING, SnapshotStatus.AVAILABLE],
  [SnapshotStatus.AVAILABLE, SnapshotStatus.EXPIRED]
]);

const STATUS_NAMES: Map<SnapshotStatus, string> = new Map([
  [SnapshotStatus.CREATING, '创建中'],
  [SnapshotStatus.VALIDATING, '校验中'],
  [SnapshotStatus.AVAILABLE, '可用'],
  [SnapshotStatus.EXPIRED, '已过期']
]);

export class SnapshotStateService {
  private static instance: SnapshotStateService;
  
  static getInstance(): SnapshotStateService {
    if (!this.instance) {
      this.instance = new SnapshotStateService();
    }
    return this.instance;
  }

  private constructor() {}

  canTransition(from: SnapshotStatus, to: SnapshotStatus): boolean {
    return STATUS_TRANSITIONS.get(from) === to;
  }

  getNextStatus(current: SnapshotStatus): SnapshotStatus | null {
    return STATUS_TRANSITIONS.get(current) || null;
  }

  getStatusName(status: SnapshotStatus): string {
    return STATUS_NAMES.get(status) || status;
  }

  transitionSnapshot(snapshotId: string, to: SnapshotStatus): { success: boolean; error?: string } {
    const snapshot = this.getSnapshot(snapshotId);
    if (!snapshot) {
      return { success: false, error: 'Snapshot not found' };
    }

    if (!this.canTransition(snapshot.status, to)) {
      const currentName = this.getStatusName(snapshot.status);
      const toName = this.getStatusName(to);
      return { 
        success: false, 
        error: `状态流转不合法：无法从「${currentName}」直接跳转到「${toName}」，状态必须按顺序流转` 
      };
    }

    const stmt = db.prepare('UPDATE snapshots SET status = ? WHERE id = ?');
    stmt.run(to, snapshotId);
    
    return { success: true };
  }

  getSnapshot(id: string): Snapshot | null {
    const row = db.prepare('SELECT * FROM snapshots WHERE id = ?').get(id) as any;
    if (!row) return null;
    
    return {
      id: row.id,
      createdAt: row.created_at,
      dataRange: row.data_range,
      fileSize: row.file_size,
      status: row.status as SnapshotStatus,
      sha256: row.sha256,
      filePath: row.file_path,
      errorReason: row.error_reason
    };
  }

  markValidationFailed(snapshotId: string, errorReason: string): void {
    const stmt = db.prepare(`
      UPDATE snapshots 
      SET status = ?, error_reason = ? 
      WHERE id = ?
    `);
    stmt.run(SnapshotStatus.CREATING, errorReason, snapshotId);
  }

  updateSHA256(snapshotId: string, sha256: string): void {
    const stmt = db.prepare('UPDATE snapshots SET sha256 = ? WHERE id = ?');
    stmt.run(sha256, snapshotId);
  }
}
