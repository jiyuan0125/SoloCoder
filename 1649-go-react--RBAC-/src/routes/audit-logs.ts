import express from 'express';
import * as auditService from '../services/auditService';

const router = express.Router();

router.get('/', (req, res) => {
  try {
    const { operationType, startTime, endTime } = req.query;
    const logs = auditService.getAuditLogs(
      operationType as string | undefined,
      startTime as string | undefined,
      endTime as string | undefined
    );
    
    res.json(logs);
  } catch (e: any) {
    res.status(500).json({ error: e.message });
  }
});

export default router;
