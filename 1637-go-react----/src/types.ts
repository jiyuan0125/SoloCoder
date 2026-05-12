export type ArtifactType = 'docker' | 'jar' | 'npm' | 'static';

export interface Project {
  id: string;
  name: string;
  description?: string;
  storage_quota_bytes: number;
  created_at: string;
}

export interface Artifact {
  id: string;
  project_id: string;
  name: string;
  type: ArtifactType;
  version: string;
  storage_path: string;
  sha256: string;
  size_bytes: number;
  download_count: number;
  is_latest: boolean;
  is_snapshot: boolean;
  created_at: string;
  updated_at: string;
}

export interface ProjectStorage {
  project_id: string;
  used_bytes: number;
  quota_bytes: number;
}
