import { Router, Request, Response, NextFunction } from 'express';
import {
  allocateFund,
  recordExpenditure,
  submitBudgetAdjustment,
  processBudgetAdjustment
} from '../services/fundService';
import { BudgetCategory } from '../types';

const router = Router();

router.post('/allocate', (req: Request, res: Response, next: NextFunction) => {
  try {
    const { projectId, stage, percentage, amount, allocationDate, paymentMethod, receivingAccount } = req.body;
    allocateFund(projectId, stage, percentage, amount, allocationDate, paymentMethod, receivingAccount);
    
    const stageNames: Record<string, string> = {
      initial: '启动资金',
      mid_term: '中期资金',
      final: '终期资金'
    };
    
    res.json({
      success: true,
      message: `${stageNames[stage] || stage}拨付成功`
    });
  } catch (err) {
    next(err);
  }
});

router.post('/expenditure', (req: Request, res: Response, next: NextFunction) => {
  try {
    const { projectId, category, amount, description, expenditureDate } = req.body;
    recordExpenditure(projectId, category as BudgetCategory, amount, description, expenditureDate);
    res.json({
      success: true,
      message: '支出记录成功'
    });
  } catch (err) {
    next(err);
  }
});

router.post('/budget-adjustment', (req: Request, res: Response, next: NextFunction) => {
  try {
    const { projectId, fromCategory, toCategory, amount, reason } = req.body;
    submitBudgetAdjustment(projectId, fromCategory as BudgetCategory, toCategory as BudgetCategory, amount, reason);
    res.json({
      success: true,
      message: '预算调剂申请已提交'
    });
  } catch (err) {
    next(err);
  }
});

router.post('/budget-adjustment/:id/process', (req: Request, res: Response, next: NextFunction) => {
  try {
    const requestId = parseInt(req.params.id, 10);
    const { status, approver } = req.body;
    processBudgetAdjustment(requestId, status, approver);
    res.json({
      success: true,
      message: status === 'approved' ? '预算调剂申请已批准' : '预算调剂申请已拒绝'
    });
  } catch (err) {
    next(err);
  }
});

export default router;
