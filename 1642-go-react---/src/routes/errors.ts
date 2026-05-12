import { Router, Request, Response } from 'express';
import { processErrorEvent, updateErrorGroupStatus, listErrorGroups, getErrorGroupTrend } from '../services/errorService';
import type { ErrorStatus, ErrorEventPayload, ErrorGroupFilters } from '../types';

const router = Router();

function getStatusFilterValue(status: string | undefined): ErrorStatus | undefined {
  const validStatuses: ErrorStatus[] = ['unhandled', 'acknowledged', 'resolved', 'ignored'];
  if (status && validStatuses.includes(status as ErrorStatus)) {
    return status as ErrorStatus;
  }
  return undefined;
}

router.post('/errors', (req: Request, res: Response) => {
  try {
    const payload = req.body as ErrorEventPayload;
    const result = processErrorEvent(payload);
    res.status(201).json(result);
  } catch (error: any) {
    const statusCode = error.statusCode || 500;
    res.status(statusCode).json({ error: error.message });
  }
});

router.get('/error-groups', (req: Request, res: Response) => {
  try {
    const filters: ErrorGroupFilters = {};
    const status = req.query.status as string | undefined;
    const isHighFrequency = req.query.isHighFrequency as string | undefined;

    if (status) {
      const parsed = getStatusFilterValue(status);
      if (!parsed) {
        return res.status(400).json({ error: `无效的 status 筛选值: ${status}` });
      }
      filters.status = parsed;
    }

    if (isHighFrequency !== undefined) {
      if (isHighFrequency === 'true' || isHighFrequency === '1') {
        filters.isHighFrequency = true;
      } else if (isHighFrequency === 'false' || isHighFrequency === '0') {
        filters.isHighFrequency = false;
      } else {
        return res.status(400).json({ error: `无效的 isHighFrequency 筛选值: ${isHighFrequency}` });
      }
    }

    const groups = listErrorGroups(filters);
    res.status(200).json(groups);
  } catch (error: any) {
    const statusCode = error.statusCode || 500;
    res.status(statusCode).json({ error: error.message });
  }
});

router.put('/error-groups/:id/status', (req: Request, res: Response) => {
  try {
    const { id } = req.params;
    const { status } = req.body as { status: ErrorStatus };

    if (!status) {
      return res.status(400).json({ error: '缺少必填字段: status' });
    }

    const result = updateErrorGroupStatus(id, status);
    res.status(200).json(result);
  } catch (error: any) {
    const statusCode = error.statusCode || 500;
    res.status(statusCode).json({ error: error.message });
  }
});

router.get('/error-groups/:id/trend', (req: Request, res: Response) => {
  try {
    const { id } = req.params;
    const trend = getErrorGroupTrend(id);
    res.status(200).json(trend);
  } catch (error: any) {
    const statusCode = error.statusCode || 500;
    res.status(statusCode).json({ error: error.message });
  }
});

export default router;
