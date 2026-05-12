import { Router, Request, Response } from 'express';
import { getAllScales, getScaleById } from '../database';

const router = Router();

router.get('/', async (req: Request, res: Response) => {
  try {
    const scales = await getAllScales();
    res.json(scales);
  } catch (err: any) {
    res.status(500).json({ error: err.message });
  }
});

router.get('/:id', async (req: Request, res: Response) => {
  try {
    const scale = await getScaleById(req.params.id);
    if (!scale) {
      res.status(404).json({ error: 'Scale not found' });
      return;
    }
    res.json(scale);
  } catch (err: any) {
    res.status(500).json({ error: err.message });
  }
});

export default router;
