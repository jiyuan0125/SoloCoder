import { Router, Request, Response } from 'express';
import { publishReport, updatePublished, getPublishedRecords, getLatestPublished } from '../services/publishService';

const router = Router();

router.post('/', (req: Request, res: Response) => {
  const { reportId, publisherId, publisherName, content } = req.body;

  if (!reportId || !publisherId || !publisherName || !content) {
    return res.status(400).json({ error: '缺少必要参数' });
  }

  try {
    const result = publishReport({ reportId, publisherId, publisherName, content });

    if (result.notFound) {
      return res.status(404).json({ error: '灾情记录不存在' });
    }

    if (result.notVerified) {
      return res.status(400).json({ error: '未核查的灾情不能发布' });
    }

    return res.status(201).json({
      message: '发布成功',
      record: result.record
    });
  } catch (error) {
    console.error('发布失败:', error);
    return res.status(500).json({ error: '服务器内部错误' });
  }
});

router.put('/:reportId', (req: Request, res: Response) => {
  const { reportId } = req.params;
  const { publisherId, publisherName, content } = req.body;

  if (!publisherId || !publisherName || !content) {
    return res.status(400).json({ error: '缺少必要参数' });
  }

  try {
    const result = updatePublished({ reportId, publisherId, publisherName, content });

    if (result.notFound) {
      return res.status(404).json({ error: '未找到已发布记录' });
    }

    if (result.limitExceeded) {
      return res.status(400).json({ error: '发布内容修改不能超过3次' });
    }

    return res.json({
      message: '修改成功',
      record: result.record
    });
  } catch (error) {
    console.error('修改发布内容失败:', error);
    return res.status(500).json({ error: '服务器内部错误' });
  }
});

router.get('/:reportId', (req: Request, res: Response) => {
  const { reportId } = req.params;
  const { latest } = req.query;

  try {
    if (latest === 'true') {
      const record = getLatestPublished(reportId);
      if (!record) {
        return res.status(404).json({ error: '未找到已发布记录' });
      }
      return res.json(record);
    }

    const records = getPublishedRecords(reportId);
    return res.json(records);
  } catch (error) {
    console.error('查询发布记录失败:', error);
    return res.status(500).json({ error: '服务器内部错误' });
  }
});

export default router;
