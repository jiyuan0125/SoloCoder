import { Router, Request, Response } from 'express';
import { asyncHandler } from '../middleware/errorHandler';
import { careService } from '../services/care';
import { CarePlanUpdateRequest, CareRecordRequest } from '../types';

const router = Router();

router.get('/:elderId/care-plan', asyncHandler(async (req: Request, res: Response) => {
  const elderId = parseInt(req.params.elderId, 10);
  const plan = careService.getOrCreateCarePlan(elderId);
  res.json({ success: true, data: plan });
}));

router.put('/:elderId/care-plan', asyncHandler(async (req: Request, res: Response) => {
  const elderId = parseInt(req.params.elderId, 10);
  const request = req.body as CarePlanUpdateRequest;
  const result = careService.updateCarePlan(elderId, request);
  res.json({ success: true, data: result });
}));

router.get('/:elderId/care-plan/changes', asyncHandler(async (req: Request, res: Response) => {
  const elderId = parseInt(req.params.elderId, 10);
  const plan = careService.getOrCreateCarePlan(elderId);
  const changes = careService.getCarePlanChanges(plan.id);
  res.json({ success: true, data: changes });
}));

router.get('/:elderId/care-records', asyncHandler(async (req: Request, res: Response) => {
  const elderId = parseInt(req.params.elderId, 10);
  const { startDate, endDate } = req.query;
  const records = careService.getCareRecords(
    elderId, 
    startDate as string | undefined, 
    endDate as string | undefined
  );
  res.json({ success: true, data: records });
}));

router.post('/:elderId/care-records', asyncHandler(async (req: Request, res: Response) => {
  const elderId = parseInt(req.params.elderId, 10);
  const request = req.body as CareRecordRequest;
  const result = careService.createCareRecord(elderId, request);
  res.status(201).json({ success: true, data: result });
}));

router.get('/health-warnings', asyncHandler(async (req: Request, res: Response) => {
  const { elderId, status } = req.query;
  const warnings = careService.getHealthWarnings(
    elderId ? parseInt(elderId as string, 10) : undefined,
    status as string | undefined
  );
  res.json({ success: true, data: warnings });
}));

router.get('/health-warnings/:warningId', asyncHandler(async (req: Request, res: Response) => {
  const warningId = parseInt(req.params.warningId, 10);
  const warning = careService.getHealthWarning(warningId);
  
  if (!warning) {
    throw new Error('NOT_FOUND: 预警不存在');
  }
  
  res.json({ success: true, data: warning });
}));

router.get('/family-view-reminders', asyncHandler(async (_req: Request, res: Response) => {
  const reminders = careService.checkFamilyViewReminders();
  res.json({ success: true, data: reminders });
}));

export const careRoutes = router;
