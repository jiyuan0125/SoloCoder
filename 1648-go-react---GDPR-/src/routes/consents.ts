import { Router, Request, Response } from 'express';
import { v4 as uuidv4 } from 'uuid';
import { Consent } from '../types';
import { store } from '../store';

const router = Router();

router.post('/', (req: Request, res: Response) => {
  const { userId, categoryId, granted } = req.body;

  if (!userId || typeof userId !== 'string') {
    return res.status(400).json({ error: 'userId is required and must be a string' });
  }

  if (!categoryId || typeof categoryId !== 'string') {
    return res.status(400).json({ error: 'categoryId is required and must be a string' });
  }

  const category = store.getDataCategory(categoryId);
  if (!category) {
    return res.status(404).json({ error: 'Data category not found' });
  }

  const consent: Consent = {
    id: uuidv4(),
    userId,
    categoryId,
    granted: granted === true,
    grantedAt: new Date()
  };

  store.addConsent(consent);
  return res.status(201).json(consent);
});

router.delete('/:consentId', (req: Request, res: Response) => {
  const consent = store.getConsent(req.params.consentId);
  if (!consent) {
    return res.status(404).json({ error: 'Consent not found' });
  }

  if (!consent.granted || consent.withdrawnAt) {
    return res.status(400).json({ error: 'Consent is already withdrawn or not granted' });
  }

  const updated = store.addConsent({
    ...consent,
    granted: false,
    withdrawnAt: new Date()
  });

  return res.json(updated);
});

router.get('/:id', (req: Request, res: Response) => {
  const consent = store.getConsent(req.params.id);
  if (!consent) {
    return res.status(404).json({ error: 'Consent not found' });
  }
  return res.json(consent);
});

export default router;
