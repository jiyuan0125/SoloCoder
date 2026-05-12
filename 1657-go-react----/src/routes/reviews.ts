import { Router, Request, Response } from 'express';
import { ReviewService, ReviewNotFoundException } from '../services/reviewService';

const router = Router();
const reviewService = ReviewService.getInstance();

router.get('/pending', (req: Request, res: Response) => {
  const pending = reviewService.findPending();
  res.json(pending);
});

router.get('/:id', (req: Request, res: Response) => {
  try {
    const review = reviewService.getReviewWithAssessment(req.params.id);
    res.json(review);
  } catch (e) {
    if (e instanceof ReviewNotFoundException) {
      return res.status(404).json({ error: 'Review record not found' });
    }
    res.status(500).json({ error: 'Internal server error' });
  }
});

router.post('/:id/decide', (req: Request, res: Response) => {
  try {
    const { decision, reviewerId } = req.body;

    if (decision !== 'NORMAL' && decision !== 'FRAUD') {
      return res.status(400).json({ 
        error: 'Invalid decision. Must be either NORMAL or FRAUD.' 
      });
    }

    const result = reviewService.decide(req.params.id, decision, reviewerId);
    res.json(result);
  } catch (e) {
    if (e instanceof ReviewNotFoundException) {
      return res.status(404).json({ error: 'Review record not found' });
    }
    console.error('Review decision error:', e);
    res.status(500).json({ error: 'Internal server error' });
  }
});

export default router;
