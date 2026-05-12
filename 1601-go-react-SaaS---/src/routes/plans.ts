import { Router, Request, Response } from 'express';
import { planService } from '../services/planService';
import { Plan } from '../types';
import { formatMoneyForResponse } from '../utils/currency';

const router = Router();

router.get('/', (req: Request, res: Response) => {
  const plans = planService.getAllPlans();
  const response = plans.map((plan: Plan) => ({
    id: plan.id,
    name: plan.name,
    priceMonthly: formatMoneyForResponse(plan.priceMonthlyCents),
    maxUsers: plan.maxUsers,
    apiCallsLimit: plan.apiCallsLimit,
    storageLimitGB: plan.storageLimitGB,
    overageApiCallCost: formatMoneyForResponse(plan.overageApiCallCostCents),
    overageStorageCostPerGB: formatMoneyForResponse(plan.overageStorageCostCentsPerGB),
  }));
  res.json(response);
});

router.delete('/:planId', (req: Request, res: Response) => {
  const { planId } = req.params;
  const result = planService.deletePlan(planId as any);

  if (!result.success) {
    return res.status(400).json({ error: result.message });
  }

  res.json({ success: true });
});

export default router;
