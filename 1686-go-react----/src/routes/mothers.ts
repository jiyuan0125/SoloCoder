import { Router, Request, Response } from 'express';
import { createBooking, checkIn, checkOut, getMotherById, getAllMothers } from '../services/motherService';

const router = Router();

const handleError = (res: Response, error: Error): void => {
  switch (error.message) {
    case 'INVALID_ROOM_TYPE':
    case 'INVALID_DATE':
    case 'DUE_DATE_IN_PAST':
    case 'CHECK_IN_TOO_EARLY':
    case 'INVALID_STATUS':
    case 'NOT_CHECKED_IN':
    case 'CHECKOUT_BEFORE_CHECKIN':
      res.status(400).json({ error: error.message });
      break;
    case 'MOTHER_NOT_FOUND':
      res.status(404).json({ error: 'Mother not found' });
      break;
    default:
      res.status(500).json({ error: 'Internal server error', message: error.message });
  }
};

router.get('/', (_req: Request, res: Response): void => {
  try {
    const mothers = getAllMothers();
    res.json(mothers);
  } catch (error) {
    handleError(res, error as Error);
  }
});

router.get('/:id', (req: Request, res: Response): void => {
  try {
    const mother = getMotherById(req.params.id);
    if (!mother) {
      res.status(404).json({ error: 'Mother not found' });
      return;
    }
    res.json(mother);
  } catch (error) {
    handleError(res, error as Error);
  }
});

router.post('/book', (req: Request, res: Response): void => {
  try {
    const { name, expectedDueDate, roomType, emergencyContactName, emergencyContactPhone, expectedStayDays } = req.body;
    
    if (!name || !expectedDueDate || !roomType || !emergencyContactName || !emergencyContactPhone || !expectedStayDays) {
      res.status(400).json({ error: 'Missing required fields' });
      return;
    }

    const mother = createBooking(
      name,
      expectedDueDate,
      roomType,
      emergencyContactName,
      emergencyContactPhone,
      expectedStayDays
    );

    res.status(201).json(mother);
  } catch (error) {
    handleError(res, error as Error);
  }
});

router.post('/:id/check-in', (req: Request, res: Response): void => {
  try {
    const { checkInDate } = req.body;
    const mother = checkIn(req.params.id, checkInDate);
    res.json(mother);
  } catch (error) {
    handleError(res, error as Error);
  }
});

router.post('/:id/check-out', (req: Request, res: Response): void => {
  try {
    const { checkOutDate } = req.body;
    const result = checkOut(req.params.id, checkOutDate);
    res.json(result);
  } catch (error) {
    handleError(res, error as Error);
  }
});

export default router;
