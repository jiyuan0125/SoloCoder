import express, { Request, Response } from 'express';
import { OpportunityService } from '../services/opportunityService';
import { OpportunityStage } from '../types';

const router = express.Router();

router.post('/', (req: Request, res: Response) => {
  try {
    const { lead_id, expected_amount, assigned_sales_id, expected_close_date } = req.body;
    
    if (!lead_id) {
      return res.status(400).json({ error: '线索ID不能为空' });
    }
    if (expected_amount === undefined || expected_amount === null) {
      return res.status(400).json({ error: '预计金额不能为空' });
    }
    if (expected_amount < 0) {
      return res.status(400).json({ error: '预计金额不能为负数' });
    }
    if (!assigned_sales_id) {
      return res.status(400).json({ error: '负责销售人员ID不能为空' });
    }

    const opportunity = OpportunityService.createOpportunity({
      lead_id,
      expected_amount,
      assigned_sales_id,
      expected_close_date,
    });

    res.status(201).json(opportunity);
  } catch (error: any) {
    if (error.message === '只有已转化的线索才能创建商机') {
      return res.status(400).json({ error: error.message });
    }
    res.status(400).json({ error: error.message });
  }
});

router.get('/', (req: Request, res: Response) => {
  try {
    const opportunities = OpportunityService.getAllOpportunities();
    res.json(opportunities);
  } catch (error: any) {
    res.status(500).json({ error: error.message });
  }
});

router.get('/funnel', (req: Request, res: Response) => {
  try {
    const stats = OpportunityService.getFunnelStatistics();
    res.json(stats);
  } catch (error: any) {
    res.status(500).json({ error: error.message });
  }
});

router.get('/:id', (req: Request, res: Response) => {
  try {
    const opportunity = OpportunityService.getOpportunityById(req.params.id);
    if (!opportunity) {
      return res.status(404).json({ error: '商机不存在' });
    }
    res.json(opportunity);
  } catch (error: any) {
    res.status(500).json({ error: error.message });
  }
});

router.put('/:id/stage', (req: Request, res: Response) => {
  try {
    const { stage, sales_person_id, reason, lost_to } = req.body;
    
    if (!stage) {
      return res.status(400).json({ error: '阶段不能为空' });
    }
    if (!sales_person_id) {
      return res.status(400).json({ error: '销售人员ID不能为空' });
    }

    if (!OpportunityService.isValidStage(stage)) {
      return res.status(400).json({ error: '无效的阶段值' });
    }

    const opportunity = OpportunityService.updateStage(req.params.id, {
      stage: stage as OpportunityStage,
      sales_person_id,
      reason,
      lost_to,
    });

    res.json(opportunity);
  } catch (error: any) {
    if (error.message === '商机不存在') {
      return res.status(404).json({ error: error.message });
    }
    if (error.message === '赢单和输单之后不能再修改阶段') {
      return res.status(400).json({ error: error.message });
    }
    if (error.message === '输单需要填写输给谁') {
      return res.status(400).json({ error: error.message });
    }
    res.status(400).json({ error: error.message });
  }
});

export default router;
