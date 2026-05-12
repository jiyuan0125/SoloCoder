import { Router } from 'express';
import {
  startProfiling,
  stopProfiling,
  getRecords,
  getFlamegraph,
  compareRecords
} from './controllers/profilingController';

const router = Router();

router.post('/profiling/:appId/start', startProfiling);
router.post('/profiling/:appId/stop', stopProfiling);
router.get('/profiling/:appId/records', getRecords);
router.get('/profiling/:appId/records/:recordId/flamegraph', getFlamegraph);
router.post('/profiling/compare', compareRecords);

export default router;
