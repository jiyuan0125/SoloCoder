import { Request, Response } from 'express';
import { assessmentService, ScoreOutOfRangeError } from '../services/assessmentService';
import { CreateAssessmentInput } from '../types';

export const assessmentController = {
  createAssessment: (req: Request, res: Response) => {
    try {
      const input: CreateAssessmentInput = req.body;
      const assessment = assessmentService.createAssessment(input);
      if (!assessment) {
        res.status(404).json({ error: '老人不存在' });
        return;
      }
      res.status(201).json(assessment);
    } catch (error: any) {
      if (error instanceof ScoreOutOfRangeError) {
        res.status(400).json({ error: error.message });
      } else {
        res.status(400).json({ error: error.message });
      }
    }
  },

  getAssessment: (req: Request, res: Response) => {
    const assessment = assessmentService.getAssessmentById(req.params.id);
    if (!assessment) {
      res.status(404).json({ error: '评估记录不存在' });
      return;
    }
    res.json(assessment);
  },

  getAssessmentsByElder: (req: Request, res: Response) => {
    const assessments = assessmentService.getAssessmentsByElderId(req.params.elderId);
    res.json(assessments);
  },

  getLatestAssessment: (req: Request, res: Response) => {
    const assessment = assessmentService.getLatestAssessment(req.params.elderId);
    if (!assessment) {
      res.status(404).json({ error: '未找到评估记录' });
      return;
    }
    res.json(assessment);
  },
};
