import { Router, Request, Response } from 'express';
import { sensitiveWordManager } from '../utils/sensitiveWords';

const router = Router();

router.get('/', (req: Request, res: Response) => {
  res.json({
    words: sensitiveWordManager.list(),
    version: sensitiveWordManager.getVersion()
  });
});

router.post('/', (req: Request, res: Response) => {
  const { word } = req.body as { word?: string };
  if (!word || typeof word !== 'string' || word.length === 0) {
    res.status(400).json({ error: '缺少敏感词' });
    return;
  }
  sensitiveWordManager.add(word);
  res.status(201).json({ success: true });
});

router.delete('/', (req: Request, res: Response) => {
  const { word } = req.body as { word?: string };
  if (!word || typeof word !== 'string') {
    res.status(400).json({ error: '缺少敏感词' });
    return;
  }
  sensitiveWordManager.remove(word);
  res.json({ success: true });
});

router.put('/', (req: Request, res: Response) => {
  const { oldWord, newWord } = req.body as { oldWord?: string; newWord?: string };
  if (!oldWord || !newWord || typeof oldWord !== 'string' || typeof newWord !== 'string') {
    res.status(400).json({ error: '缺少参数' });
    return;
  }
  sensitiveWordManager.update(oldWord, newWord);
  res.json({ success: true });
});

export default router;
