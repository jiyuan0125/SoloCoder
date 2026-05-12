import { Router, Request, Response } from 'express';
import { startVerification, completeVerification, getVerificationRecords } from '../services/verificationService';
import { VerificationResult } from '../types';

const router = Router();

router.post('/start', (req: Request, res: Response) => {
  const { reportId, verifierId, verifierName } = req.body;

  if (!reportId || !verifierId || !verifierName) {
    return res.status(400).json({ error: '缺少必要参数' });
  }

  try {
    const result = startVerification({ reportId, verifierId, verifierName });

    if (result.notFound) {
      return res.status(404).json({ error: '上报记录不存在' });
    }

    if (result.conflict) {
      return res.status(409).json({ error: '该灾情正在核查中，无法重复核查' });
    }

    if (!result.success) {
      return res.status(400).json({ error: '当前状态不允许核查' });
    }

    return res.status(201).json({
      message: '核查已开始',
      record: result.record
    });
  } catch (error) {
    console.error('开始核查失败:', error);
    return res.status(500).json({ error: '服务器内部错误' });
  }
});

router.post('/complete', (req: Request, res: Response) => {
  const { reportId, verifierId, result, comments } = req.body;

  if (!reportId || !verifierId || !result) {
    return res.status(400).json({ error: '缺少必要参数' });
  }

  if (result !== VerificationResult.PASSED && result !== VerificationResult.NEEDS_SUPPLEMENT) {
    return res.status(400).json({ error: '核查结果无效' });
  }

  try {
    const completeResult = completeVerification({ reportId, verifierId, result, comments });

    if (completeResult.notFound) {
      return res.status(404).json({ error: '上报记录不存在' });
    }

    if (completeResult.conflict) {
      return res.status(409).json({ error: '该灾情没有进行中的核查任务' });
    }

    return res.json({
      message: result === VerificationResult.PASSED ? '核查通过，数据已确认' : '需补充信息，已退回修改',
      result
    });
  } catch (error) {
    console.error('完成核查失败:', error);
    return res.status(500).json({ error: '服务器内部错误' });
  }
});

router.get('/:reportId', (req: Request, res: Response) => {
  const { reportId } = req.params;

  try {
    const records = getVerificationRecords(reportId);
    return res.json(records);
  } catch (error) {
    console.error('查询核查记录失败:', error);
    return res.status(500).json({ error: '服务器内部错误' });
  }
});

export default router;
