import { Request, Response } from 'express';
import { visitService } from '../services/visitService';
import { CompleteVisitInput } from '../types';

export const visitController = {
  getVisitPlan: (req: Request, res: Response) => {
    const plan = visitService.getVisitPlanById(req.params.id);
    if (!plan) {
      res.status(404).json({ error: '巡访计划不存在' });
      return;
    }
    res.json(plan);
  },

  getVisitPlansByElder: (req: Request, res: Response) => {
    const plans = visitService.getVisitPlansByElderId(req.params.elderId);
    res.json(plans);
  },

  getPendingVisitPlans: (req: Request, res: Response) => {
    const plans = visitService.getPendingVisitPlans(req.params.elderId);
    res.json(plans);
  },

  completeVisit: (req: Request, res: Response) => {
    const input: CompleteVisitInput = { ...req.body, planId: req.params.id };
    const plan = visitService.completeVisit(input);
    if (!plan) {
      res.status(404).json({ error: '巡访计划不存在' });
      return;
    }
    res.json(plan);
  },

  getReportedVisits: (_req: Request, res: Response) => {
    const plans = visitService.getReportedVisits();
    res.json(plans);
  },
};
