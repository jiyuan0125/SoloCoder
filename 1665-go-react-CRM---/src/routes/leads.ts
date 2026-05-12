import express, { Request, Response } from 'express';
import { LeadService } from '../services/leadService';
import { LeadStatus } from '../types';

const router = express.Router();

router.post('/', (req: Request, res: Response) => {
  try {
    const { source_channel, contact_info, company_name, contact_person, requirements } = req.body;
    
    if (!source_channel || !source_channel.trim()) {
      return res.status(400).json({ error: '来源渠道不能为空' });
    }
    if (!contact_info || !contact_info.trim()) {
      return res.status(400).json({ error: '联系方式不能为空' });
    }

    const lead = LeadService.createLead({
      source_channel,
      contact_info,
      company_name,
      contact_person,
      requirements,
    });

    res.status(201).json(lead);
  } catch (error: any) {
    res.status(400).json({ error: error.message });
  }
});

router.get('/', (req: Request, res: Response) => {
  try {
    const leads = LeadService.getAllLeads();
    res.json(leads);
  } catch (error: any) {
    res.status(500).json({ error: error.message });
  }
});

router.get('/pool', (req: Request, res: Response) => {
  try {
    const leads = LeadService.getPoolLeads();
    res.json(leads);
  } catch (error: any) {
    res.status(500).json({ error: error.message });
  }
});

router.get('/:id', (req: Request, res: Response) => {
  try {
    const lead = LeadService.getLeadById(req.params.id);
    if (!lead) {
      return res.status(404).json({ error: '线索不存在' });
    }
    res.json(lead);
  } catch (error: any) {
    res.status(500).json({ error: error.message });
  }
});

router.post('/:id/claim', (req: Request, res: Response) => {
  try {
    const { sales_person_id } = req.body;
    
    if (!sales_person_id) {
      return res.status(400).json({ error: '销售人员ID不能为空' });
    }

    const lead = LeadService.claimLead(req.params.id, sales_person_id);
    res.json(lead);
  } catch (error: any) {
    if (error.message === '线索已被认领') {
      return res.status(409).json({ error: '线索已被认领' });
    }
    if (error.message === '线索不存在') {
      return res.status(404).json({ error: error.message });
    }
    res.status(400).json({ error: error.message });
  }
});

router.put('/:id/status', (req: Request, res: Response) => {
  try {
    const { status, sales_person_id, reason } = req.body;
    
    if (!status) {
      return res.status(400).json({ error: '状态不能为空' });
    }
    if (!sales_person_id) {
      return res.status(400).json({ error: '销售人员ID不能为空' });
    }

    const validStatuses: LeadStatus[] = ['new', 'contacted', 'rejected', 'converted'];
    if (!validStatuses.includes(status as LeadStatus)) {
      return res.status(400).json({ error: '无效的状态值' });
    }

    const lead = LeadService.updateLeadStatus(req.params.id, {
      status: status as LeadStatus,
      sales_person_id,
      reason,
    });

    res.json(lead);
  } catch (error: any) {
    if (error.message === '线索不存在') {
      return res.status(404).json({ error: error.message });
    }
    if (
      error.message === '已拒绝的线索不能再修改状态' ||
      error.message === '已转化的线索不能再修改状态'
    ) {
      return res.status(400).json({ error: error.message });
    }
    res.status(400).json({ error: error.message });
  }
});

export default router;
