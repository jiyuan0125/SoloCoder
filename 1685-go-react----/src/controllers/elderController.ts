import { Request, Response } from 'express';
import { elderService } from '../services/elderService';
import { CreateElderInput } from '../types';

export const elderController = {
  createElder: (req: Request, res: Response) => {
    try {
      const input: CreateElderInput = req.body;
      const elder = elderService.createElder(input);
      res.status(201).json(elder);
    } catch (error: any) {
      res.status(400).json({ error: error.message });
    }
  },

  getElder: (req: Request, res: Response) => {
    const elder = elderService.getElderById(req.params.id);
    if (!elder) {
      res.status(404).json({ error: '老人不存在' });
      return;
    }
    res.json(elder);
  },

  getAllElders: (_req: Request, res: Response) => {
    const elders = elderService.getAllElders();
    res.json(elders);
  },
};
