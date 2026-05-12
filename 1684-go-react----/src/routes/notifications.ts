import { Router, Request, Response } from 'express';
import { asyncHandler } from '../middleware/errorHandler';
import { notificationService } from '../services/notification';
import { careService } from '../services/care';

const router = Router();

router.post('/health-warnings/:warningId/notify-doctor', asyncHandler(async (req: Request, res: Response) => {
  const warningId = parseInt(req.params.warningId, 10);
  const { doctorName } = req.body as { doctorName: string };
  
  if (!doctorName || doctorName.trim() === '') {
    throw new Error('BAD_REQUEST: 医生姓名不能为空');
  }
  
  const result = notificationService.notifyDoctor(warningId, doctorName);
  res.json({ success: true, data: result });
}));

router.get('/health-warnings/:warningId/snapshot', asyncHandler(async (req: Request, res: Response) => {
  const warningId = parseInt(req.params.warningId, 10);
  const snapshot = notificationService.getWarningSnapshot(warningId);
  res.json({ success: true, data: snapshot });
}));

router.get('/notifications', asyncHandler(async (req: Request, res: Response) => {
  const { warningId } = req.query;
  const notifications = notificationService.getDoctorNotifications(
    warningId ? parseInt(warningId as string, 10) : undefined
  );
  res.json({ success: true, data: notifications });
}));

router.post('/health-warnings/:warningId/resolve', asyncHandler(async (req: Request, res: Response) => {
  const warningId = parseInt(req.params.warningId, 10);
  const warning = notificationService.resolveWarning(warningId);
  res.json({ success: true, data: warning });
}));

export const notificationRoutes = router;
