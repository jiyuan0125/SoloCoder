export enum SnapshotStatus {
  CREATING = 'creating',
  VALIDATING = 'validating',
  AVAILABLE = 'available',
  EXPIRED = 'expired'
}

export interface Snapshot {
  id: string;
  createdAt: string;
  dataRange: string;
  fileSize: number;
  status: SnapshotStatus;
  sha256: string | null;
  filePath: string;
  errorReason: string | null;
}

export interface BackupTask {
  id: string;
  name: string;
  cronExpression: string;
  enabled: boolean;
  dataRange: string;
}

export interface RestoreOperation {
  id: string;
  snapshotId: string;
  startTime: string;
  endTime: string | null;
  status: 'pending' | 'in_progress' | 'completed' | 'failed' | 'rolled_back';
  errorReason: string | null;
}
