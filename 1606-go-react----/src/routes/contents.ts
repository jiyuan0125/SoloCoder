import { Router, Request, Response } from 'express';
import { contentService } from '../services/contentService';
import { ReviewResult } from '../types';

const router = Router();

router.post('/', (req: Request, res: Response) => {
  const { text, imageUrl } = req.body;
  const content = contentService.create(text, imageUrl);
  return res.status(201).json(content);
});

router.get('/', (_req: Request, res: Response) => {
  const contents = contentService.findAll();
  return res.json(contents);
});

router.get('/:id', (req: Request, res: Response) => {
  const content = contentService.findById(req.params.id);
  if (!content) {
    return res.status(404).json({ error: '内容不存在' });
  }
  return res.json(content);
});

router.post('/:id/auto-review', (req: Request, res: Response) => {
  const result = contentService.autoReview(req.params.id);
  
  if (!result.success) {
    const statusCode = result.statusCode || 400;
    return res.status(statusCode).json({ 
      error: result.error || '自动审核失败',
      currentStatus: result.content?.status
    });
  }
  
  return res.json(result.content);
});

router.post('/:id/manual-review', (req: Request, res: Response) => {
  const { result } = req.body;
  
  if (!['pass', 'reject', 'pending'].includes(result)) {
    return res.status(400).json({ error: '审核结果必须是 pass、reject 或 pending' });
  }
  
  const manualResult = contentService.manualReview(
    req.params.id,
    result as ReviewResult
  );
  
  if (!manualResult.success) {
    const statusCode = manualResult.statusCode || 400;
    return res.status(statusCode).json({ 
      error: manualResult.error || '人工复审失败',
      currentStatus: manualResult.content?.status
    });
  }
  
  return res.json(manualResult.content);
});

router.post('/:id/finalize', (req: Request, res: Response) => {
  const { result } = req.body;
  
  if (!['pass', 'reject'].includes(result)) {
    return res.status(400).json({ error: '最终判定必须是 pass 或 reject' });
  }
  
  const finalizeResult = contentService.finalize(
    req.params.id,
    result as ReviewResult
  );
  
  if (!finalizeResult.success) {
    const statusCode = finalizeResult.statusCode || 400;
    return res.status(statusCode).json({ 
      error: finalizeResult.error || '最终判定失败',
      currentStatus: finalizeResult.content?.status
    });
  }
  
  return res.json(finalizeResult.content);
});

export default router;
