import { Router, Request, Response } from 'express';
import {
  createGrayscaleConfig,
  getGrayscaleConfig,
  startGrayscale,
  advanceGrayscaleStep,
  pauseGrayscale,
  resumeGrayscale
} from '../services/grayscaleService';

const router = Router();

router.post('/:experimentId/config', (req: Request, res: Response): void => {
  try {
    const { steps } = req.body;

    if (!steps || !Array.isArray(steps)) {
      res.status(400).json({ error: 'steps array is required' });
      return;
    }

    const config = createGrayscaleConfig(req.params.experimentId, steps);
    res.json(config);
  } catch (error: any) {
    res.status(400).json({ error: error.message });
  }
});

router.get('/:experimentId/config', (req: Request, res: Response): void => {
  try {
    const config = getGrayscaleConfig(req.params.experimentId);
    if (!config) {
      res.status(404).json({ error: 'Grayscale config not found' });
      return;
    }
    res.json(config);
  } catch (error: any) {
    res.status(500).json({ error: error.message });
  }
});

router.post('/:experimentId/start', (req: Request, res: Response): void => {
  try {
    const config = startGrayscale(req.params.experimentId);
    res.json(config);
  } catch (error: any) {
    res.status(400).json({ error: error.message });
  }
});

router.post('/:experimentId/advance', (req: Request, res: Response): void => {
  try {
    const result = advanceGrayscaleStep(req.params.experimentId);
    res.json(result);
  } catch (error: any) {
    res.status(400).json({ error: error.message });
  }
});

router.post('/:experimentId/pause', (req: Request, res: Response): void => {
  try {
    pauseGrayscale(req.params.experimentId);
    res.json({ success: true });
  } catch (error: any) {
    res.status(400).json({ error: error.message });
  }
});

router.post('/:experimentId/resume', (req: Request, res: Response): void => {
  try {
    const config = resumeGrayscale(req.params.experimentId);
    res.json(config);
  } catch (error: any) {
    res.status(400).json({ error: error.message });
  }
});

export default router;
