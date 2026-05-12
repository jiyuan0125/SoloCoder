import { Router, Request, Response } from 'express';
import {
  search,
  getSearchSuggestions,
  getHotWords,
  getNoResultRate
} from '../services/searchService';

const router = Router();

router.get('/', (req: Request, res: Response) => {
  const query = req.query.q as string;
  
  if (!query || query.trim() === '') {
    return res.status(400).json({
      success: false,
      message: '搜索词不能为空'
    });
  }
  
  const result = search(query);
  
  if (!result.success) {
    return res.status(400).json(result);
  }
  
  res.json(result);
});

router.get('/suggest', (req: Request, res: Response) => {
  const prefix = req.query.prefix as string || '';
  const suggestions = getSearchSuggestions(prefix);
  
  res.json({
    success: true,
    suggestions
  });
});

router.get('/hotwords', (req: Request, res: Response) => {
  const limit = parseInt(req.query.limit as string) || 20;
  const hotwords = getHotWords(Math.min(limit, 50));
  
  res.json({
    success: true,
    data: hotwords
  });
});

router.get('/no-result-rate', (req: Request, res: Response) => {
  const stats = getNoResultRate();
  
  res.json({
    success: true,
    data: stats
  });
});

export default router;
