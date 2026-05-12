import { randomUUID } from 'crypto';
import { db, DEFAULT_QUOTA_BYTES } from '../database';
import type { Project } from '../types';

export function createProject(name: string, description?: string, storageQuotaBytes?: number): Project {
  const id = randomUUID();
  const quota = storageQuotaBytes ?? DEFAULT_QUOTA_BYTES;

  const stmt = db.prepare(`
    INSERT INTO projects (id, name, description, storage_quota_bytes)
    VALUES (?, ?, ?, ?)
  `);

  const result = stmt.run(id, name, description ?? null, quota);

  return getProjectById(id) as Project;
}

export function getProjectById(id: string): Project | null {
  const stmt = db.prepare(`
    SELECT * FROM projects WHERE id = ?
  `);
  return stmt.get(id) as Project | null;
}

export function getProjectStorage(projectId: string): { used_bytes: number; quota_bytes: number } | null {
  const project = getProjectById(projectId);
  if (!project) return null;

  const stmt = db.prepare(`
    SELECT COALESCE(SUM(size_bytes), 0) as used_bytes
    FROM artifacts
    WHERE project_id = ?
  `);

  const result = stmt.get(projectId) as { used_bytes: number };

  return {
    used_bytes: result.used_bytes,
    quota_bytes: project.storage_quota_bytes
  };
}
