import { db } from '../database';
import { Assessment, CreateAssessmentInput, ServiceLevel } from '../types';
import { generateId, now, calculateServiceLevel } from '../utils';
import { elderService } from './elderService';

function mapAssessment(row: any): Assessment {
  return {
    id: row.id,
    elderId: row.elder_id,
    selfCareScore: row.self_care_score,
    cognitiveScore: row.cognitive_score,
    totalScore: row.total_score,
    serviceLevel: row.service_level as ServiceLevel,
    assessmentDate: row.assessment_date,
    createdAt: row.created_at,
  };
}

export class ScoreOutOfRangeError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'ScoreOutOfRangeError';
  }
}

export const assessmentService = {
  validateScores(selfCareScore: number, cognitiveScore: number): void {
    if (selfCareScore < 0 || selfCareScore > 100) {
      throw new ScoreOutOfRangeError('自理能力评分必须在0-100之间');
    }
    if (cognitiveScore < 0 || cognitiveScore > 100) {
      throw new ScoreOutOfRangeError('认知能力评分必须在0-100之间');
    }
  },

  createAssessment(input: CreateAssessmentInput): Assessment | null {
    const elder = elderService.getElderById(input.elderId);
    if (!elder) return null;

    assessmentService.validateScores(input.selfCareScore, input.cognitiveScore);

    const totalScore = input.selfCareScore + input.cognitiveScore;
    const serviceLevel = calculateServiceLevel(totalScore);
    const id = generateId();
    const timestamp = now();
    const assessmentDate = input.assessmentDate || timestamp;

    const stmt = db.prepare(`
      INSERT INTO assessments (
        id, elder_id, self_care_score, cognitive_score, total_score, 
        service_level, assessment_date, created_at
      ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `);
    stmt.run(
      id,
      input.elderId,
      input.selfCareScore,
      input.cognitiveScore,
      totalScore,
      serviceLevel,
      assessmentDate,
      timestamp
    );

    if (serviceLevel !== elder.serviceLevel) {
      elderService.updateServiceLevelWithTransition(input.elderId, serviceLevel);
    }

    return assessmentService.getAssessmentById(id);
  },

  getAssessmentById(id: string): Assessment | null {
    const row = db.prepare('SELECT * FROM assessments WHERE id = ?').get(id);
    return row ? mapAssessment(row) : null;
  },

  getAssessmentsByElderId(elderId: string): Assessment[] {
    const rows = db.prepare(
      'SELECT * FROM assessments WHERE elder_id = ? ORDER BY assessment_date DESC'
    ).all(elderId);
    return rows.map(mapAssessment);
  },

  getLatestAssessment(elderId: string): Assessment | null {
    const row = db.prepare(
      'SELECT * FROM assessments WHERE elder_id = ? ORDER BY assessment_date DESC LIMIT 1'
    ).get(elderId);
    return row ? mapAssessment(row) : null;
  },
};
