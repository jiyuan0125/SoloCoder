import { Router, Request, Response } from 'express';
import { createReport, getReportById, getAllReports, updateReport } from '../services/reportService';
import { ReportStatus } from '../types';

const router = Router();

router.post('/', (req: Request, res: Response) => {
  const { disasterType, occurrenceTime, location, affectedPopulation, evacuatedPopulation, deathMissingCount, cropAreaAffected, housesDamaged, directEconomicLoss } = req.body;

  if (!disasterType || !occurrenceTime) {
    return res.status(400).json({
      error: '灾害类型和发生时间为必填项'
    });
  }

  try {
    const result = createReport({
      disasterType,
      occurrenceTime,
      location,
      affectedPopulation,
      evacuatedPopulation,
      deathMissingCount,
      cropAreaAffected,
      housesDamaged,
      directEconomicLoss
    });

    return res.status(201).json({
      report: result.report,
      merged: result.merged,
      message: result.merged ? '相似灾情已自动合并' : '灾情上报成功'
    });
  } catch (error) {
    console.error('上报失败:', error);
    return res.status(500).json({ error: '服务器内部错误' });
  }
});

router.get('/', (req: Request, res: Response) => {
  const { status } = req.query;
  
  try {
    const reports = getAllReports(status as string | undefined);
    return res.json(reports);
  } catch (error) {
    console.error('查询失败:', error);
    return res.status(500).json({ error: '服务器内部错误' });
  }
});

router.get('/:id', (req: Request, res: Response) => {
  const { id } = req.params;

  try {
    const report = getReportById(id);
    
    if (!report) {
      return res.status(404).json({ error: '灾情记录不存在' });
    }

    return res.json(report);
  } catch (error) {
    console.error('查询失败:', error);
    return res.status(500).json({ error: '服务器内部错误' });
  }
});

router.put('/:id', (req: Request, res: Response) => {
  const { id } = req.params;
  const { affectedPopulation, evacuatedPopulation, deathMissingCount, cropAreaAffected, housesDamaged, directEconomicLoss } = req.body;

  try {
    const report = getReportById(id);
    
    if (!report) {
      return res.status(404).json({ error: '灾情记录不存在' });
    }

    const isVerified = report.status === ReportStatus.VERIFIED || report.status === ReportStatus.PUBLISHED;
    
    if (isVerified) {
      return res.status(405).json({
        error: '已核查的数据不可随意修改，请提交修改申请'
      });
    }

    const result = updateReport(id, {
      affectedPopulation,
      evacuatedPopulation,
      deathMissingCount,
      cropAreaAffected,
      housesDamaged,
      directEconomicLoss
    }, false);

    if (result.success) {
      const updatedReport = getReportById(id);
      return res.json(updatedReport);
    }

    return res.status(500).json({ error: '更新失败' });
  } catch (error) {
    console.error('更新失败:', error);
    return res.status(500).json({ error: '服务器内部错误' });
  }
});

export default router;
