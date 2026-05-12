import { Request, Response, NextFunction } from 'express';
import { riskService } from '../services/riskService';

export function riskAssessor(req: Request, res: Response, next: NextFunction): void {
  if (!req.realIP) {
    res.status(400).json({ error: '无法解析客户端IP' });
    return;
  }

  const assessment = riskService.assessRisk(req.realIP, req.path);
  
  req.riskScore = assessment.score;
  req.riskFactors = assessment.factors;

  if (!assessment.allowed) {
    res.status(403).json({ 
      error: '风险分数过高，拒绝访问',
      riskScore: assessment.score,
      threshold: riskService.getThreshold(),
      factors: assessment.factors
    });
    return;
  }

  next();
}
