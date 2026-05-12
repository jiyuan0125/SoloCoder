import express, { Request, Response } from 'express';
import * as salespersonService from '../services/salespersonService';

const router = express.Router();

router.get('/', (_req: Request, res: Response) => {
  const salespeople = salespersonService.getSalespeople();
  res.json(salespeople);
});

router.get('/:id', (req: Request, res: Response) => {
  const id = Array.isArray(req.params.id) ? req.params.id[0] : req.params.id;
  const salesperson = salespersonService.getSalespersonById(id);
  if (salesperson) {
    res.json(salesperson);
  } else {
    res.status(404).json({ error: '销售不存在' });
  }
});

router.post('/', (req: Request, res: Response) => {
  const { name, teamId } = req.body;
  if (!name || !teamId) {
    return res.status(400).json({ error: 'name 和 teamId 是必填项' });
  }
  const created = salespersonService.createSalesperson({ name, teamId });
  res.status(201).json(created);
});

router.put('/:id', (req: Request, res: Response) => {
  const id = Array.isArray(req.params.id) ? req.params.id[0] : req.params.id;
  const updated = salespersonService.updateSalesperson(id, req.body);
  if (updated) {
    res.json(updated);
  } else {
    res.status(404).json({ error: '销售不存在' });
  }
});

router.delete('/:id', (req: Request, res: Response) => {
  const id = Array.isArray(req.params.id) ? req.params.id[0] : req.params.id;
  const deleted = salespersonService.deleteSalesperson(id);
  if (deleted) {
    res.status(204).send();
  } else {
    res.status(404).json({ error: '销售不存在' });
  }
});

export default router;
