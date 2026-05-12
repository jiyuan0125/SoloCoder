import { Router, Request, Response } from 'express';
import * as artifactService from '../services/artifactService';
import * as fs from 'fs';
import path from 'path';

const router = Router();

router.get('/:id', (req: Request, res: Response) => {
  const { id } = req.params;

  const artifact = artifactService.getArtifactById(id);
  if (!artifact) {
    return res.status(404).json({ error: 'artifact not found' });
  }

  return res.json(artifact);
});

router.delete('/:id', (req: Request, res: Response) => {
  const { id } = req.params;

  const result = artifactService.deleteArtifact(id);
  if (!result.success) {
    return res.status(result.error!.status).json({ error: result.error!.message });
  }

  return res.status(204).send();
});

router.post('/:id/tag-latest', (req: Request, res: Response) => {
  const { id } = req.params;

  const result = artifactService.tagAsLatest(id);
  if (!result.success) {
    return res.status(result.error!.status).json({ error: result.error!.message });
  }

  return res.json(result.artifact);
});

router.get('/:id/download', (req: Request, res: Response) => {
  const { id } = req.params;

  const artifact = artifactService.getArtifactById(id);
  if (!artifact) {
    return res.status(404).json({ error: 'artifact not found' });
  }

  if (!fs.existsSync(artifact.storage_path)) {
    return res.status(404).json({ error: 'artifact file not found' });
  }

  artifactService.incrementDownloadCount(id);

  const filename = `${artifact.name}-${artifact.version}${path.extname(artifact.storage_path)}`;
  res.setHeader('Content-Disposition', `attachment; filename="${filename}"`);
  res.setHeader('Content-Type', 'application/octet-stream');

  const fileStream = fs.createReadStream(artifact.storage_path);
  fileStream.pipe(res);
});

export default router;
