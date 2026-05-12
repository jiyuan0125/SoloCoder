import { Router, Request, Response } from 'express';
import * as ratingService from '../services/ratingService';

const router = Router();

router.get('/user/:userId', async (req: Request, res: Response) => {
  try {
    const userId = parseInt(req.params.userId, 10);
    if (isNaN(userId)) {
      return res.status(400).json({ error: 'Invalid user ID' });
    }
    const ratings = await ratingService.getRatingsByUser(userId);
    res.json(ratings);
  } catch (err) {
    res.status(500).json({ error: 'Internal server error' });
  }
});

router.get('/user/:userId/product/:productId', async (req: Request, res: Response) => {
  try {
    const userId = parseInt(req.params.userId, 10);
    const productId = parseInt(req.params.productId, 10);
    if (isNaN(userId) || isNaN(productId)) {
      return res.status(400).json({ error: 'Invalid user or product ID' });
    }
    const rating = await ratingService.getRating(userId, productId);
    if (!rating) {
      return res.status(404).json({ error: 'Rating not found' });
    }
    res.json(rating);
  } catch (err) {
    res.status(500).json({ error: 'Internal server error' });
  }
});

router.post('/', async (req: Request, res: Response) => {
  try {
    const { userId, productId, score } = req.body;
    if (
      typeof userId !== 'number' ||
      typeof productId !== 'number' ||
      typeof score !== 'number'
    ) {
      return res.status(400).json({ error: 'userId, productId, and score must be numbers' });
    }
    if (!ratingService.isValidScore(score)) {
      return res.status(400).json({ error: 'Score must be an integer between 1 and 5' });
    }
    const rating = await ratingService.upsertRating(userId, productId, score);
    res.status(201).json(rating);
  } catch (err) {
    res.status(500).json({ error: 'Internal server error' });
  }
});

router.put('/user/:userId/product/:productId', async (req: Request, res: Response) => {
  try {
    const userId = parseInt(req.params.userId, 10);
    const productId = parseInt(req.params.productId, 10);
    const { score } = req.body;
    if (isNaN(userId) || isNaN(productId)) {
      return res.status(400).json({ error: 'Invalid user or product ID' });
    }
    if (typeof score !== 'number') {
      return res.status(400).json({ error: 'Score must be a number' });
    }
    if (!ratingService.isValidScore(score)) {
      return res.status(400).json({ error: 'Score must be an integer between 1 and 5' });
    }
    const rating = await ratingService.updateRating(userId, productId, score);
    if (!rating) {
      return res.status(404).json({ error: 'Rating not found' });
    }
    res.json(rating);
  } catch (err) {
    res.status(500).json({ error: 'Internal server error' });
  }
});

router.delete('/user/:userId/product/:productId', async (req: Request, res: Response) => {
  try {
    const userId = parseInt(req.params.userId, 10);
    const productId = parseInt(req.params.productId, 10);
    if (isNaN(userId) || isNaN(productId)) {
      return res.status(400).json({ error: 'Invalid user or product ID' });
    }
    const success = await ratingService.deleteRating(userId, productId);
    if (!success) {
      return res.status(404).json({ error: 'Rating not found' });
    }
    res.status(204).send();
  } catch (err) {
    res.status(500).json({ error: 'Internal server error' });
  }
});

export default router;
