import { Router, Request, Response } from 'express';
import { getAlertById, updateAlertStatus } from '../services/alertService';

const router = Router();

router.get('/:id', async (req: Request, res: Response) => {
  try {
    const alert = await getAlertById(req.params.id);
    if (!alert) {
      res.status(404).json({ error: 'Alert not found' });
      return;
    }
    res.json(alert);
  } catch (err: any) {
    res.status(500).json({ error: err.message });
  }
});

router.post('/:id/process', async (req: Request, res: Response) => {
  try {
    const { processedBy } = req.body;
    const alert = await updateAlertStatus(req.params.id, 'processing', processedBy);
    
    if (!alert) {
      res.status(404).json({ error: 'Alert not found' });
      return;
    }
    
    res.json(alert);
  } catch (err: any) {
    if (err.message === 'Invalid status transition') {
      res.status(400).json({ error: err.message });
    } else {
      res.status(500).json({ error: err.message });
    }
  }
});

router.post('/:id/resolve', async (req: Request, res: Response) => {
  try {
    const { processedBy, processingRecord } = req.body;

    if (!processingRecord) {
      res.status(400).json({ error: 'Processing record is required' });
      return;
    }

    const alert = await updateAlertStatus(
      req.params.id, 
      'resolved', 
      processedBy, 
      processingRecord
    );
    
    if (!alert) {
      res.status(404).json({ error: 'Alert not found' });
      return;
    }
    
    res.json(alert);
  } catch (err: any) {
    if (err.message === 'Invalid status transition') {
      res.status(400).json({ error: err.message });
    } else if (err.message === 'Processing record is required for resolution') {
      res.status(400).json({ error: err.message });
    } else {
      res.status(500).json({ error: err.message });
    }
  }
});

export default router;
