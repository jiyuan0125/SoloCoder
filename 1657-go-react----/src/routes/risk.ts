import { Router, Request, Response } from 'express';
import { RiskService } from '../services/riskService';
import { Transaction } from '../types';

const router = Router();
const riskService = RiskService.getInstance();

router.post('/assess', (req: Request, res: Response) => {
  try {
    const transaction: Transaction = req.body;

    if (!transaction.id) {
      return res.status(400).json({ error: 'Transaction ID is required' });
    }

    if (transaction.amount === undefined || transaction.amount === null) {
      return res.status(400).json({ error: 'Transaction amount is required' });
    }

    const result = riskService.assessTransaction(transaction);
    const { assessment, isNew } = result;

    const responseData = {
      transactionId: assessment.transactionId,
      score: assessment.score,
      status: assessment.status,
      matchedRuleIds: assessment.matchedRuleIds,
      accountFrozen: assessment.accountFrozen,
      isNew
    };

    if (assessment.freezeFailed) {
      return res.status(500).json({
        ...responseData,
        error: 'Account freeze operation failed. Risk flag is recorded and will be retried.'
      });
    }

    if (assessment.status === 'BLOCKED' || assessment.status === 'PENDING_REVIEW') {
      return res.status(403).json({
        ...responseData,
        message: assessment.status === 'BLOCKED' 
          ? 'Transaction blocked due to high risk' 
          : 'Transaction flagged for manual review'
      });
    }

    if (assessment.status === 'FLAGGED') {
      return res.status(200).json({
        ...responseData,
        flagged: true,
        message: 'Transaction approved but flagged as suspicious'
      });
    }

    res.status(200).json({
      ...responseData,
      message: 'Transaction approved'
    });
  } catch (e) {
    console.error('Assessment error:', e);
    res.status(500).json({ error: 'Internal server error' });
  }
});

router.post('/retry-freezes', (req: Request, res: Response) => {
  try {
    const successCount = riskService.retryFreezes();
    res.json({ successCount, message: `Successfully retried ${successCount} freezes` });
  } catch (e) {
    console.error('Retry freezes error:', e);
    res.status(500).json({ error: 'Internal server error' });
  }
});

export default router;
