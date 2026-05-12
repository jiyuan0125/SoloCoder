import { Router, Request, Response } from 'express';
import { createBaby, addBabyLog, getBabyById, getBabyLogs, getBabiesByMother } from '../services/babyService';

const router = Router();

const handleError = (res: Response, error: Error): void => {
  switch (error.message) {
    case 'INVALID_FEEDING_TYPE':
    case 'NEGATIVE_WEIGHT':
    case 'NEGATIVE_TEMPERATURE':
    case 'INVALID_LENGTH':
      res.status(400).json({ error: error.message });
      break;
    case 'MOTHER_NOT_FOUND':
    case 'BABY_NOT_FOUND':
      res.status(404).json({ error: error.message });
      break;
    default:
      res.status(500).json({ error: 'Internal server error', message: error.message });
  }
};

router.get('/mother/:motherId', (req: Request, res: Response): void => {
  try {
    const babies = getBabiesByMother(req.params.motherId);
    res.json(babies);
  } catch (error) {
    handleError(res, error as Error);
  }
});

router.get('/:id', (req: Request, res: Response): void => {
  try {
    const baby = getBabyById(req.params.id);
    if (!baby) {
      res.status(404).json({ error: 'Baby not found' });
      return;
    }
    res.json(baby);
  } catch (error) {
    handleError(res, error as Error);
  }
});

router.post('/', (req: Request, res: Response): void => {
  try {
    const { motherId, name, birthWeight, birthLength, feedingType, birthDate } = req.body;
    
    if (!motherId || !name || !birthWeight || !birthLength || !feedingType || !birthDate) {
      res.status(400).json({ error: 'Missing required fields' });
      return;
    }

    const baby = createBaby(
      motherId,
      name,
      birthWeight,
      birthLength,
      feedingType,
      birthDate
    );

    res.status(201).json(baby);
  } catch (error) {
    handleError(res, error as Error);
  }
});

router.get('/:id/logs', (req: Request, res: Response): void => {
  try {
    const logs = getBabyLogs(req.params.id);
    res.json(logs);
  } catch (error) {
    handleError(res, error as Error);
  }
});

router.post('/:id/logs', (req: Request, res: Response): void => {
  try {
    const { logDate, feedingTime, feedingAmount, diaperChange, temperature, sleepDuration, jaundiceIndex, currentWeight } = req.body;
    
    if (!logDate) {
      res.status(400).json({ error: 'Missing logDate' });
      return;
    }

    const log = addBabyLog(
      req.params.id,
      logDate,
      feedingTime,
      feedingAmount,
      diaperChange,
      temperature,
      sleepDuration,
      jaundiceIndex,
      currentWeight
    );

    res.status(201).json(log);
  } catch (error) {
    handleError(res, error as Error);
  }
});

export default router;
