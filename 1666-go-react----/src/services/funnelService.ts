import { db } from '../db';
import { Stage, FunnelData, FunnelQuery } from '../types';
import { STAGE_ORDER, getStageIndex } from '../constants';
import { salespersonExists, getTeamSalespeople } from './salespersonService';

interface HistoricalTransition {
  from: Stage;
  to: Stage;
  count: number;
}

interface CurrentStageOpportunity {
  opportunity_id: string;
  current_stage: Stage;
  amount: number;
}

interface StageEntry {
  stage: Stage;
  count: number;
}

export interface FunnelResult {
  success: boolean;
  data?: FunnelData[];
  error?: { status: number; message: string };
}

function getSalespersonIdsForFilter(salespersonId?: string, teamId?: string): string[] | null {
  if (salespersonId) {
    return [salespersonId];
  }
  if (teamId) {
    const teamMembers = getTeamSalespeople(teamId);
    return teamMembers.map((m) => m.id);
  }
  return null;
}

function buildSalespersonFilter(salespersonIds: string[] | null): { clause: string; params: any[] } {
  if (!salespersonIds || salespersonIds.length === 0) {
    return { clause: '', params: [] };
  }
  const placeholders = salespersonIds.map(() => '?').join(', ');
  return {
    clause: `AND o.salesperson_id IN (${placeholders})`,
    params: salespersonIds,
  };
}

function getHistoricalTransitions(salespersonIds: string[] | null): HistoricalTransition[] {
  const { clause, params } = buildSalespersonFilter(salespersonIds);

  const query = `
    SELECT
      sh.from_stage as fromStage,
      sh.to_stage as toStage,
      COUNT(DISTINCT sh.opportunity_id) as count
    FROM stage_history sh
    INNER JOIN opportunities o ON sh.opportunity_id = o.id
    WHERE sh.from_stage IS NOT NULL
      AND sh.from_stage != '输单'
      AND sh.to_stage != '输单'
      ${clause}
    GROUP BY sh.from_stage, sh.to_stage
  `;

  const rows = db.prepare(query).all(...params) as { fromStage: string; toStage: string; count: number }[];

  return rows.map((r) => ({
    from: r.fromStage as Stage,
    to: r.toStage as Stage,
    count: r.count,
  }));
}

function calculateHistoricalConversionRates(
  transitions: HistoricalTransition[]
): Map<Stage, { nextStageRate: number; winRate: number }> {
  const stageCounts = new Map<Stage, number>();
  const advancementCounts = new Map<Stage, number>();

  for (const transition of transitions) {
    const fromIdx = getStageIndex(transition.from);
    const toIdx = getStageIndex(transition.to);

    stageCounts.set(
      transition.from,
      (stageCounts.get(transition.from) || 0) + transition.count
    );

    if (toIdx > fromIdx && toIdx - fromIdx === 1) {
      advancementCounts.set(
        transition.from,
        (advancementCounts.get(transition.from) || 0) + transition.count
      );
    }
  }

  const rates = new Map<Stage, { nextStageRate: number; winRate: number }>();

  for (let i = STAGE_ORDER.length - 1; i >= 0; i--) {
    const stage = STAGE_ORDER[i];

    if (stage === '赢单') {
      rates.set(stage, { nextStageRate: 1, winRate: 1 });
      continue;
    }

    const total = stageCounts.get(stage) || 0;
    const advanced = advancementCounts.get(stage) || 0;
    const nextStage = STAGE_ORDER[i + 1];

    const nextStageRate = total === 0 ? 0 : advanced / total;
    const nextWinRate = nextStage ? (rates.get(nextStage)?.winRate || 0) : 0;
    const winRate = nextStageRate * nextWinRate;

    rates.set(stage, { nextStageRate, winRate });
  }

  return rates;
}

function getStageEntriesInTimeRange(
  startDate: string,
  endDate: string,
  salespersonIds: string[] | null
): StageEntry[] {
  const { clause, params } = buildSalespersonFilter(salespersonIds);

  const query = `
    SELECT
      sh.to_stage as stage,
      COUNT(DISTINCT sh.opportunity_id) as count
    FROM stage_history sh
    INNER JOIN opportunities o ON sh.opportunity_id = o.id
    WHERE sh.timestamp >= ?
      AND sh.timestamp <= ?
      ${clause}
    GROUP BY sh.to_stage
  `;

  const rows = db.prepare(query).all(startDate, endDate, ...params) as { stage: string; count: number }[];

  return rows.map((r) => ({
    stage: r.stage as Stage,
    count: r.count,
  }));
}

function getActiveOpportunitiesByStage(salespersonIds: string[] | null): CurrentStageOpportunity[] {
  const { clause, params } = buildSalespersonFilter(salespersonIds);

  const query = `
    SELECT
      o.id as opportunity_id,
      o.current_stage as currentStage,
      o.amount as amount
    FROM opportunities o
    WHERE o.current_stage != '输单'
      AND o.current_stage != '赢单'
      ${clause}
  `;

  const rows = db.prepare(query).all(...params) as {
    opportunity_id: string;
    currentStage: string;
    amount: number;
  }[];

  return rows.map((r) => ({
    opportunity_id: r.opportunity_id,
    current_stage: r.currentStage as Stage,
    amount: r.amount,
  }));
}

function getStageAmountsFromActive(
  opportunities: CurrentStageOpportunity[]
): Map<Stage, number> {
  const amounts = new Map<Stage, number>();

  for (const opp of opportunities) {
    const current = amounts.get(opp.current_stage) || 0;
    amounts.set(opp.current_stage, current + opp.amount);
  }

  return amounts;
}

export function calculateFunnel(query: FunnelQuery): FunnelResult {
  if (query.salespersonId && !salespersonExists(query.salespersonId)) {
    return {
      success: false,
      error: { status: 404, message: '销售ID不存在' },
    };
  }

  const salespersonIds = getSalespersonIdsForFilter(query.salespersonId, query.teamId);

  if (query.teamId && (!salespersonIds || salespersonIds.length === 0)) {
    return {
      success: true,
      data: [],
    };
  }

  const readTx = db.transaction(() => {
    const stageEntries = getStageEntriesInTimeRange(query.startDate, query.endDate, salespersonIds);
    const transitions = getHistoricalTransitions(salespersonIds);
    const activeOpps = getActiveOpportunitiesByStage(salespersonIds);

    return { stageEntries, transitions, activeOpps };
  });

  const { stageEntries, transitions, activeOpps } = readTx();

  const stageCountMap = new Map<Stage, number>();
  for (const entry of stageEntries) {
    if (entry.stage !== '输单') {
      stageCountMap.set(entry.stage, entry.count);
    }
  }

  const conversionRates = calculateHistoricalConversionRates(transitions);
  const stageAmounts = getStageAmountsFromActive(activeOpps);

  const funnelData: FunnelData[] = [];

  for (let i = 0; i < STAGE_ORDER.length; i++) {
    const stage = STAGE_ORDER[i];
    const count = stageCountMap.get(stage) || 0;

    let conversionRate = 0;
    if (i < STAGE_ORDER.length - 1) {
      const nextStage = STAGE_ORDER[i + 1];
      const nextCount = stageCountMap.get(nextStage) || 0;
      conversionRate = count === 0 ? 0 : nextCount / count;
    } else {
      conversionRate = 1;
    }

    const amount = stageAmounts.get(stage) || 0;
    const rates = conversionRates.get(stage) || { nextStageRate: 0, winRate: 0 };
    const predictedWinAmount = amount * rates.winRate;

    funnelData.push({
      stage,
      count,
      conversionRate,
      amount,
      predictedWinAmount,
    });
  }

  return { success: true, data: funnelData };
}
