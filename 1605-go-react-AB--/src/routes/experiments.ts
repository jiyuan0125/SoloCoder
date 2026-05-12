import { Router, Request, Response } from 'express';
import {
  createExperiment,
  getExperimentById,
  getAllExperiments,
  updateExperimentStatus,
  updateVariants,
  deleteExperiment
} from '../services/experimentService';
import { ExperimentStatus, TrafficType } from '../types';

const router = Router();

router.post('/', (req: Request, res: Response): void => {
  try {
    const { name, trafficType, variants, metrics } = req.body;

    if (!name || !trafficType || !variants || !Array.isArray(variants)) {
      res.status(400).json({ error: 'Missing required fields' });
      return;
    }

    const validTrafficTypes = Object.values(TrafficType);
    if (!validTrafficTypes.includes(trafficType as TrafficType)) {
      res.status(400).json({ error: 'Invalid traffic type' });
      return;
    }

    const experiment = createExperiment(
      name,
      trafficType as TrafficType,
      variants,
      metrics || []
    );

    res.status(201).json(experiment);
  } catch (error: any) {
    res.status(400).json({ error: error.message });
  }
});

router.get('/', (_req: Request, res: Response): void => {
  try {
    const experiments = getAllExperiments();
    res.json(experiments);
  } catch (error: any) {
    res.status(500).json({ error: error.message });
  }
});

router.get('/:id', (req: Request, res: Response): void => {
  try {
    const experiment = getExperimentById(req.params.id);
    if (!experiment) {
      res.status(404).json({ error: 'Experiment not found' });
      return;
    }
    res.json(experiment);
  } catch (error: any) {
    res.status(500).json({ error: error.message });
  }
});

router.post('/:id/start', (req: Request, res: Response): void => {
  try {
    const experiment = updateExperimentStatus(req.params.id, ExperimentStatus.RUNNING);
    res.json(experiment);
  } catch (error: any) {
    res.status(400).json({ error: error.message });
  }
});

router.post('/:id/pause', (req: Request, res: Response): void => {
  try {
    const experiment = updateExperimentStatus(req.params.id, ExperimentStatus.PAUSED);
    res.json(experiment);
  } catch (error: any) {
    res.status(400).json({ error: error.message });
  }
});

router.post('/:id/resume', (req: Request, res: Response): void => {
  try {
    const experiment = updateExperimentStatus(req.params.id, ExperimentStatus.RUNNING);
    res.json(experiment);
  } catch (error: any) {
    res.status(400).json({ error: error.message });
  }
});

router.post('/:id/end', (req: Request, res: Response): void => {
  try {
    const experiment = updateExperimentStatus(req.params.id, ExperimentStatus.ENDED);
    res.json(experiment);
  } catch (error: any) {
    res.status(400).json({ error: error.message });
  }
});

router.put('/:id/variants', (req: Request, res: Response): void => {
  try {
    const { variants } = req.body;
    if (!variants || !Array.isArray(variants)) {
      res.status(400).json({ error: 'Variants array is required' });
      return;
    }

    const experiment = updateVariants(req.params.id, variants);
    res.json(experiment);
  } catch (error: any) {
    res.status(400).json({ error: error.message });
  }
});

router.delete('/:id', (req: Request, res: Response): void => {
  try {
    deleteExperiment(req.params.id);
    res.status(204).send();
  } catch (error: any) {
    res.status(400).json({ error: error.message });
  }
});

export default router;
