import { v4 as uuidv4 } from 'uuid';
import { db } from '../db';
import { Opportunity, CreateOpportunityRequest, UpdateStageRequest, Stage, StageHistory } from '../types';
import { isValidStageTransition } from '../constants';
import { salespersonExists } from './salespersonService';

interface DbOpportunity {
  id: string;
  name: string;
  customer_id: string;
  customer_name: string;
  amount: number;
  salesperson_id: string;
  current_stage: string;
  created_at: string;
  updated_at: string;
  closed_at: string | null;
  loss_reason: string | null;
}

interface DbStageHistory {
  id: string;
  opportunity_id: string;
  from_stage: string | null;
  to_stage: string;
  reason: string | null;
  timestamp: string;
}

function mapToOpportunity(dbRow: DbOpportunity): Opportunity {
  return {
    id: dbRow.id,
    name: dbRow.name,
    customerId: dbRow.customer_id,
    customerName: dbRow.customer_name,
    amount: dbRow.amount,
    salespersonId: dbRow.salesperson_id,
    currentStage: dbRow.current_stage as Stage,
    createdAt: dbRow.created_at,
    updatedAt: dbRow.updated_at,
    closedAt: dbRow.closed_at,
    lossReason: dbRow.loss_reason,
  };
}

function mapToStageHistory(dbRow: DbStageHistory): StageHistory {
  return {
    id: dbRow.id,
    opportunityId: dbRow.opportunity_id,
    fromStage: dbRow.from_stage as Stage | null,
    toStage: dbRow.to_stage as Stage,
    reason: dbRow.reason,
    timestamp: dbRow.timestamp,
  };
}

export interface StageUpdateResult {
  success: boolean;
  opportunity?: Opportunity;
  error?: {
    status: number;
    message: string;
  };
}

export function getOpportunities(): Opportunity[] {
  const rows = db.prepare('SELECT * FROM opportunities ORDER BY created_at DESC').all() as DbOpportunity[];
  return rows.map(mapToOpportunity);
}

export function getOpportunityById(id: string): Opportunity | null {
  const row = db.prepare('SELECT * FROM opportunities WHERE id = ?').get(id) as DbOpportunity | undefined;
  return row ? mapToOpportunity(row) : null;
}

export function getOpportunityHistory(opportunityId: string): StageHistory[] {
  const rows = db.prepare(
    'SELECT * FROM stage_history WHERE opportunity_id = ? ORDER BY timestamp ASC'
  ).all(opportunityId) as DbStageHistory[];
  return rows.map(mapToStageHistory);
}

export interface CreateOpportunityResult {
  success: boolean;
  opportunity?: Opportunity;
  error?: {
    status: number;
    message: string;
  };
}

export function createOpportunity(data: CreateOpportunityRequest): CreateOpportunityResult {
  if (data.amount < 0) {
    return {
      success: false,
      error: { status: 400, message: '金额不能为负数' },
    };
  }

  if (!salespersonExists(data.salespersonId)) {
    return {
      success: false,
      error: { status: 404, message: '负责销售不存在' },
    };
  }

  const id = uuidv4();
  const historyId = uuidv4();
  const initialStage: Stage = '初次接触';

  const tx = db.transaction(() => {
    db.prepare(`
      INSERT INTO opportunities (
        id, name, customer_id, customer_name, amount, salesperson_id,
        current_stage, updated_at
      ) VALUES (?, ?, ?, ?, ?, ?, ?, datetime('now'))
    `).run(
      id,
      data.name,
      data.customerId,
      data.customerName,
      data.amount,
      data.salespersonId,
      initialStage
    );

    db.prepare(`
      INSERT INTO stage_history (id, opportunity_id, from_stage, to_stage, reason)
      VALUES (?, ?, NULL, ?, NULL)
    `).run(historyId, id, initialStage);
  });

  try {
    tx();
  } catch (e) {
    return {
      success: false,
      error: { status: 500, message: '创建商机失败' },
    };
  }

  const created = getOpportunityById(id);
  if (!created) {
    return {
      success: false,
      error: { status: 500, message: '创建商机失败' },
    };
  }

  return { success: true, opportunity: created };
}

export function updateStage(opportunityId: string, data: UpdateStageRequest): StageUpdateResult {
  const opportunity = getOpportunityById(opportunityId);
  if (!opportunity) {
    return {
      success: false,
      error: { status: 404, message: '商机不存在' },
    };
  }

  if (opportunity.currentStage === '赢单' || opportunity.currentStage === '输单') {
    return {
      success: false,
      error: { status: 400, message: '已结束的商机不能变更阶段' },
    };
  }

  if (opportunity.currentStage === data.newStage) {
    return {
      success: false,
      error: { status: 400, message: '目标阶段与当前阶段相同' },
    };
  }

  const validation = isValidStageTransition(opportunity.currentStage, data.newStage);
  if (!validation.valid) {
    return {
      success: false,
      error: { status: 400, message: validation.message || '无效的阶段转换' },
    };
  }

  if (data.newStage === '输单' && !data.reason) {
    return {
      success: false,
      error: { status: 400, message: '输单必须提供原因' },
    };
  }

  const historyId = uuidv4();
  const isClosing = data.newStage === '赢单' || data.newStage === '输单';

  const tx = db.transaction(() => {
    if (isClosing) {
      db.prepare(`
        UPDATE opportunities
        SET current_stage = ?, updated_at = datetime('now'),
            closed_at = datetime('now'), loss_reason = ?
        WHERE id = ?
      `).run(
        data.newStage,
        data.newStage === '输单' ? data.reason! : null,
        opportunityId
      );
    } else {
      db.prepare(`
        UPDATE opportunities
        SET current_stage = ?, updated_at = datetime('now'),
            closed_at = NULL, loss_reason = NULL
        WHERE id = ?
      `).run(data.newStage, opportunityId);
    }

    db.prepare(`
      INSERT INTO stage_history (id, opportunity_id, from_stage, to_stage, reason)
      VALUES (?, ?, ?, ?, ?)
    `).run(historyId, opportunityId, opportunity.currentStage, data.newStage, data.reason || null);
  });

  try {
    tx();
  } catch (e) {
    return {
      success: false,
      error: { status: 500, message: '阶段更新失败，事务已回滚' },
    };
  }

  const updated = getOpportunityById(opportunityId);
  if (!updated) {
    return {
      success: false,
      error: { status: 500, message: '阶段更新失败' },
    };
  }

  return { success: true, opportunity: updated };
}

export function deleteOpportunity(id: string): boolean {
  const stmt = db.prepare('DELETE FROM opportunities WHERE id = ?');
  const result = stmt.run(id);
  return (result.changes ?? 0) > 0;
}

export function updateOpportunityDetails(
  id: string,
  data: Partial<Omit<CreateOpportunityRequest, 'salespersonId'>> & { salespersonId?: string }
): {
  success: boolean;
  opportunity?: Opportunity;
  error?: { status: number; message: string };
} {
  const existing = getOpportunityById(id);
  if (!existing) {
    return { success: false, error: { status: 404, message: '商机不存在' } };
  }

  if (data.amount !== undefined && data.amount < 0) {
    return { success: false, error: { status: 400, message: '金额不能为负数' } };
  }

  if (data.salespersonId !== undefined && !salespersonExists(data.salespersonId)) {
    return { success: false, error: { status: 404, message: '负责销售不存在' } };
  }

  const updates: string[] = [];
  const values: any[] = [];

  if (data.name !== undefined) { updates.push('name = ?'); values.push(data.name); }
  if (data.customerId !== undefined) { updates.push('customer_id = ?'); values.push(data.customerId); }
  if (data.customerName !== undefined) { updates.push('customer_name = ?'); values.push(data.customerName); }
  if (data.amount !== undefined) { updates.push('amount = ?'); values.push(data.amount); }
  if (data.salespersonId !== undefined) { updates.push('salesperson_id = ?'); values.push(data.salespersonId); }

  if (updates.length === 0) {
    return { success: true, opportunity: existing };
  }

  updates.push('updated_at = datetime(\'now\')');
  values.push(id);

  db.prepare(`UPDATE opportunities SET ${updates.join(', ')} WHERE id = ?`).run(...values);

  const updated = getOpportunityById(id);
  if (!updated) {
    return { success: false, error: { status: 500, message: '更新失败' } };
  }

  return { success: true, opportunity: updated };
}
