import { Request, Response } from 'express';
import { healthService } from '../services/healthService';
import { CreateHealthRecordInput } from '../types';

export const healthController = {
  createHealthRecord: (req: Request, res: Response) => {
    try {
      const input: CreateHealthRecordInput = req.body;
      const record = healthService.createHealthRecord(input);
      if (!record) {
        res.status(404).json({ error: '老人不存在' });
        return;
      }
      res.status(201).json(record);
    } catch (error: any) {
      res.status(400).json({ error: error.message });
    }
  },

  getHealthRecord: (req: Request, res: Response) => {
    const record = healthService.getHealthRecordById(req.params.id);
    if (!record) {
      res.status(404).json({ error: '健康记录不存在' });
      return;
    }
    res.json(record);
  },

  getHealthRecordsByElder: (req: Request, res: Response) => {
    const records = healthService.getHealthRecordsByElderId(req.params.elderId);
    res.json(records);
  },

  getAbnormalRecords: (req: Request, res: Response) => {
    const elderId = req.query.elderId as string | undefined;
    const records = healthService.getAbnormalRecords(elderId);
    res.json(records);
  },
};
