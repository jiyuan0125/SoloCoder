import { Router, Request, Response } from 'express';
import { asyncHandler } from '../middleware/errorHandler';
import { familyService } from '../services/family';
import { careService } from '../services/care';
import { VisitAppointmentRequest } from '../types';

const router = Router();

router.get('/elders/:elderId/family-members', asyncHandler(async (req: Request, res: Response) => {
  const elderId = parseInt(req.params.elderId, 10);
  const members = familyService.getFamilyMembers(elderId);
  res.json({ success: true, data: members });
}));

router.post('/elders/:elderId/family-members', asyncHandler(async (req: Request, res: Response) => {
  const elderId = parseInt(req.params.elderId, 10);
  const { name, phone, relation } = req.body as { name: string; phone: string; relation: string };
  
  if (!name || !phone || !relation) {
    throw new Error('BAD_REQUEST: 姓名、电话和关系不能为空');
  }
  
  const member = familyService.addFamilyMember(elderId, name, phone, relation);
  res.status(201).json({ success: true, data: member });
}));

router.post('/family-members/:memberId/view', asyncHandler(async (req: Request, res: Response) => {
  const memberId = parseInt(req.params.memberId, 10);
  const member = familyService.updateLastView(memberId);
  
  if (!member) {
    throw new Error('NOT_FOUND: 家属不存在');
  }
  
  res.json({ success: true, data: member });
}));

router.get('/elders/:elderId/care-records', asyncHandler(async (req: Request, res: Response) => {
  const elderId = parseInt(req.params.elderId, 10);
  const { startDate, endDate } = req.query;
  const records = careService.getCareRecords(
    elderId,
    startDate as string | undefined,
    endDate as string | undefined
  );
  res.json({ success: true, data: records });
}));

router.post('/family-members/:memberId/messages', asyncHandler(async (req: Request, res: Response) => {
  const memberId = parseInt(req.params.memberId, 10);
  const { elderId, content } = req.body as { elderId: number; content: string };
  
  if (!elderId || !content || content.trim() === '') {
    throw new Error('BAD_REQUEST: 老人ID和留言内容不能为空');
  }
  
  const message = familyService.sendMessage(memberId, elderId, content);
  res.status(201).json({ success: true, data: message });
}));

router.get('/elders/:elderId/messages', asyncHandler(async (req: Request, res: Response) => {
  const elderId = parseInt(req.params.elderId, 10);
  const { familyMemberId } = req.query;
  const messages = familyService.getMessages(
    elderId,
    familyMemberId ? parseInt(familyMemberId as string, 10) : undefined
  );
  res.json({ success: true, data: messages });
}));

router.post('/appointments', asyncHandler(async (req: Request, res: Response) => {
  const request = req.body as VisitAppointmentRequest;
  const appointment = familyService.createAppointment(request);
  res.status(201).json({ success: true, data: appointment });
}));

router.get('/appointments', asyncHandler(async (req: Request, res: Response) => {
  const { elderId, familyMemberId, date } = req.query;
  const appointments = familyService.getAppointments(
    elderId ? parseInt(elderId as string, 10) : undefined,
    familyMemberId ? parseInt(familyMemberId as string, 10) : undefined,
    date as string | undefined
  );
  res.json({ success: true, data: appointments });
}));

router.post('/appointments/:id/cancel', asyncHandler(async (req: Request, res: Response) => {
  const appointmentId = parseInt(req.params.id, 10);
  const appointment = familyService.cancelAppointment(appointmentId);
  
  if (!appointment) {
    throw new Error('NOT_FOUND: 预约不存在');
  }
  
  res.json({ success: true, data: appointment });
}));

export const familyRoutes = router;
