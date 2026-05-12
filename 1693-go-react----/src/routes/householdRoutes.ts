import { Router, Request, Response } from 'express';
import * as householdService from '../services/householdService';

const router = Router();

function getOperator(req: Request): string {
  return (req.headers['x-operator'] as string) || 'system';
}

function handleError(res: Response, err: any): void {
  const errorMap: Record<string, { status: number; message: string }> = {
    'HOUSEHOLD_NOT_FOUND': { status: 404, message: '户ID不存在' },
    'POVERTY_CAUSES_REQUIRED': { status: 400, message: '致贫原因为空' },
    'INVALID_POVERTY_CAUSE': { status: 400, message: '无效的致贫原因' },
    'INVALID_POVERTY_LEVEL': { status: 400, message: '贫困户级别不在范围内' },
    'NOT_IN_MONITORING_PERIOD': { status: 400, message: '不在跟踪监测期内' }
  };

  const mapped = errorMap[err.message];
  if (mapped) {
    res.status(mapped.status).json({ error: mapped.message });
  } else {
    res.status(500).json({ error: err.message });
  }
}

router.post('/', (req: Request, res: Response) => {
  try {
    const operator = getOperator(req);
    const household = householdService.createHousehold(req.body, operator);
    res.status(201).json(household);
  } catch (err) {
    handleError(res, err);
  }
});

router.get('/:householdId', (req: Request, res: Response) => {
  try {
    const household = householdService.getHouseholdById(req.params.householdId);
    if (!household) {
      res.status(404).json({ error: '户ID不存在' });
      return;
    }
    res.json(household);
  } catch (err) {
    handleError(res, err);
  }
});

router.put('/:householdId/archive', (req: Request, res: Response) => {
  try {
    const operator = getOperator(req);
    const household = householdService.updateArchive(
      req.params.householdId,
      req.body,
      operator
    );
    res.json(household);
  } catch (err) {
    handleError(res, err);
  }
});

router.post('/:householdId/out-of-poverty', (req: Request, res: Response) => {
  try {
    const operator = getOperator(req);
    const household = householdService.markOutOfPoverty(req.params.householdId, operator);
    res.json(household);
  } catch (err) {
    handleError(res, err);
  }
});

router.post('/:householdId/return-monitoring', (req: Request, res: Response) => {
  try {
    const operator = getOperator(req);
    const household = householdService.markReturnMonitoring(req.params.householdId, operator);
    res.json(household);
  } catch (err) {
    handleError(res, err);
  }
});

router.get('/:householdId/histories', (req: Request, res: Response) => {
  try {
    const histories = householdService.getHouseholdHistories(req.params.householdId);
    res.json(histories);
  } catch (err) {
    handleError(res, err);
  }
});

router.get('/', (req: Request, res: Response) => {
  try {
    const households = householdService.getAllHouseholds();
    res.json(households);
  } catch (err) {
    handleError(res, err);
  }
});

export default router;
