import { Router, Request, Response } from 'express';
import * as planService from '../services/planService';

const router = Router();

function getOperator(req: Request): string {
  return (req.headers['x-operator'] as string) || 'system';
}

function handleError(res: Response, err: any): void {
  const errorMap: Record<string, { status: number; message: string }> = {
    'HOUSEHOLD_NOT_FOUND': { status: 404, message: '户ID不存在' },
    'PLAN_NOT_FOUND': { status: 404, message: '帮扶计划不存在' },
    'PLAN_ALREADY_EXISTS': { status: 400, message: '该户已存在帮扶计划' },
    'INVALID_MEASURE_TYPE': { status: 400, message: '措施类型不在范围内' },
    'INVALID_EVALUATION_RESULT': { status: 400, message: '评估结果不在范围内' },
    'MEASURE_NOT_FOUND': { status: 404, message: '帮扶措施不存在' },
    'EVALUATION_RECORD_NOT_FOUND': { status: 404, message: '评估记录不存在' },
    'NO_REDRAFT_NEEDED': { status: 400, message: '该措施无需重新制定' },
    'NOT_IN_MONITORING_PERIOD': { status: 400, message: '不在跟踪监测期内' }
  };

  const mapped = errorMap[err.message];
  if (mapped) {
    res.status(mapped.status).json({ error: mapped.message });
  } else {
    res.status(500).json({ error: err.message });
  }
}

router.post('/household/:householdId', (req: Request, res: Response) => {
  try {
    const operator = getOperator(req);
    const plan = planService.createPlan(req.params.householdId, req.body, operator);
    res.status(201).json(plan);
  } catch (err) {
    handleError(res, err);
  }
});

router.get('/household/:householdId', (req: Request, res: Response) => {
  try {
    const plan = planService.getPlanByHousehold(req.params.householdId);
    if (!plan) {
      res.status(404).json({ error: '帮扶计划不存在' });
      return;
    }
    res.json(plan);
  } catch (err) {
    handleError(res, err);
  }
});

router.get('/:planId', (req: Request, res: Response) => {
  try {
    const plan = planService.getPlanById(req.params.planId);
    if (!plan) {
      res.status(404).json({ error: '帮扶计划不存在' });
      return;
    }
    res.json(plan);
  } catch (err) {
    handleError(res, err);
  }
});

router.put('/household/:householdId/adjust', (req: Request, res: Response) => {
  try {
    const operator = getOperator(req);
    const { measures, reason } = req.body;
    const plan = planService.adjustPlan(req.params.householdId, measures, reason, operator);
    res.json(plan);
  } catch (err) {
    handleError(res, err);
  }
});

router.post('/household/:householdId/approve', (req: Request, res: Response) => {
  try {
    const operator = getOperator(req);
    const plan = planService.approvePlan(req.params.householdId, operator);
    res.json(plan);
  } catch (err) {
    handleError(res, err);
  }
});

router.post('/:planId/measures/:measureId/evaluate', (req: Request, res: Response) => {
  try {
    const operator = getOperator(req);
    const plan = planService.addEvaluation(
      req.params.planId,
      req.params.measureId,
      req.body,
      operator
    );
    res.json(plan);
  } catch (err) {
    handleError(res, err);
  }
});

router.get('/:planId/measures/:measureId/redraft', (req: Request, res: Response) => {
  try {
    const data = planService.getMeasureForRedraft(
      req.params.planId,
      req.params.measureId
    );
    res.json(data);
  } catch (err) {
    handleError(res, err);
  }
});

router.post('/:planId/measures/:measureId/redraft', (req: Request, res: Response) => {
  try {
    const operator = getOperator(req);
    const plan = planService.redraftMeasure(
      req.params.planId,
      req.params.measureId,
      req.body,
      operator
    );
    res.json(plan);
  } catch (err) {
    handleError(res, err);
  }
});

router.post('/household/:householdId/return-to-poverty', (req: Request, res: Response) => {
  try {
    const operator = getOperator(req);
    const result = planService.handleReturnToPoverty(req.params.householdId, operator);
    res.json(result);
  } catch (err) {
    handleError(res, err);
  }
});

router.get('/:planId/histories', (req: Request, res: Response) => {
  try {
    const histories = planService.getPlanHistories(req.params.planId);
    res.json(histories);
  } catch (err) {
    handleError(res, err);
  }
});

router.get('/', (req: Request, res: Response) => {
  try {
    const plans = planService.getAllPlans();
    res.json(plans);
  } catch (err) {
    handleError(res, err);
  }
});

export default router;
