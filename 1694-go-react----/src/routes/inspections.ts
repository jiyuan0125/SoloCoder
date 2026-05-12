import { Router, Request, Response, NextFunction } from 'express';
import {
  submitMidTermInspection,
  submitFinalInspection,
  processMidTermInspection,
  processFinalInspection,
  submitRectificationCompletion
} from '../services/inspectionService';

const router = Router();

router.post('/mid-term/submit', (req: Request, res: Response, next: NextFunction) => {
  try {
    const { projectId } = req.body;
    submitMidTermInspection(projectId);
    res.json({
      success: true,
      message: '中期验收申请已提交'
    });
  } catch (err) {
    next(err);
  }
});

router.post('/final/submit', (req: Request, res: Response, next: NextFunction) => {
  try {
    const { projectId } = req.body;
    submitFinalInspection(projectId);
    res.json({
      success: true,
      message: '终期验收申请已提交'
    });
  } catch (err) {
    next(err);
  }
});

router.post('/mid-term/process', (req: Request, res: Response, next: NextFunction) => {
  try {
    const { projectId, result, comments, inspector } = req.body;
    const outcome = processMidTermInspection(projectId, result, comments || null, inspector);
    res.json({
      success: true,
      message: outcome.message || '中期验收处理完成',
      data: outcome
    });
  } catch (err) {
    next(err);
  }
});

router.post('/final/process', (req: Request, res: Response, next: NextFunction) => {
  try {
    const { projectId, result, comments, inspector } = req.body;
    const outcome = processFinalInspection(projectId, result, comments || null, inspector);
    res.json({
      success: true,
      message: outcome.message || '终期验收处理完成',
      data: outcome
    });
  } catch (err) {
    next(err);
  }
});

router.post('/rectification/complete', (req: Request, res: Response, next: NextFunction) => {
  try {
    const { projectId, stage } = req.body;
    submitRectificationCompletion(projectId, stage);
    res.json({
      success: true,
      message: '整改完成提交成功'
    });
  } catch (err) {
    next(err);
  }
});

export default router;
