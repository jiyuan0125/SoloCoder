import { Router } from 'express';
import { createLog, listLogs, listLogsByService, getTraceLogs, getStats } from './controllers/logController';

const router = Router();

router.post('/logs', createLog);
router.get('/logs', listLogs);
router.get('/logs/stats', getStats);
router.get('/logs/:serviceName', listLogsByService);
router.get('/logs/trace/:traceId', getTraceLogs);

export default router;
