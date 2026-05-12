import { Router, Request, Response } from 'express';
import multer from 'multer';
import * as projectService from '../services/projectService';
import * as artifactService from '../services/artifactService';
import type { ArtifactType } from '../types';

const router = Router();
const upload = multer({ dest: 'data/uploads/' });

router.post('/', (req: Request, res: Response) => {
  const { name, description, storageQuotaBytes } = req.body;

  if (!name || typeof name !== 'string') {
    return res.status(400).json({ error: 'name is required' });
  }

  if (storageQuotaBytes !== undefined && typeof storageQuotaBytes !== 'number') {
    return res.status(400).json({ error: 'storageQuotaBytes must be a number' });
  }

  try {
    const project = projectService.createProject(name, description, storageQuotaBytes);
    return res.status(201).json(project);
  } catch (error: any) {
    if (error?.code === 'SQLITE_CONSTRAINT_UNIQUE') {
      return res.status(409).json({ error: 'project name already exists' });
    }
    return res.status(500).json({ error: 'failed to create project' });
  }
});

router.get('/:id/storage', (req: Request, res: Response) => {
  const { id } = req.params;

  const storage = projectService.getProjectStorage(id);
  if (!storage) {
    return res.status(404).json({ error: 'project not found' });
  }

  return res.json({
    project_id: id,
    used_bytes: storage.used_bytes,
    quota_bytes: storage.quota_bytes
  });
});

router.post('/:id/artifacts', upload.single('file'), (req: Request, res: Response) => {
  const { id: projectId } = req.params;
  const { name, type, version, checksum } = req.body;

  if (!req.file) {
    return res.status(400).json({ error: 'file is required' });
  }

  if (!name || typeof name !== 'string') {
    return res.status(400).json({ error: 'name is required' });
  }

  if (!type || !['docker', 'jar', 'npm', 'static'].includes(type)) {
    return res.status(400).json({ error: 'type must be one of: docker, jar, npm, static' });
  }

  if (!version || typeof version !== 'string') {
    return res.status(400).json({ error: 'version is required' });
  }

  const result = artifactService.uploadArtifact(
    projectId,
    name,
    type as ArtifactType,
    version,
    req.file.path,
    checksum
  );

  if (!result.success) {
    return res.status(result.error!.status).json({ error: result.error!.message });
  }

  return res.status(201).json(result.artifact);
});

router.get('/:id/artifacts', (req: Request, res: Response) => {
  const { id: projectId } = req.params;
  const { type, name } = req.query;

  const project = projectService.getProjectById(projectId);
  if (!project) {
    return res.status(404).json({ error: 'project not found' });
  }

  const filters: { type?: ArtifactType; name?: string } = {};
  if (type && ['docker', 'jar', 'npm', 'static'].includes(type as string)) {
    filters.type = type as ArtifactType;
  }
  if (name && typeof name === 'string') {
    filters.name = name;
  }

  const artifacts = artifactService.listArtifacts(projectId, filters);
  return res.json(artifacts);
});

export default router;
