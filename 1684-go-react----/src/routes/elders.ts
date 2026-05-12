import { Router, Request, Response } from 'express';
import { asyncHandler } from '../middleware/errorHandler';
import { elderService } from '../services/elder';
import { CheckInRequest, CheckOutRequest } from '../types';

const router = Router();

router.get('/', asyncHandler(async (_req: Request, res: Response) => {
  const { status } = _req.query;
  const elders = elderService.getElders(status as string);
  res.json({ success: true, data: elders });
}));

router.get('/occupancy', asyncHandler(async (_req: Request, res: Response) => {
  const stats = elderService.getOccupancyStats();
  res.json({ success: true, data: stats });
}));

router.get('/:id', asyncHandler(async (req: Request, res: Response) => {
  const elderId = parseInt(req.params.id, 10);
  const elder = elderService.getElder(elderId);
  
  if (!elder) {
    throw new Error('NOT_FOUND: 老人不存在');
  }
  
  res.json({ success: true, data: elder });
}));

router.post('/', asyncHandler(async (req: Request, res: Response) => {
  const request = req.body as CheckInRequest;
  const result = elderService.checkIn(request);
  res.status(201).json({ success: true, data: result });
}));

router.post('/:id/checkout', asyncHandler(async (req: Request, res: Response) => {
  const elderId = parseInt(req.params.id, 10);
  const request = req.body as CheckOutRequest;
  const result = elderService.checkOut(elderId, request);
  res.json({ success: true, data: result });
}));

router.get('/:id/billing', asyncHandler(async (req: Request, res: Response) => {
  const elderId = parseInt(req.params.id, 10);
  const records = elderService.getBillingRecords(elderId);
  const deposit = elderService.getCurrentDeposit(elderId);
  const unpaid = elderService.calculateUnpaidAmount(elderId);
  
  res.json({ 
    success: true, 
    data: { records, deposit, unpaid } 
  });
}));

export const elderRoutes = router;
