import { Router, Request, Response } from 'express';
import { db, getScaleById } from '../database';
import { v4 as uuidv4 } from 'uuid';
import { Assessment, Answer, AssessmentResult, Scale } from '../types';
import { 
  isAssessmentExpired, 
  isWithinRepeatPeriod, 
  calculateRawScore, 
  getNormEntry, 
  generateReport 
} from '../services/scoringService';
import { createAlert, getRelatedUsers } from '../services/alertService';
import { sendAlertNotification } from '../services/notificationService';

const router = Router();

const getAssessmentById = (id: string): Promise<Assessment | null> => {
  return new Promise((resolve, reject) => {
    db.get('SELECT * FROM assessments WHERE id = ?', [id], (err, row: any) => {
      if (err) reject(err);
      else if (!row) resolve(null);
      else resolve({
        id: row.id,
        userId: row.user_id,
        scaleId: row.scale_id,
        startTime: row.start_time,
        status: row.status,
        answers: JSON.parse(row.answers || '[]'),
        lastActiveTime: row.last_active_time
      });
    });
  });
};

const getLastCompletedResult = (userId: string, scaleId: string): Promise<AssessmentResult | null> => {
  return new Promise((resolve, reject) => {
    db.get(
      `SELECT * FROM assessment_results 
       WHERE user_id = ? AND scale_id = ? 
       ORDER BY completed_at DESC 
       LIMIT 1`,
      [userId, scaleId],
      (err, row: any) => {
        if (err) reject(err);
        else if (!row) resolve(null);
        else resolve({
          id: row.id,
          assessmentId: row.assessment_id,
          userId: row.user_id,
          scaleId: row.scale_id,
          rawScore: row.raw_score,
          standardScore: row.standard_score,
          level: row.level,
          dimensionScores: row.dimension_scores ? JSON.parse(row.dimension_scores) : undefined,
          completedAt: row.completed_at,
          report: row.report
        });
      }
    );
  });
};

const getActiveAssessment = (userId: string, scaleId: string): Promise<Assessment | null> => {
  return new Promise((resolve, reject) => {
    db.get(
      `SELECT * FROM assessments 
       WHERE user_id = ? AND scale_id = ? AND status IN ('in_progress', 'paused')
       ORDER BY start_time DESC 
       LIMIT 1`,
      [userId, scaleId],
      (err, row: any) => {
        if (err) reject(err);
        else if (!row) resolve(null);
        else resolve({
          id: row.id,
          userId: row.user_id,
          scaleId: row.scale_id,
          startTime: row.start_time,
          status: row.status,
          answers: JSON.parse(row.answers || '[]'),
          lastActiveTime: row.last_active_time
        });
      }
    );
  });
};

const updateAssessment = (assessment: Assessment): Promise<void> => {
  return new Promise((resolve, reject) => {
    db.run(
      `UPDATE assessments SET status = ?, answers = ?, last_active_time = ? WHERE id = ?`,
      [
        assessment.status,
        JSON.stringify(assessment.answers),
        assessment.lastActiveTime,
        assessment.id
      ],
      (err) => {
        if (err) reject(err);
        else resolve();
      }
    );
  });
};

router.post('/start', async (req: Request, res: Response) => {
  try {
    const { userId, scaleId } = req.body;

    if (!userId || !scaleId) {
      res.status(400).json({ error: 'userId and scaleId are required' });
      return;
    }

    const scale = await getScaleById(scaleId);
    if (!scale) {
      res.status(404).json({ error: 'Scale not found' });
      return;
    }

    const activeAssessment = await getActiveAssessment(userId, scaleId);
    if (activeAssessment && !isAssessmentExpired(activeAssessment.lastActiveTime)) {
      res.status(409).json({ 
        error: 'Cannot retake the same scale within 30 days',
        activeAssessmentId: activeAssessment.id 
      });
      return;
    }

    const lastResult = await getLastCompletedResult(userId, scaleId);
    if (lastResult && isWithinRepeatPeriod(lastResult.completedAt)) {
      res.status(409).json({ error: 'Cannot retake the same scale within 30 days' });
      return;
    }

    const now = Date.now();
    const assessment: Assessment = {
      id: uuidv4(),
      userId,
      scaleId,
      startTime: now,
      status: 'in_progress',
      answers: [],
      lastActiveTime: now
    };

    db.run(
      `INSERT INTO assessments (id, user_id, scale_id, start_time, status, answers, last_active_time)
       VALUES (?, ?, ?, ?, ?, ?, ?)`,
      [
        assessment.id,
        assessment.userId,
        assessment.scaleId,
        assessment.startTime,
        assessment.status,
        JSON.stringify(assessment.answers),
        assessment.lastActiveTime
      ],
      (err) => {
        if (err) {
          res.status(500).json({ error: err.message });
          return;
        }
        res.status(201).json(assessment);
      }
    );
  } catch (err: any) {
    res.status(500).json({ error: err.message });
  }
});

router.post('/:id/answer', async (req: Request, res: Response) => {
  try {
    const assessmentId = req.params.id;
    const { questionId, optionId, answerTime } = req.body;

    if (!questionId || !optionId || answerTime === undefined) {
      res.status(400).json({ error: 'questionId, optionId, and answerTime are required' });
      return;
    }

    const assessment = await getAssessmentById(assessmentId);
    if (!assessment) {
      res.status(404).json({ error: 'Assessment not found' });
      return;
    }

    if (assessment.status === 'completed' || assessment.status === 'expired') {
      res.status(400).json({ error: 'Assessment is already completed or expired' });
      return;
    }

    if (isAssessmentExpired(assessment.lastActiveTime)) {
      assessment.status = 'expired';
      await updateAssessment(assessment);
      res.status(400).json({ error: '测评已超时' });
      return;
    }

    const scale = await getScaleById(assessment.scaleId);
    if (!scale) {
      res.status(404).json({ error: 'Scale not found' });
      return;
    }

    const question = scale.questions.find(q => q.id === questionId);
    if (!question) {
      res.status(404).json({ error: 'Question not found' });
      return;
    }

    const option = question.options.find(o => o.id === optionId);
    if (!option) {
      res.status(404).json({ error: 'Option not found' });
      return;
    }

    const existingAnswerIndex = assessment.answers.findIndex(a => a.questionId === questionId);
    const newAnswer: Answer = {
      questionId,
      optionId,
      score: option.score,
      answerTime
    };

    if (existingAnswerIndex >= 0) {
      assessment.answers[existingAnswerIndex] = newAnswer;
    } else {
      assessment.answers.push(newAnswer);
    }

    assessment.lastActiveTime = Date.now();
    await updateAssessment(assessment);

    res.json({ success: true, assessment });
  } catch (err: any) {
    res.status(500).json({ error: err.message });
  }
});

router.post('/:id/pause', async (req: Request, res: Response) => {
  try {
    const assessmentId = req.params.id;
    const assessment = await getAssessmentById(assessmentId);

    if (!assessment) {
      res.status(404).json({ error: 'Assessment not found' });
      return;
    }

    if (assessment.status !== 'in_progress') {
      res.status(400).json({ error: 'Only in-progress assessments can be paused' });
      return;
    }

    if (isAssessmentExpired(assessment.lastActiveTime)) {
      assessment.status = 'expired';
      await updateAssessment(assessment);
      res.status(400).json({ error: '测评已超时' });
      return;
    }

    assessment.status = 'paused';
    assessment.lastActiveTime = Date.now();
    await updateAssessment(assessment);

    res.json(assessment);
  } catch (err: any) {
    res.status(500).json({ error: err.message });
  }
});

router.post('/:id/resume', async (req: Request, res: Response) => {
  try {
    const assessmentId = req.params.id;
    const assessment = await getAssessmentById(assessmentId);

    if (!assessment) {
      res.status(404).json({ error: 'Assessment not found' });
      return;
    }

    if (assessment.status !== 'paused') {
      res.status(400).json({ error: 'Only paused assessments can be resumed' });
      return;
    }

    if (isAssessmentExpired(assessment.lastActiveTime)) {
      assessment.status = 'expired';
      await updateAssessment(assessment);
      res.status(400).json({ error: '测评已超时' });
      return;
    }

    assessment.status = 'in_progress';
    assessment.lastActiveTime = Date.now();
    await updateAssessment(assessment);

    res.json(assessment);
  } catch (err: any) {
    res.status(500).json({ error: err.message });
  }
});

router.post('/:id/complete', async (req: Request, res: Response) => {
  try {
    const assessmentId = req.params.id;
    const assessment = await getAssessmentById(assessmentId);

    if (!assessment) {
      res.status(404).json({ error: 'Assessment not found' });
      return;
    }

    if (assessment.status === 'completed' || assessment.status === 'expired') {
      res.status(400).json({ error: 'Assessment is already completed or expired' });
      return;
    }

    if (isAssessmentExpired(assessment.lastActiveTime)) {
      assessment.status = 'expired';
      await updateAssessment(assessment);
      res.status(400).json({ error: '测评已超时' });
      return;
    }

    const scale = await getScaleById(assessment.scaleId);
    if (!scale) {
      res.status(404).json({ error: 'Scale not found' });
      return;
    }

    const answeredQuestionIds = new Set(assessment.answers.map(a => a.questionId));
    const allQuestionIds = new Set(scale.questions.map(q => q.id));
    const missingQuestions = [...allQuestionIds].filter(id => !answeredQuestionIds.has(id));

    if (missingQuestions.length > 0) {
      res.status(400).json({ 
        error: 'Not all questions answered', 
        missingQuestions 
      });
      return;
    }

    const { rawScore, dimensionScores } = calculateRawScore(scale, assessment.answers);
    const normEntry = getNormEntry(scale.normTable, rawScore);
    const report = generateReport(scale, rawScore, normEntry.standardScore, normEntry.level, dimensionScores);

    const now = Date.now();
    const result: AssessmentResult = {
      id: uuidv4(),
      assessmentId: assessment.id,
      userId: assessment.userId,
      scaleId: assessment.scaleId,
      rawScore,
      standardScore: normEntry.standardScore,
      level: normEntry.level,
      dimensionScores: Object.keys(dimensionScores).length > 0 ? dimensionScores : undefined,
      completedAt: now,
      report
    };

    assessment.status = 'completed';
    assessment.lastActiveTime = now;
    await updateAssessment(assessment);

    db.run(
      `INSERT INTO assessment_results (id, assessment_id, user_id, scale_id, raw_score, standard_score, level, dimension_scores, completed_at, report)
       VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
      [
        result.id,
        result.assessmentId,
        result.userId,
        result.scaleId,
        result.rawScore,
        result.standardScore,
        result.level,
        result.dimensionScores ? JSON.stringify(result.dimensionScores) : null,
        result.completedAt,
        result.report
      ],
      async (err) => {
        if (err) {
          res.status(500).json({ error: err.message });
          return;
        }

        try {
          if (result.level === 'moderate' || result.level === 'severe') {
            const alert = await createAlert(result);
            if (alert) {
              const relatedUsers = await getRelatedUsers(result.userId);
              for (const recipientId of relatedUsers) {
                await sendAlertNotification(alert, result, scale, recipientId);
              }
            }
          }
        } catch (alertErr: any) {
          console.error('Alert creation failed:', alertErr.message);
        }

        res.status(201).json(result);
      }
    );
  } catch (err: any) {
    res.status(500).json({ error: err.message });
  }
});

router.get('/:id', async (req: Request, res: Response) => {
  try {
    const assessment = await getAssessmentById(req.params.id);
    if (!assessment) {
      res.status(404).json({ error: 'Assessment not found' });
      return;
    }
    res.json(assessment);
  } catch (err: any) {
    res.status(500).json({ error: err.message });
  }
});

router.get('/:id/result', async (req: Request, res: Response) => {
  try {
    const assessment = await getAssessmentById(req.params.id);
    if (!assessment) {
      res.status(404).json({ error: 'Assessment not found' });
      return;
    }

    if (assessment.status !== 'completed') {
      res.status(400).json({ error: 'Assessment not completed yet' });
      return;
    }

    const result = await new Promise<AssessmentResult | null>((resolve, reject) => {
      db.get(
        'SELECT * FROM assessment_results WHERE assessment_id = ?',
        [req.params.id],
        (err, row: any) => {
          if (err) reject(err);
          else if (!row) resolve(null);
          else resolve({
            id: row.id,
            assessmentId: row.assessment_id,
            userId: row.user_id,
            scaleId: row.scale_id,
            rawScore: row.raw_score,
            standardScore: row.standard_score,
            level: row.level,
            dimensionScores: row.dimension_scores ? JSON.parse(row.dimension_scores) : undefined,
            completedAt: row.completed_at,
            report: row.report
          });
        }
      );
    });

    if (!result) {
      res.status(404).json({ error: 'Result not found' });
      return;
    }

    res.json(result);
  } catch (err: any) {
    res.status(500).json({ error: err.message });
  }
});

export default router;
