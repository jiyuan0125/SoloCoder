import { db } from '../database';
import { Elder, CreateElderInput, ServiceLevel } from '../types';
import { generateId, now, addMonths } from '../utils';
import { visitService } from './visitService';

function mapElder(row: any): Elder {
  return {
    id: row.id,
    name: row.name,
    idCard: row.id_card,
    phone: row.phone,
    address: row.address,
    emergencyContact: row.emergency_contact,
    emergencyContactPhone: row.emergency_contact_phone,
    serviceLevel: row.service_level as ServiceLevel,
    transitionUntil: row.transition_until,
    previousLevel: row.previous_level as ServiceLevel | null,
    visitPlanNeedsManualAdjustment: row.visit_plan_needs_manual_adjustment === 1,
    createdAt: row.created_at,
    updatedAt: row.updated_at,
  };
}

export const elderService = {
  createElder(input: CreateElderInput): Elder {
    const id = generateId();
    const timestamp = now();
    const stmt = db.prepare(`
      INSERT INTO elders (
        id, name, id_card, phone, address, emergency_contact, emergency_contact_phone,
        service_level, transition_until, previous_level, visit_plan_needs_manual_adjustment,
        created_at, updated_at
      ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, NULL, NULL, 0, ?, ?)
    `);
    stmt.run(
      id,
      input.name,
      input.idCard,
      input.phone,
      input.address,
      input.emergencyContact,
      input.emergencyContactPhone,
      ServiceLevel.LEVEL_1,
      timestamp,
      timestamp
    );
    const elder = elderService.getElderById(id)!;
    visitService.initializeVisitPlan(elder.id);
    return elder;
  },

  getElderById(id: string): Elder | null {
    const row = db.prepare('SELECT * FROM elders WHERE id = ?').get(id);
    return row ? mapElder(row) : null;
  },

  getAllElders(): Elder[] {
    const rows = db.prepare('SELECT * FROM elders ORDER BY created_at DESC').all();
    return rows.map(mapElder);
  },

  updateServiceLevelWithTransition(
    elderId: string,
    newLevel: ServiceLevel
  ): Elder | null {
    const elder = elderService.getElderById(elderId);
    if (!elder) return null;

    if (elder.serviceLevel === newLevel) {
      return elder;
    }

    const previousLevel = elder.serviceLevel;
    const transitionUntil = addMonths(new Date(), 1).toISOString();

    const tx = db.transaction(() => {
      const timestamp = now();
      db.prepare(`
        UPDATE elders 
        SET service_level = ?, 
            previous_level = ?, 
            transition_until = ?, 
            visit_plan_needs_manual_adjustment = 0,
            updated_at = ?
        WHERE id = ?
      `).run(newLevel, previousLevel, transitionUntil, timestamp, elderId);

      const success = visitService.adjustVisitPlanOnLevelChange(elderId, newLevel, previousLevel, transitionUntil);
      if (!success) {
        db.prepare(`
          UPDATE elders 
          SET visit_plan_needs_manual_adjustment = 1,
              updated_at = ?
          WHERE id = ?
        `).run(timestamp, elderId);
      }
    });

    try {
      tx();
    } catch (e) {
      const timestamp = now();
      db.prepare(`
        UPDATE elders 
        SET visit_plan_needs_manual_adjustment = 1,
            updated_at = ?
        WHERE id = ?
      `).run(timestamp, elderId);
    }

    return elderService.getElderById(elderId);
  },

  getEffectiveServiceLevel(elder: Elder): ServiceLevel {
    if (!elder.transitionUntil) {
      return elder.serviceLevel;
    }
    const nowDate = new Date();
    const transitionEnd = new Date(elder.transitionUntil);
    if (nowDate < transitionEnd && elder.previousLevel) {
      const levelOrder: Record<ServiceLevel, number> = {
        [ServiceLevel.LEVEL_1]: 1,
        [ServiceLevel.LEVEL_2]: 2,
        [ServiceLevel.LEVEL_3]: 3,
      };
      return levelOrder[elder.previousLevel] > levelOrder[elder.serviceLevel]
        ? elder.previousLevel
        : elder.serviceLevel;
    }
    return elder.serviceLevel;
  },
};
