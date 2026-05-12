import { Router, Request, Response } from 'express';
import { createModificationRequest, approveRequest, rejectRequest, getPendingRequests } from '../services/approvalService';

const router = Router();

router.post('/', (req: Request, res: Response) => {
  const { reportId, requesterId, requesterName, changes, reason } = req.body;

  if (!reportId || !requesterId || !requesterName || !changes) {
    return res.status(400).json({ error: '缺少必要参数' });
  }

  try {
    const result = createModificationRequest({ reportId, requesterId, requesterName, changes, reason });

    if (result.notFound) {
      return res.status(404).json({ error: '灾情记录不存在' });
    }

    return res.status(201).json({
      message: '修改申请已提交，等待市级审批',
      requestId: result.requestId
    });
  } catch (error) {
    console.error('提交修改申请失败:', error);
    return res.status(500).json({ error: '服务器内部错误' });
  }
});

router.post('/:requestId/approve', (req: Request, res: Response) => {
  const { requestId } = req.params;
  const { approverId } = req.body;

  if (!approverId) {
    return res.status(400).json({ error: '缺少审批人ID' });
  }

  try {
    const result = approveRequest({ requestId, approverId });

    if (result.notFound) {
      return res.status(404).json({ error: '申请记录不存在' });
    }

    return res.json({ message: '审批通过，数据已更新' });
  } catch (error) {
    console.error('审批失败:', error);
    return res.status(500).json({ error: '服务器内部错误' });
  }
});

router.post('/:requestId/reject', (req: Request, res: Response) => {
  const { requestId } = req.params;
  const { approverId } = req.body;

  if (!approverId) {
    return res.status(400).json({ error: '缺少审批人ID' });
  }

  try {
    const result = rejectRequest(requestId, approverId);

    if (result.notFound) {
      return res.status(404).json({ error: '申请记录不存在' });
    }

    return res.json({ message: '申请已拒绝' });
  } catch (error) {
    console.error('拒绝申请失败:', error);
    return res.status(500).json({ error: '服务器内部错误' });
  }
});

router.get('/pending', (_req: Request, res: Response) => {
  try {
    const requests = getPendingRequests();
    return res.json(requests);
  } catch (error) {
    console.error('查询待审批申请失败:', error);
    return res.status(500).json({ error: '服务器内部错误' });
  }
});

export default router;
