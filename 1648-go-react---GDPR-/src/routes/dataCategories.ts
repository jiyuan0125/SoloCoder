import { Router, Request, Response } from 'express';
import { v4 as uuidv4 } from 'uuid';
import { DataCategory, LegalBasis } from '../types';
import { store } from '../store';
import { isValidLegalBasis } from '../utils';

const router = Router();

interface CreateDataCategoryBody {
  name: string;
  purpose: string;
  legalBasis: LegalBasis;
  retentionDays: number;
}

router.post('/', (req: Request, res: Response) => {
  const body = req.body as Partial<CreateDataCategoryBody>;
  const { name, purpose, legalBasis, retentionDays } = body;

  if (!name || typeof name !== 'string' || name.trim() === '') {
    return res.status(400).json({ error: 'name is required and must be a non-empty string' });
  }

  if (!purpose || typeof purpose !== 'string' || purpose.trim() === '') {
    return res.status(400).json({ error: 'purpose is required and must be a non-empty string' });
  }

  if (!legalBasis || !isValidLegalBasis(legalBasis)) {
    return res.status(400).json({ error: `legalBasis must be one of: consent, contract, legal_obligation, legitimate_interest` });
  }

  if (retentionDays === undefined || typeof retentionDays !== 'number' || !Number.isInteger(retentionDays) || retentionDays <= 0) {
    return res.status(400).json({ error: 'retentionDays is required and must be a positive integer greater than 0' });
  }

  const category: DataCategory = {
    id: uuidv4(),
    name: name.trim(),
    purpose: purpose.trim(),
    legalBasis,
    retentionDays,
    createdAt: new Date()
  };

  store.addDataCategory(category);
  return res.status(201).json(category);
});

router.get('/', (req: Request, res: Response) => {
  return res.json(store.getAllDataCategories());
});

router.get('/:id', (req: Request, res: Response) => {
  const category = store.getDataCategory(req.params.id);
  if (!category) {
    return res.status(404).json({ error: 'Data category not found' });
  }
  return res.json(category);
});

export default router;
