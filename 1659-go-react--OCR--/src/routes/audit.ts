import { Router, Request, Response } from 'express';
import { auditText, batchAuditText, decideTextAudit } from '../services/textAudit';
import { auditImage, batchAuditImages, decideImageAudit } from '../services/imageAudit';
import { getQueueItems } from '../db';

const router = Router();

router.post('/text/single', (req: Request, res: Response) => {
  const { contentId, text } = req.body as { contentId?: string; text?: string };
  if (!contentId || typeof contentId !== 'string') {
    res.status(400).json({ error: '缺少 contentId' });
    return;
  }
  if (text === undefined || typeof text !== 'string') {
    res.status(400).json({ error: '缺少文本内容' });
    return;
  }

  const result = auditText(contentId, text);
  if ('error' in result) {
    res.status(result.statusCode).json({ error: result.error });
    return;
  }
  res.json(result);
});

router.post('/text/batch', (req: Request, res: Response) => {
  const { items } = req.body as { items?: Array<{ contentId: string; text: string }> };
  if (!items || !Array.isArray(items)) {
    res.status(400).json({ error: '缺少 items 数组' });
    return;
  }
  const results = batchAuditText(items);
  res.json({ results });
});

router.post('/image/single', (req: Request, res: Response) => {
  const { contentId, imageUrl } = req.body as { contentId?: string; imageUrl?: string };
  if (!contentId || typeof contentId !== 'string') {
    res.status(400).json({ error: '缺少 contentId' });
    return;
  }
  if (!imageUrl || typeof imageUrl !== 'string') {
    res.status(400).json({ error: '缺少 imageUrl' });
    return;
  }

  const result = auditImage(contentId, imageUrl);
  res.json(result);
});

router.post('/image/batch', (req: Request, res: Response) => {
  const { items } = req.body as { items?: Array<{ contentId: string; imageUrl: string }> };
  if (!items || !Array.isArray(items)) {
    res.status(400).json({ error: '缺少 items 数组' });
    return;
  }
  const results = batchAuditImages(items);
  res.json({ results });
});

router.get('/queue', (req: Request, res: Response) => {
  const items = getQueueItems();
  res.json({ items });
});

router.post('/content/:contentId/audit/:auditId/action', (req: Request, res: Response) => {
  const { contentId, auditId } = req.params;
  const { action, type } = req.body as { action?: string; type?: string };

  if (!action || !['approved', 'rejected', 'warned'].includes(action)) {
    res.status(400).json({ error: '无效的 action' });
    return;
  }

  const decision = action as 'approved' | 'rejected' | 'warned';
  let result;

  if (type === 'image') {
    result = decideImageAudit(contentId, auditId, decision);
  } else {
    result = decideTextAudit(contentId, auditId, decision);
  }

  if (!result.success) {
    res.status(result.statusCode || 500).json({ error: result.error || '未知错误' });
    return;
  }

  res.json({ success: true });
});

export default router;
