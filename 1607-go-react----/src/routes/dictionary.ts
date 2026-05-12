import { Router, Request, Response } from 'express';
import {
  listDictionary,
  addDictionaryWord,
  deleteDictionaryWord,
  resetDictionary
} from '../services/dictionaryService';

const router = Router();

router.get('/', (req: Request, res: Response) => {
  const words = listDictionary();
  
  res.json({
    success: true,
    data: words,
    total: words.length
  });
});

router.post('/', (req: Request, res: Response) => {
  const { word } = req.body;
  
  if (!word) {
    return res.status(400).json({
      success: false,
      message: '词条不能为空'
    });
  }
  
  const result = addDictionaryWord(word);
  
  if (!result.success) {
    return res.status(400).json(result);
  }
  
  res.status(201).json(result);
});

router.delete('/:word', (req: Request, res: Response) => {
  const { word } = req.params;
  const result = deleteDictionaryWord(word);
  
  if (!result.success) {
    return res.status(400).json(result);
  }
  
  res.json(result);
});

router.post('/reset', (req: Request, res: Response) => {
  const result = resetDictionary();
  res.json(result);
});

export default router;
