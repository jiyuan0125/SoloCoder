import { Router, Request, Response } from 'express';
import {
  createStaff,
  createShift,
  assignStaffToShift,
  swapShifts,
  getAllStaff,
  getAllShifts,
  getShiftAssignments,
  getStaffById,
  getShiftById,
  assignBatchToShift,
  checkShiftHasBothRoles
} from '../services/schedulingService';

const router = Router();

const handleError = (res: Response, error: Error): void => {
  switch (error.message) {
    case 'INVALID_STAFF_ROLE':
    case 'INVALID_SHIFT_TYPE':
    case 'WORK_DAYS_RULE_VIOLATION':
    case 'SHIFT_ROLES_INCOMPLETE':
      res.status(400).json({ error: error.message });
      break;
    case 'SHIFT_CONFLICT':
      res.status(409).json({ error: 'Shift conflict' });
      break;
    case 'STAFF_NOT_FOUND':
    case 'SHIFT_NOT_FOUND':
    case 'ASSIGNMENT_NOT_FOUND':
      res.status(404).json({ error: error.message });
      break;
    default:
      res.status(500).json({ error: 'Internal server error', message: error.message });
  }
};

router.get('/staff', (_req: Request, res: Response): void => {
  try {
    const staff = getAllStaff();
    res.json(staff);
  } catch (error) {
    handleError(res, error as Error);
  }
});

router.get('/staff/:staffId', (req: Request, res: Response): void => {
  try {
    const staff = getStaffById(req.params.staffId);
    if (!staff) {
      res.status(404).json({ error: 'Staff not found' });
      return;
    }
    res.json(staff);
  } catch (error) {
    handleError(res, error as Error);
  }
});

router.post('/staff', (req: Request, res: Response): void => {
  try {
    const { name, role } = req.body;
    if (!name || !role) {
      res.status(400).json({ error: 'Missing required fields' });
      return;
    }
    const staff = createStaff(name, role);
    res.status(201).json(staff);
  } catch (error) {
    handleError(res, error as Error);
  }
});

router.get('/shifts', (_req: Request, res: Response): void => {
  try {
    const shifts = getAllShifts();
    res.json(shifts);
  } catch (error) {
    handleError(res, error as Error);
  }
});

router.get('/shifts/:shiftId', (req: Request, res: Response): void => {
  try {
    const shift = getShiftById(req.params.shiftId);
    if (!shift) {
      res.status(404).json({ error: 'Shift not found' });
      return;
    }
    res.json(shift);
  } catch (error) {
    handleError(res, error as Error);
  }
});

router.post('/shifts', (req: Request, res: Response): void => {
  try {
    const { shiftType, shiftDate } = req.body;
    if (!shiftType || !shiftDate) {
      res.status(400).json({ error: 'Missing required fields' });
      return;
    }
    const shift = createShift(shiftType, shiftDate);
    res.status(201).json(shift);
  } catch (error) {
    handleError(res, error as Error);
  }
});

router.get('/shifts/:shiftId/assignments', (req: Request, res: Response): void => {
  try {
    const assignments = getShiftAssignments(req.params.shiftId);
    res.json(assignments);
  } catch (error) {
    handleError(res, error as Error);
  }
});

router.post('/assign', (req: Request, res: Response): void => {
  try {
    const { staffId, shiftId } = req.body;
    if (!staffId || !shiftId) {
      res.status(400).json({ error: 'Missing required fields' });
      return;
    }
    const assignment = assignStaffToShift(staffId, shiftId);
    res.status(201).json(assignment);
  } catch (error) {
    handleError(res, error as Error);
  }
});

router.post('/assign-batch', (req: Request, res: Response): void => {
  try {
    const { staffIds, shiftId } = req.body;
    if (!Array.isArray(staffIds) || staffIds.length === 0 || !shiftId) {
      res.status(400).json({ error: 'Missing required fields: staffIds (array) and shiftId' });
      return;
    }
    const assignments = assignBatchToShift(staffIds, shiftId);
    res.status(201).json(assignments);
  } catch (error) {
    handleError(res, error as Error);
  }
});

router.get('/shifts/:shiftId/validate', (req: Request, res: Response): void => {
  try {
    const shift = getShiftById(req.params.shiftId);
    if (!shift) {
      res.status(404).json({ error: 'SHIFT_NOT_FOUND' });
      return;
    }
    const isValid = checkShiftHasBothRoles(req.params.shiftId);
    res.json({ valid: isValid, shift });
  } catch (error) {
    handleError(res, error as Error);
  }
});

router.post('/staff/:staffId/shifts/:shiftId/swap', (req: Request, res: Response): void => {
  try {
    const { targetStaffId, targetShiftId } = req.body;
    
    if (!targetStaffId || !targetShiftId) {
      res.status(400).json({ error: 'Missing targetStaffId or targetShiftId' });
      return;
    }

    const staff1 = getStaffById(req.params.staffId);
    if (!staff1) {
      res.status(404).json({ error: 'STAFF_NOT_FOUND' });
      return;
    }

    const shift1 = getShiftById(req.params.shiftId);
    if (!shift1) {
      res.status(404).json({ error: 'SHIFT_NOT_FOUND' });
      return;
    }

    const staff2 = getStaffById(targetStaffId);
    if (!staff2) {
      res.status(404).json({ error: 'STAFF_NOT_FOUND' });
      return;
    }

    const shift2 = getShiftById(targetShiftId);
    if (!shift2) {
      res.status(404).json({ error: 'SHIFT_NOT_FOUND' });
      return;
    }

    const result = swapShifts(req.params.staffId, req.params.shiftId, targetStaffId, targetShiftId);
    res.json(result);
  } catch (error) {
    handleError(res, error as Error);
  }
});

export default router;
