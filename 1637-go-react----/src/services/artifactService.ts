import { randomUUID, createHash } from 'crypto';
import { db, MAX_SNAPSHOT_VERSIONS } from '../database';
import { getProjectStorage } from './projectService';
import type { Artifact, ArtifactType } from '../types';
import * as fs from 'fs';
import * as path from 'path';

const STORAGE_ROOT = path.join(process.cwd(), 'data', 'storage');

function ensureStorageDir(): void {
  if (!fs.existsSync(STORAGE_ROOT)) {
    fs.mkdirSync(STORAGE_ROOT, { recursive: true });
  }
}

function isSnapshotVersion(version: string): boolean {
  return version.toUpperCase().includes('SNAPSHOT');
}

function calculateSha256(filePath: string): string {
  const hash = createHash('sha256');
  const fileBuffer = fs.readFileSync(filePath);
  hash.update(fileBuffer);
  return hash.digest('hex');
}

function findExistingArtifact(projectId: string, name: string, type: ArtifactType, version: string): Artifact | null {
  const stmt = db.prepare(`
    SELECT * FROM artifacts
    WHERE project_id = ? AND name = ? AND type = ? AND version = ?
  `);
  return stmt.get(projectId, name, type, version) as Artifact | null;
}

function findLatestArtifact(projectId: string, name: string, type: ArtifactType): Artifact | null {
  const stmt = db.prepare(`
    SELECT * FROM artifacts
    WHERE project_id = ? AND name = ? AND type = ? AND is_latest = 1
  `);
  return stmt.get(projectId, name, type) as Artifact | null;
}

function updateLatestTags(projectId: string, name: string, type: ArtifactType, newLatestId: string): void {
  db.prepare(`
    UPDATE artifacts
    SET is_latest = 0, updated_at = CURRENT_TIMESTAMP
    WHERE project_id = ? AND name = ? AND type = ? AND id != ?
  `).run(projectId, name, type, newLatestId);

  db.prepare(`
    UPDATE artifacts
    SET is_latest = 1, updated_at = CURRENT_TIMESTAMP
    WHERE id = ?
  `).run(newLatestId);
}

function cleanupOldSnapshots(projectId: string, name: string, type: ArtifactType): void {
  const snapshots = db.prepare(`
    SELECT id, storage_path FROM artifacts
    WHERE project_id = ? AND name = ? AND type = ? AND is_snapshot = 1
    ORDER BY created_at DESC
  `).all(projectId, name, type) as Artifact[];

  if (snapshots.length > MAX_SNAPSHOT_VERSIONS) {
    const toDelete = snapshots.slice(MAX_SNAPSHOT_VERSIONS);
    for (const artifact of toDelete) {
      deleteArtifactInternal(artifact.id, artifact.storage_path);
    }
  }
}

function deleteArtifactInternal(id: string, storagePath: string): void {
  db.prepare(`DELETE FROM artifacts WHERE id = ?`).run(id);
  if (fs.existsSync(storagePath)) {
    fs.unlinkSync(storagePath);
  }
}

export interface UploadResult {
  success: boolean;
  artifact?: Artifact;
  error?: {
    status: number;
    message: string;
  };
}

export function uploadArtifact(
  projectId: string,
  name: string,
  type: ArtifactType,
  version: string,
  tempFilePath: string,
  clientChecksum?: string
): UploadResult {
  ensureStorageDir();

  const projectStorage = getProjectStorage(projectId);
  if (!projectStorage) {
    fs.unlinkSync(tempFilePath);
    return {
      success: false,
      error: { status: 404, message: 'project not found' }
    };
  }

  const fileSize = fs.statSync(tempFilePath).size;

  if (projectStorage.used_bytes + fileSize > projectStorage.quota_bytes) {
    fs.unlinkSync(tempFilePath);
    return {
      success: false,
      error: {
        status: 403,
        message: `project storage quota exceeded, limit ${projectStorage.quota_bytes} bytes`
      }
    };
  }

  const actualChecksum = calculateSha256(tempFilePath);

  if (clientChecksum && actualChecksum !== clientChecksum) {
    fs.unlinkSync(tempFilePath);
    return {
      success: false,
      error: {
        status: 400,
        message: `checksum mismatch, expected ${clientChecksum} got ${actualChecksum}`
      }
    };
  }

  const isSnapshot = isSnapshotVersion(version);
  const existingArtifact = findExistingArtifact(projectId, name, type, version);

  if (existingArtifact) {
    if (!isSnapshot) {
      fs.unlinkSync(tempFilePath);
      return {
        success: false,
        error: {
          status: 403,
          message: 'release version cannot be overwritten'
        }
      };
    }

    const oldPath = existingArtifact.storage_path;
    if (fs.existsSync(oldPath)) {
      fs.unlinkSync(oldPath);
    }

    const newId = randomUUID();
    const storagePath = path.join(STORAGE_ROOT, newId);
    fs.renameSync(tempFilePath, storagePath);

    db.prepare(`
      UPDATE artifacts
      SET id = ?, storage_path = ?, sha256 = ?, size_bytes = ?, updated_at = CURRENT_TIMESTAMP
      WHERE project_id = ? AND name = ? AND type = ? AND version = ?
    `).run(newId, storagePath, actualChecksum, fileSize, projectId, name, type, version);

    const updated = getArtifactById(newId);
    cleanupOldSnapshots(projectId, name, type);

    return { success: true, artifact: updated as Artifact };
  }

  const id = randomUUID();
  const storagePath = path.join(STORAGE_ROOT, id);
  fs.renameSync(tempFilePath, storagePath);

  const currentLatest = findLatestArtifact(projectId, name, type);

  db.prepare(`
    INSERT INTO artifacts (
      id, project_id, name, type, version, storage_path,
      sha256, size_bytes, is_latest, is_snapshot
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0, ?)
  `).run(id, projectId, name, type, version, storagePath, actualChecksum, fileSize, isSnapshot ? 1 : 0);

  if (!currentLatest) {
    updateLatestTags(projectId, name, type, id);
  }

  cleanupOldSnapshots(projectId, name, type);

  const artifact = getArtifactById(id);
  return { success: true, artifact: artifact as Artifact };
}

export function getArtifactById(id: string): Artifact | null {
  const stmt = db.prepare(`SELECT * FROM artifacts WHERE id = ?`);
  return stmt.get(id) as Artifact | null;
}

export function listArtifacts(
  projectId: string,
  options?: { type?: ArtifactType; name?: string }
): Artifact[] {
  let sql = `SELECT * FROM artifacts WHERE project_id = ?`;
  const params: (string | ArtifactType)[] = [projectId];

  if (options?.type) {
    sql += ` AND type = ?`;
    params.push(options.type);
  }

  if (options?.name) {
    sql += ` AND name = ?`;
    params.push(options.name);
  }

  sql += ` ORDER BY created_at DESC`;

  return db.prepare(sql).all(...params) as Artifact[];
}

export interface DeleteResult {
  success: boolean;
  error?: { status: number; message: string };
}

export function deleteArtifact(id: string): DeleteResult {
  const artifact = getArtifactById(id);
  if (!artifact) {
    return {
      success: false,
      error: { status: 404, message: 'artifact not found' }
    };
  }

  if (artifact.is_latest) {
    return {
      success: false,
      error: { status: 400, message: 'cannot delete latest artifact, remove tag first' }
    };
  }

  deleteArtifactInternal(id, artifact.storage_path);
  return { success: true };
}

export interface TagLatestResult {
  success: boolean;
  artifact?: Artifact;
  error?: { status: number; message: string };
}

export function tagAsLatest(id: string): TagLatestResult {
  const artifact = getArtifactById(id);
  if (!artifact) {
    return {
      success: false,
      error: { status: 404, message: 'artifact not found' }
    };
  }

  updateLatestTags(artifact.project_id, artifact.name, artifact.type, id);

  const updated = getArtifactById(id);
  return { success: true, artifact: updated as Artifact };
}

export function incrementDownloadCount(id: string): void {
  db.prepare(`
    UPDATE artifacts
    SET download_count = download_count + 1
    WHERE id = ?
  `).run(id);
}
