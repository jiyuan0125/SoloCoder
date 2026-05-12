import db from '../db';
import { ReviewRecord } from '../types';
import { v4 as uuidv4 } from 'uuid';

export class ReviewRepository {
  private static instance: ReviewRepository;

  static getInstance(): ReviewRepository {
    if (!ReviewRepository.instance) {
      ReviewRepository.instance = new ReviewRepository();
    }
    return ReviewRepository.instance;
  }

  create(review: Omit<ReviewRecord, 'id' | 'createdAt'>): ReviewRecord {
    const now = Date.now();
    const id = uuidv4();
    const stmt = db.prepare(`
      INSERT INTO reviews (
        id, assessmentId, transactionId, reviewerId,
        decision, decidedAt, createdAt
      ) VALUES (?, ?, ?, ?, ?, ?, ?)
    `);
    stmt.run(
      id,
      review.assessmentId,
      review.transactionId,
      review.reviewerId || null,
      review.decision,
      review.decidedAt || null,
      now
    );
    return this.findById(id)!;
  }

  findById(id: string): ReviewRecord | null {
    const row = db.prepare('SELECT * FROM reviews WHERE id = ?').get(id);
    return row ? this.mapRow(row) : null;
  }

  findByAssessmentId(assessmentId: string): ReviewRecord | null {
    const row = db.prepare('SELECT * FROM reviews WHERE assessmentId = ?').get(assessmentId);
    return row ? this.mapRow(row) : null;
  }

  findPending(): ReviewRecord[] {
    const rows = db.prepare(`
      SELECT r.* FROM reviews r
      WHERE r.decision = 'PENDING'
      ORDER BY r.createdAt ASC
    `).all();
    return rows.map(row => this.mapRow(row));
  }

  decide(
    id: string,
    decision: 'NORMAL' | 'FRAUD',
    reviewerId?: string
  ): ReviewRecord | null {
    const existing = this.findById(id);
    if (!existing) return null;

    const stmt = db.prepare(`
      UPDATE reviews 
      SET decision = ?, reviewerId = ?, decidedAt = ?
      WHERE id = ?
    `);
    stmt.run(decision, reviewerId || null, Date.now(), id);
    return this.findById(id);
  }

  private mapRow(row: any): ReviewRecord {
    return {
      id: row.id,
      assessmentId: row.assessmentId,
      transactionId: row.transactionId,
      reviewerId: row.reviewerId,
      decision: row.decision,
      decidedAt: row.decidedAt,
      createdAt: row.createdAt
    };
  }
}
