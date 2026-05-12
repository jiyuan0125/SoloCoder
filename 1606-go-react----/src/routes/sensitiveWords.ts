import { Router, Request, Response } from 'express';
import { sensitiveWordService } from '../services/sensitiveWordService';
import { SensitiveWordLevel } from '../types';

const router = Router();

router.post('/', (req: Request, res: Response) => {
  const { word, level } = req.body;
  
  if (!word || typeof word !== 'string') {
    return res.status(400).json({ error: '词不能为空' });
  }
  
  if (![1, 2, 3].includes(level)) {
    return res.status(400).json({ error: '等级必须为 1、2 或 3' });
  }
  
  const sensitiveWord = sensitiveWordService.create(word, level as SensitiveWordLevel);
  return res.status(201).json(sensitiveWord);
});

router.get('/', (_req: Request, res: Response) => {
  const words = sensitiveWordService.findAll();
  return res.json(words);
});

router.get('/:id', (req: Request, res: Response) => {
  const word = sensitiveWordService.findById(req.params.id);
  if (!word) {
    return res.status(404).json({ error: '敏感词不存在' });
  }
  return res.json(word);
});

router.put('/:id', (req: Request, res: Response) => {
  const { word, level } = req.body;
  
  if (level !== undefined && ![1, 2, 3].includes(level)) {
    return res.status(400).json({ error: '等级必须为 1、2 或 3' });
  }
  
  const updated = sensitiveWordService.update(
    req.params.id,
    word,
    level as SensitiveWordLevel | undefined
  );
  
  if (!updated) {
    return res.status(404).json({ error: '敏感词不存在' });
  }
  
  return res.json(updated);
});

router.delete('/:id', (req: Request, res: Response) => {
  const result = sensitiveWordService.delete(req.params.id);
  
  if (!result.success) {
    if (result.error && result.error.includes('引用')) {
      return res.status(400).json({ error: result.error });
    }
    return res.status(404).json({ error: result.error || '敏感词不存在' });
  }
  
  return res.status(204).send();
});

export default router;
