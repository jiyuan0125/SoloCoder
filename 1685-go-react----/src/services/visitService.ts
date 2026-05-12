import { db } from '../database';
import {
  VisitPlan,
  VisitStatus,
  ServiceLevel,
  CompleteVisitInput,
} from '../types';
import { generateId, now, formatDate, addDays, getVisitFrequencyDays, getHigherLevel } from '../utils';
import { elderService } from './elderService';

function mapVisitPlan(row: any): VisitPlan {
  return {
    id: row.id,
    elderId: row.elder_id,
    scheduledDate: row.scheduled_date,
    status: row.status as VisitStatus,
    completedAt: row.completed_at,
    notes: row.notes,
    createdAt: row.created_at,
  };
}

export const visitService = {
  getVisitPlanById(id: string): VisitPlan | null {
    const row = db.prepare('SELECT * FROM visit_plans WHERE id = ?').get(id);
    return row ? mapVisitPlan(row) : null;
  },

  getVisitPlansByElderId(elderId: string): VisitPlan[] {
    const rows = db.prepare(
      'SELECT * FROM visit_plans WHERE elder_id = ? ORDER BY scheduled_date DESC'
    ).all(elderId);
    return rows.map(mapVisitPlan);
  },

  getPendingVisitPlans(elderId: string): VisitPlan[] {
    const rows = db.prepare(
      'SELECT * FROM visit_plans WHERE elder_id = ? AND status IN (?, ?) ORDER BY scheduled_date ASC'
    ).all(elderId, VisitStatus.SCHEDULED, VisitStatus.MISSED);
    return rows.map(mapVisitPlan);
  },

  initializeVisitPlan(elderId: string): void {
    const elder = elderService.getElderById(elderId);
    if (!elder) return;

    const effectiveLevel = elderService.getEffectiveServiceLevel(elder);
    const frequencyDays = getVisitFrequencyDays(effectiveLevel);
    const nextDate = formatDate(addDays(new Date(), frequencyDays));

    const id = generateId();
    const timestamp = now();

    db.prepare(`
      INSERT INTO visit_plans (
        id, elder_id, scheduled_date, status, completed_at, notes, created_at
      ) VALUES (?, ?, ?, ?, NULL, NULL, ?)
    `).run(id, elderId, nextDate, VisitStatus.SCHEDULED, timestamp);
  },

  createNextVisitPlan(elderId: string): void {
    const elder = elderService.getElderById(elderId);
    if (!elder) return;

    const effectiveLevel = elderService.getEffectiveServiceLevel(elder);
    const frequencyDays = getVisitFrequencyDays(effectiveLevel);

    const lastPlanned = db.prepare(
      'SELECT MAX(scheduled_date) as max_date FROM visit_plans WHERE elder_id = ?'
    ).get(elderId) as { max_date: string } | undefined;

    const startDate = lastPlanned?.max_date
      ? new Date(lastPlanned.max_date)
      : new Date();

    const nextDate = formatDate(addDays(startDate, frequencyDays));

    const existing = db.prepare(
      'SELECT id FROM visit_plans WHERE elder_id = ? AND scheduled_date = ?'
    ).get(elderId, nextDate);

    if (existing) return;

    const id = generateId();
    const timestamp = now();

    db.prepare(`
      INSERT INTO visit_plans (
        id, elder_id, scheduled_date, status, completed_at, notes, created_at
      ) VALUES (?, ?, ?, ?, NULL, NULL, ?)
    `).run(id, elderId, nextDate, VisitStatus.SCHEDULED, timestamp);
  },

  adjustVisitPlanOnLevelChange(
    elderId: string,
    newLevel: ServiceLevel,
    previousLevel: ServiceLevel,
    transitionUntil: string
  ): boolean {
    try {
      const transitionEndDate = new Date(transitionUntil);
      const today = new Date();

      const pendingPlans = db.prepare(
        'SELECT * FROM visit_plans WHERE elder_id = ? AND status = ? ORDER BY scheduled_date ASC'
      ).all(elderId, VisitStatus.SCHEDULED) as Array<{ id: string; scheduled_date: string }>;

      for (const plan of pendingPlans) {
        const scheduledDate = new Date(plan.scheduled_date);
        let effectiveLevel: ServiceLevel;

        if (scheduledDate < transitionEndDate) {
          effectiveLevel = getHigherLevel(previousLevel, newLevel);
        } else {
          effectiveLevel = newLevel;
        }

        const currentFrequency = getVisitFrequencyDays(ServiceLevel.LEVEL_1);
        const expectedFrequency = getVisitFrequencyDays(effectiveLevel);

        if (currentFrequency !== expectedFrequency) {
          db.prepare('DELETE FROM visit_plans WHERE id = ?').run(plan.id);
        }
      }

      visitService.rebuildVisitPlans(elderId, newLevel, previousLevel, transitionEndDate);
      return true;
    } catch (e) {
      return false;
    }
  },

  rebuildVisitPlans(
    elderId: string,
    newLevel: ServiceLevel,
    previousLevel: ServiceLevel,
    transitionEndDate: Date
  ): void {
    const today = new Date();
    let currentDate = today;

    for (let i = 0; i < 12; i++) {
      let effectiveLevel: ServiceLevel;
      if (currentDate < transitionEndDate) {
        effectiveLevel = getHigherLevel(previousLevel, newLevel);
      } else {
        effectiveLevel = newLevel;
      }

      const frequencyDays = getVisitFrequencyDays(effectiveLevel);
      const nextDate = formatDate(addDays(currentDate, frequencyDays));

      const existing = db.prepare(
        'SELECT id FROM visit_plans WHERE elder_id = ? AND scheduled_date = ?'
      ).get(elderId, nextDate);

      if (!existing) {
        const id = generateId();
        const timestamp = now();
        db.prepare(`
          INSERT INTO visit_plans (
            id, elder_id, scheduled_date, status, completed_at, notes, created_at
          ) VALUES (?, ?, ?, ?, NULL, NULL, ?)
        `).run(id, elderId, nextDate, VisitStatus.SCHEDULED, timestamp);
      }

      currentDate = new Date(nextDate);
    }
  },

  completeVisit(input: CompleteVisitInput): VisitPlan | null {
    const plan = visitService.getVisitPlanById(input.planId);
    if (!plan) return null;

    const completedAt = now();
    let newStatus: VisitStatus;

    if (input.visited) {
      newStatus = VisitStatus.COMPLETED;
    } else {
      const recentPlans = db.prepare(`
        SELECT * FROM visit_plans 
        WHERE elder_id = ? 
          AND status IN (?, ?)
          AND id != ?
        ORDER BY scheduled_date DESC 
        LIMIT 1
      `).all(plan.elderId, VisitStatus.MISSED, VisitStatus.REPORTED, input.planId);

      const hasPreviousMissed = recentPlans.some(
        (p: any) => p.status === VisitStatus.MISSED || p.status === VisitStatus.REPORTED
      );

      if (hasPreviousMissed && !input.contactedByPhone) {
        newStatus = VisitStatus.REPORTED;
      } else {
        newStatus = VisitStatus.MISSED;
      }
    }

    db.prepare(`
      UPDATE visit_plans 
      SET status = ?, completed_at = ?, notes = ?
      WHERE id = ?
    `).run(newStatus, completedAt, input.notes || null, input.planId);

    if (newStatus === VisitStatus.COMPLETED || newStatus === VisitStatus.MISSED) {
      visitService.createNextVisitPlan(plan.elderId);
    }

    return visitService.getVisitPlanById(input.planId);
  },

  getReportedVisits(): VisitPlan[] {
    const rows = db.prepare(
      'SELECT * FROM visit_plans WHERE status = ? ORDER BY scheduled_date DESC'
    ).all(VisitStatus.REPORTED);
    return rows.map(mapVisitPlan);
  },
};
