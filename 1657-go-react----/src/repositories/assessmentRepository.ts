import db from '../db';
import { RiskAssessment } from '../types';
import { v4 as uuidv4 } from 'uuid';

export class AssessmentRepository {
  private static instance: AssessmentRepository;

  static getInstance(): AssessmentRepository {
    if (!AssessmentRepository.instance) {
      AssessmentRepository.instance = new AssessmentRepository();
    }
    return AssessmentRepository.instance;
  }

  create(assessment: Omit<RiskAssessment, 'id' | 'assessedAt'>): RiskAssessment {
    const now = Date.now();
    const id = uuidv4();
    const stmt = db.prepare(`
      INSERT INTO assessments (
        id, transactionId, score, status, matchedRuleIds,
        assessedAt, accountFrozen, frozenAttempted, freezeFailed
      ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
    `);
    stmt.run(
      id,
      assessment.transactionId,
      assessment.score,
      assessment.status,
      JSON.stringify(assessment.matchedRuleIds),
      now,
      assessment.accountFrozen ? 1 : 0,
      assessment.frozenAttempted ? 1 : 0,
      assessment.freezeFailed ? 1 : 0
    );
    return this.findById(id)!;
  }

  findById(id: string): RiskAssessment | null {
    const row = db.prepare('SELECT * FROM assessments WHERE id = ?').get(id);
    return row ? this.mapRow(row) : null;
  }

  findByTransactionId(transactionId: string): RiskAssessment | null {
    const row = db.prepare('SELECT * FROM assessments WHERE transactionId = ?').get(transactionId);
    return row ? this.mapRow(row) : null;
  }

  updateFreezeStatus(
    id: string,
    updates: { accountFrozen?: boolean; frozenAttempted?: boolean; freezeFailed?: boolean }
  ): RiskAssessment | null {
    const existing = this.findById(id);
    if (!existing) return null;

    const fields: string[] = [];
    const values: any[] = [];

    for (const [key, value] of Object.entries(updates)) {
      if (value !== undefined) {
        fields.push(`${key} = ?`);
        values.push(value ? 1 : 0);
      }
    }

    values.push(id);
    db.prepare(`UPDATE assessments SET ${fields.join(', ')} WHERE id = ?`).run(...values);
    return this.findById(id);
  }

  findFailedFreezes(): RiskAssessment[] {
    const rows = db.prepare(
      'SELECT * FROM assessments WHERE frozenAttempted = 1 AND freezeFailed = 1 AND accountFrozen = 0'
    ).all();
    return rows.map(row => this.mapRow(row));
  }

  private mapRow(row: any): RiskAssessment {
    return {
      id: row.id,
      transactionId: row.transactionId,
      score: row.score,
      status: row.status,
      matchedRuleIds: JSON.parse(row.matchedRuleIds),
      assessedAt: row.assessedAt,
      accountFrozen: !!row.accountFrozen,
      frozenAttempted: !!row.frozenAttempted,
      freezeFailed: !!row.freezeFailed
    };
  }
}
