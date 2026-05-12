import { Router, Request, Response } from 'express';
import { v4 as uuidv4 } from 'uuid';
import { DataSubjectRequest, RequestType } from '../types';
import { store } from '../store';
import { addDays, anonymizeValue, isOverdue } from '../utils';

const router = Router();

function createRequest(type: RequestType, userId: string, details?: Record<string, unknown>): DataSubjectRequest {
  const now = new Date();
  return {
    id: uuidv4(),
    type,
    userId,
    status: 'pending',
    createdAt: now,
    deadline: addDays(now, 30),
    details
  };
}

router.post('/access', (req: Request, res: Response) => {
  const { userId } = req.body;
  if (!userId || typeof userId !== 'string') {
    return res.status(400).json({ error: 'userId is required and must be a string' });
  }
  const request = createRequest('access', userId);
  store.addRequest(request);
  return res.status(201).json(request);
});

router.post('/correction', (req: Request, res: Response) => {
  const { userId, corrections } = req.body;
  if (!userId || typeof userId !== 'string') {
    return res.status(400).json({ error: 'userId is required and must be a string' });
  }
  const request = createRequest('correction', userId, { corrections: corrections || null });
  store.addRequest(request);
  return res.status(201).json(request);
});

router.post('/deletion', (req: Request, res: Response) => {
  const { userId } = req.body;
  if (!userId || typeof userId !== 'string') {
    return res.status(400).json({ error: 'userId is required and must be a string' });
  }
  const request = createRequest('deletion', userId);
  store.addRequest(request);
  return res.status(201).json(request);
});

router.get('/:id', (req: Request, res: Response) => {
  const request = store.getRequest(req.params.id);
  if (!request) {
    return res.status(404).json({ error: 'Request not found' });
  }
  let status = request.status;
  if (status !== 'completed' && isOverdue(request.deadline)) {
    status = 'overdue';
    store.updateRequest(request.id, { status });
  }
  return res.json({ ...request, status });
});

router.post('/:id/process', (req: Request, res: Response) => {
  const request = store.getRequest(req.params.id);
  if (!request) {
    return res.status(404).json({ error: 'Request not found' });
  }
  if (request.status === 'completed' || request.status === 'overdue') {
    return res.status(400).json({ error: 'Request is already completed or overdue' });
  }

  store.updateRequest(request.id, { status: 'processing' });
  const userData = store.getUserDataByUserId(request.userId);
  const categories = store.getAllDataCategories();

  if (request.type === 'access') {
    const exportData: Record<string, unknown> = {};
    for (const record of userData) {
      const category = categories.find(c => c.id === record.categoryId);
      const catName = category?.name || record.categoryId;
      if (!exportData[catName]) {
        exportData[catName] = [];
      }
      (exportData[catName] as Array<Record<string, unknown>>).push({
        id: record.id,
        value: record.status === 'anonymized' ? (record.anonymizedValue || record.value) : record.value,
        status: record.status,
        createdAt: record.createdAt
      });
    }
    store.updateRequest(request.id, {
      status: 'completed',
      completedAt: new Date(),
      result: { exportedData: exportData }
    });
  } else if (request.type === 'deletion') {
    const processedRecords: Array<{ id: string; action: string }> = [];
    for (const record of userData) {
      const category = categories.find(c => c.id === record.categoryId);
      if (category?.legalBasis === 'legal_obligation') {
        const anonymized = anonymizeValue(record.value);
        store.updateUserData(record.id, {
          status: 'anonymized',
          anonymizedValue: anonymized
        });
        processedRecords.push({ id: record.id, action: 'anonymized' });
      } else {
        store.updateUserData(record.id, { status: 'pending_cleanup' });
        processedRecords.push({ id: record.id, action: 'marked_for_deletion' });
      }
    }
    store.updateRequest(request.id, {
      status: 'completed',
      completedAt: new Date(),
      result: { processedRecords }
    });
  } else if (request.type === 'correction') {
    store.updateRequest(request.id, {
      status: 'completed',
      completedAt: new Date(),
      result: { note: 'Correction request processed - requires manual review of corrections data' }
    });
  }

  const updated = store.getRequest(request.id);
  return res.json(updated);
});

router.get('/access/:id/export', (req: Request, res: Response) => {
  const request = store.getRequest(req.params.id);
  if (!request) {
    return res.status(404).json({ error: 'Request not found' });
  }
  if (request.type !== 'access') {
    return res.status(400).json({ error: 'Export is only available for access requests' });
  }
  if (request.status !== 'completed') {
    return res.status(400).json({ error: 'Request has not been processed yet' });
  }

  const userData = store.getUserDataByUserId(request.userId);
  const categories = store.getAllDataCategories();
  const exportData: Record<string, unknown> = {
    userId: request.userId,
    requestId: request.id,
    exportDate: new Date().toISOString(),
    data: {}
  };

  for (const record of userData) {
    const category = categories.find(c => c.id === record.categoryId);
    const catName = category?.name || record.categoryId;
    if (!((exportData.data as Record<string, unknown>)[catName])) {
      ((exportData.data as Record<string, unknown>)[catName]) = [];
    }
    (((exportData.data as Record<string, unknown>)[catName]) as Array<Record<string, unknown>>).push({
      value: record.status === 'anonymized' ? (record.anonymizedValue || record.value) : record.value,
      collectedAt: record.createdAt.toISOString(),
      status: record.status
    });
  }

  res.setHeader('Content-Type', 'application/json');
  res.setHeader('Content-Disposition', `attachment; filename="access-export-${request.id}.json"`);
  return res.json(exportData);
});

export default router;
