import { Router, Request, Response } from 'express';
import {
  createChange,
  listChanges,
  updateChangeById,
  approveChange,
  rejectChange,
  executeChange,
  rollbackChange,
  confirmChange,
  getObserveStatus,
  validateCreateRequest,
  validateUpdateRequest
} from './service';

const router = Router();

const handleResult = (
  result: any,
  res: Response,
  successCode: number = 200
) => {
  if ('error' in result && 'code' in result) {
    return res.status(result.code).json({ error: result.error });
  }
  return res.status(successCode).json(result);
};

router.post('/changes', (req: Request, res: Response) => {
  const validation = validateCreateRequest(req.body);
  if (!validation.valid) {
    return res.status(400).json({ errors: validation.errors });
  }
  const change = createChange(req.body);
  res.status(201).json(change);
});

router.get('/changes', (_req: Request, res: Response) => {
  const changes = listChanges();
  res.json(changes);
});

router.put('/changes/:id', (req: Request, res: Response) => {
  const validation = validateUpdateRequest(req.body);
  if (!validation.valid) {
    return res.status(400).json({ errors: validation.errors });
  }
  const result = updateChangeById(req.params.id, req.body);
  handleResult(result, res);
});

router.post('/changes/:id/approve', (req: Request, res: Response) => {
  if (typeof req.body.approver !== 'string' || req.body.approver.trim().length === 0) {
    return res.status(400).json({ error: 'approver is required and must be a non-empty string' });
  }
  const result = approveChange(req.params.id, req.body);
  handleResult(result, res);
});

router.post('/changes/:id/reject', (req: Request, res: Response) => {
  if (typeof req.body.approver !== 'string' || req.body.approver.trim().length === 0) {
    return res.status(400).json({ error: 'approver is required and must be a non-empty string' });
  }
  const result = rejectChange(req.params.id, req.body);
  handleResult(result, res);
});

router.post('/changes/:id/execute', (req: Request, res: Response) => {
  const result = executeChange(req.params.id);
  if ('conflictEndsAt' in result) {
    return res.status(409).json({ 
      error: result.error,
      conflictEndsAt: result.conflictEndsAt 
    });
  }
  handleResult(result, res);
});

router.post('/changes/:id/rollback', (req: Request, res: Response) => {
  if (typeof req.body.reason !== 'string' || req.body.reason.trim().length === 0) {
    return res.status(400).json({ error: 'reason is required and must be a non-empty string' });
  }
  const result = rollbackChange(req.params.id, req.body);
  handleResult(result, res);
});

router.post('/changes/:id/confirm', (req: Request, res: Response) => {
  const result = confirmChange(req.params.id);
  handleResult(result, res);
});

router.get('/changes/:id/observe', (req: Request, res: Response) => {
  const result = getObserveStatus(req.params.id);
  handleResult(result, res);
});

export default router;
