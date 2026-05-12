import express from 'express';
import { getWarnings, checkAndGenerateProjectWarnings } from '../services/warning-service';

const router = express.Router();

router.get('/', (req, res) => {
  const projectId = req.query.project_id as string | undefined;
  const warnings = getWarnings(projectId);
  return res.json(warnings);
});

router.post('/check', (_req, res) => {
  checkAndGenerateProjectWarnings();
  const warnings = getWarnings();
  return res.json({
    message: '已执行预警检查',
    warnings
  });
});

export default router;
