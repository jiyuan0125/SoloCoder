import { v4 as uuidv4 } from 'uuid';
import { db, MEASURE_TYPES, EVALUATION_RESULTS } from '../database';
import type { AssistancePlan, AssistanceMeasure, EvaluationRecord } from '../types';
import { getHouseholdById, markReturnMonitoring } from './householdService';

function parsePlan(row: any): AssistancePlan {
  return {
    id: row.id,
    householdId: row.household_id,
    measures: JSON.parse(row.measures),
    adjustmentReason: row.adjustment_reason,
    approvedByTownship: row.approved_by_township === 1,
    approvalDate: row.approval_date,
    status: row.status,
    createdBy: row.created_by,
    createdAt: row.created_at,
    updatedAt: row.updated_at
  };
}

function addPlanHistory(
  planId: string,
  householdId: string,
  eventType: 'adjustment' | 'approval' | 'restart' | 'suspend' | 'complete',
  operator: string,
  reason?: string,
  previousStatus?: string,
  newStatus?: string
): void {
  const stmt = db.prepare(`
    INSERT INTO plan_histories (id, plan_id, household_id, event_type, previous_status, new_status, reason, operator, operation_date)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
  `);
  stmt.run(
    uuidv4(),
    planId,
    householdId,
    eventType,
    previousStatus || null,
    newStatus || null,
    reason || null,
    operator,
    new Date().toISOString()
  );
}

export function createPlan(householdId: string, data: any, operator: string): AssistancePlan {
  const household = getHouseholdById(householdId);
  if (!household) {
    throw new Error('HOUSEHOLD_NOT_FOUND');
  }

  const existing = getPlanByHousehold(householdId);
  if (existing) {
    throw new Error('PLAN_ALREADY_EXISTS');
  }

  if (data.measures && data.measures.length > 0) {
    for (const measure of data.measures) {
      if (!MEASURE_TYPES.includes(measure.type)) {
        throw new Error('INVALID_MEASURE_TYPE');
      }
    }
  }

  const measures: AssistanceMeasure[] = (data.measures || []).map((m: any) => ({
    id: uuidv4(),
    type: m.type,
    name: m.name,
    description: m.description,
    target: m.target,
    startDate: m.startDate,
    endDate: m.endDate,
    evaluations: [],
    needRedraft: false
  }));

  const now = new Date().toISOString();
  const id = uuidv4();

  db.prepare(`
    INSERT INTO assistance_plans (
      id, household_id, measures, status, created_by, created_at, updated_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?)
  `).run(
    id,
    householdId,
    JSON.stringify(measures),
    'executing',
    operator,
    now,
    now
  );

  addPlanHistory(id, householdId, 'adjustment', operator, '创建帮扶计划', undefined, 'executing');

  return getPlanByHousehold(householdId)!;
}

export function getPlanByHousehold(householdId: string): AssistancePlan | null {
  const row = db.prepare('SELECT * FROM assistance_plans WHERE household_id = ?').get(householdId);
  return row ? parsePlan(row) : null;
}

export function getPlanById(planId: string): AssistancePlan | null {
  const row = db.prepare('SELECT * FROM assistance_plans WHERE id = ?').get(planId);
  return row ? parsePlan(row) : null;
}

export function adjustPlan(
  householdId: string,
  measures: any[],
  reason: string,
  operator: string
): AssistancePlan {
  const plan = getPlanByHousehold(householdId);
  if (!plan) {
    throw new Error('PLAN_NOT_FOUND');
  }

  for (const measure of measures) {
    if (!MEASURE_TYPES.includes(measure.type)) {
      throw new Error('INVALID_MEASURE_TYPE');
    }
  }

  const previousStatus = plan.status;

  const newMeasures: AssistanceMeasure[] = measures.map((m: any) => ({
    id: m.id || uuidv4(),
    type: m.type,
    name: m.name,
    description: m.description,
    target: m.target,
    startDate: m.startDate,
    endDate: m.endDate,
    evaluations: m.evaluations || [],
    needRedraft: m.needRedraft || false
  }));

  const now = new Date().toISOString();

  db.prepare(`
    UPDATE assistance_plans SET
      measures = ?,
      adjustment_reason = ?,
      approved_by_township = 0,
      status = 'revised',
      updated_at = ?
    WHERE household_id = ?
  `).run(
    JSON.stringify(newMeasures),
    reason,
    now,
    householdId
  );

  addPlanHistory(
    plan.id,
    householdId,
    'adjustment',
    operator,
    reason,
    previousStatus,
    'revised'
  );

  return getPlanByHousehold(householdId)!;
}

export function approvePlan(householdId: string, operator: string): AssistancePlan {
  const plan = getPlanByHousehold(householdId);
  if (!plan) {
    throw new Error('PLAN_NOT_FOUND');
  }

  const previousStatus = plan.status;
  const now = new Date().toISOString();

  db.prepare(`
    UPDATE assistance_plans SET
      approved_by_township = 1,
      approval_date = ?,
      status = 'executing',
      updated_at = ?
    WHERE household_id = ?
  `).run(now, now, householdId);

  addPlanHistory(
    plan.id,
    householdId,
    'approval',
    operator,
    '乡镇批准调整后的帮扶计划',
    previousStatus,
    'executing'
  );

  return getPlanByHousehold(householdId)!;
}

export function addEvaluation(
  planId: string,
  measureId: string,
  data: any,
  operator: string
): AssistancePlan {
  const plan = getPlanById(planId);
  if (!plan) {
    throw new Error('PLAN_NOT_FOUND');
  }

  if (!EVALUATION_RESULTS.includes(data.result)) {
    throw new Error('INVALID_EVALUATION_RESULT');
  }

  const measureIndex = plan.measures.findIndex((m: AssistanceMeasure) => m.id === measureId);
  if (measureIndex === -1) {
    throw new Error('MEASURE_NOT_FOUND');
  }

  const measure = plan.measures[measureIndex];
  const previousEvaluations = measure.evaluations || [];
  const lastEval = previousEvaluations[previousEvaluations.length - 1];
  const prevCount = lastEval?.consecutiveIneffectiveCount || 0;
  const newCount = data.result === 'ineffective' ? prevCount + 1 : 0;

  const evaluation: EvaluationRecord = {
    id: uuidv4(),
    measureId,
    quarter: data.quarter,
    result: data.result,
    evaluator: operator,
    evaluationDate: new Date().toISOString(),
    notes: data.notes || '',
    consecutiveIneffectiveCount: newCount
  };

  const updatedMeasure = {
    ...measure,
    evaluations: [...measure.evaluations, evaluation],
    needRedraft: newCount >= 2
  };

  const updatedMeasures = [...plan.measures];
  updatedMeasures[measureIndex] = updatedMeasure;

  const now = new Date().toISOString();

  db.prepare(`
    UPDATE assistance_plans SET
      measures = ?,
      updated_at = ?
    WHERE id = ?
  `).run(JSON.stringify(updatedMeasures), now, planId);

  return getPlanById(planId)!;
}

export function getMeasureForRedraft(
  planId: string,
  measureId: string
): { originalMeasure: AssistanceMeasure; historyEvaluations: EvaluationRecord[] } | null {
  const plan = getPlanById(planId);
  if (!plan) {
    throw new Error('PLAN_NOT_FOUND');
  }

  const measure = plan.measures.find((m: AssistanceMeasure) => m.id === measureId);
  if (!measure) {
    throw new Error('EVALUATION_RECORD_NOT_FOUND');
  }

  if (!measure.needRedraft) {
    throw new Error('NO_REDRAFT_NEEDED');
  }

  return {
    originalMeasure: measure,
    historyEvaluations: measure.evaluations
  };
}

export function redraftMeasure(
  planId: string,
  measureId: string,
  newMeasureData: any,
  operator: string
): AssistancePlan {
  const plan = getPlanById(planId);
  if (!plan) {
    throw new Error('PLAN_NOT_FOUND');
  }

  const measureIndex = plan.measures.findIndex((m: AssistanceMeasure) => m.id === measureId);
  if (measureIndex === -1) {
    throw new Error('MEASURE_NOT_FOUND');
  }

  const existingMeasure = plan.measures[measureIndex];

  const newMeasure: AssistanceMeasure = {
    id: uuidv4(),
    type: newMeasureData.type,
    name: newMeasureData.name,
    description: newMeasureData.description,
    target: newMeasureData.target,
    startDate: newMeasureData.startDate,
    endDate: newMeasureData.endDate,
    evaluations: [],
    needRedraft: false
  };

  const updatedMeasures = [...plan.measures];
  updatedMeasures[measureIndex] = newMeasure;

  const now = new Date().toISOString();

  db.prepare(`
    UPDATE assistance_plans SET
      measures = ?,
      updated_at = ?
    WHERE id = ?
  `).run(JSON.stringify(updatedMeasures), now, planId);

  return getPlanById(planId)!;
}

export function handleReturnToPoverty(
  householdId: string,
  operator: string
): { household: any; plan: AssistancePlan | null } {
  const household = getHouseholdById(householdId);
  if (!household) {
    throw new Error('HOUSEHOLD_NOT_FOUND');
  }

  if (household.status !== 'out_of_poverty') {
    throw new Error('NOT_IN_MONITORING_PERIOD');
  }

  const plan = getPlanByHousehold(householdId);

  db.transaction(() => {
    markReturnMonitoring(householdId, operator);

    if (plan) {
      const previousStatus = plan.status;
      const now = new Date().toISOString();

      db.prepare(`
        UPDATE assistance_plans SET
          status = 'executing',
          updated_at = ?
        WHERE household_id = ?
      `).run(now, householdId);

      addPlanHistory(
        plan.id,
        householdId,
        'restart',
        operator,
        '返贫监测，恢复帮扶计划执行',
        previousStatus,
        'executing'
      );
    }
  })();

  return {
    household: getHouseholdById(householdId),
    plan: getPlanByHousehold(householdId)
  };
}

export function getPlanHistories(planId: string): any[] {
  return db.prepare(`
    SELECT * FROM plan_histories 
    WHERE plan_id = ? 
    ORDER BY operation_date DESC
  `).all(planId);
}

export function getAllPlans(): AssistancePlan[] {
  const rows = db.prepare('SELECT * FROM assistance_plans').all();
  return rows.map(parsePlan);
}
