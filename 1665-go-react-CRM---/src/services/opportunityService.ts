import { db } from '../database';
import { Opportunity, OpportunityStage, FunnelStat, LeadStatus } from '../types';
import { STAGE_ORDER, STAGE_NAMES, TERMINAL_STAGES } from '../constants';
import { v4 as uuidv4 } from 'uuid';
import { LeadService } from './leadService';

export interface CreateOpportunityData {
  lead_id: string;
  expected_amount: number;
  assigned_sales_id: string;
  expected_close_date?: string;
}

export interface UpdateStageData {
  stage: OpportunityStage;
  sales_person_id: string;
  reason?: string;
  lost_to?: string;
}

export class OpportunityService {
  private static getOpportunityByIdStmt = db.prepare(`SELECT * FROM opportunities WHERE id = ?`);
  private static insertOpportunityStmt = db.prepare(`
    INSERT INTO opportunities (
      id, lead_id, expected_amount, assigned_sales_id, stage, 
      expected_close_date, status, created_at, updated_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
  `);
  private static updateOpportunityStageStmt = db.prepare(`
    UPDATE opportunities 
    SET stage = ?, status = ?, updated_at = ? 
    WHERE id = ?
  `);
  private static insertStageHistoryStmt = db.prepare(`
    INSERT INTO opportunity_stage_history (
      id, opportunity_id, from_stage, to_stage, changed_by, 
      reason, lost_to, created_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
  `);
  private static getAllOpportunitiesStmt = db.prepare(`
    SELECT * FROM opportunities ORDER BY created_at DESC
  `);
  private static getOpportunitiesByStageStmt = db.prepare(`
    SELECT * FROM opportunities WHERE stage = ? ORDER BY created_at DESC
  `);
  private static countByStageStmt = db.prepare(`
    SELECT 
      stage,
      COUNT(*) as count,
      SUM(expected_amount) as total_amount
    FROM opportunities 
    GROUP BY stage
  `);

  static createOpportunity(data: CreateOpportunityData): Opportunity {
    if (data.expected_amount < 0) {
      throw new Error('预计金额不能为负数');
    }

    const lead = LeadService.getLeadById(data.lead_id);
    if (!lead) {
      throw new Error('线索不存在');
    }

    if (lead.status !== 'converted') {
      throw new Error('只有已转化的线索才能创建商机');
    }

    const now = Date.now();
    const id = uuidv4();
    const opportunity: Opportunity = {
      id,
      lead_id: data.lead_id,
      expected_amount: data.expected_amount,
      assigned_sales_id: data.assigned_sales_id,
      stage: 'initial_contact',
      expected_close_date: data.expected_close_date,
      status: 'active',
      created_at: now,
      updated_at: now,
    };

    this.insertOpportunityStmt.run(
      opportunity.id,
      opportunity.lead_id,
      opportunity.expected_amount,
      opportunity.assigned_sales_id,
      opportunity.stage,
      opportunity.expected_close_date,
      opportunity.status,
      opportunity.created_at,
      opportunity.updated_at
    );

    this.insertStageHistoryStmt.run(
      uuidv4(),
      opportunity.id,
      null,
      opportunity.stage,
      data.assigned_sales_id,
      '商机创建',
      null,
      now
    );

    return opportunity;
  }

  static getOpportunityById(id: string): Opportunity | undefined {
    return this.getOpportunityByIdStmt.get(id) as Opportunity | undefined;
  }

  static getAllOpportunities(): Opportunity[] {
    return this.getAllOpportunitiesStmt.all() as Opportunity[];
  }

  static getOpportunitiesByStage(stage: OpportunityStage): Opportunity[] {
    return this.getOpportunitiesByStageStmt.all(stage) as Opportunity[];
  }

  static isValidStage(stage: string): stage is OpportunityStage {
    return STAGE_ORDER.includes(stage as OpportunityStage);
  }

  static isTerminalStage(stage: OpportunityStage): boolean {
    return TERMINAL_STAGES.includes(stage);
  }

  static getStageIndex(stage: OpportunityStage): number {
    return STAGE_ORDER.indexOf(stage);
  }

  static updateStage(
    opportunityId: string,
    data: UpdateStageData
  ): Opportunity {
    const opportunity = this.getOpportunityById(opportunityId);
    if (!opportunity) {
      throw new Error('商机不存在');
    }

    if (!this.isValidStage(data.stage)) {
      throw new Error('无效的阶段');
    }

    if (this.isTerminalStage(opportunity.stage)) {
      throw new Error('赢单和输单之后不能再修改阶段');
    }

    if (data.stage === 'lost' && !data.lost_to) {
      throw new Error('输单需要填写输给谁');
    }

    const now = Date.now();
    let newStatus = opportunity.status;

    if (data.stage === 'won') {
      newStatus = 'won';
    } else if (data.stage === 'lost') {
      newStatus = 'lost';
    } else {
      newStatus = 'active';
    }

    this.updateOpportunityStageStmt.run(
      data.stage,
      newStatus,
      now,
      opportunityId
    );

    this.insertStageHistoryStmt.run(
      uuidv4(),
      opportunityId,
      opportunity.stage,
      data.stage,
      data.sales_person_id,
      data.reason,
      data.lost_to || null,
      now
    );

    const updatedOpportunity = this.getOpportunityById(opportunityId);
    if (!updatedOpportunity) {
      throw new Error('更新商机失败');
    }

    return updatedOpportunity;
  }

  static getFunnelStatistics(): FunnelStat[] {
    const stageStats = this.countByStageStmt.all() as Array<{
      stage: string;
      count: number;
      total_amount: number | null;
    }>;

    const statsMap: Record<string, { count: number; total_amount: number }> = {};
    stageStats.forEach((stat) => {
      statsMap[stat.stage] = {
        count: stat.count,
        total_amount: stat.total_amount || 0,
      };
    });

    const funnelStats: FunnelStat[] = [];

    for (let i = 0; i < STAGE_ORDER.length; i++) {
      const stage = STAGE_ORDER[i];
      const stat = statsMap[stage] || { count: 0, total_amount: 0 };

      let conversionRate = 0;
      if (i < STAGE_ORDER.length - 1 && stat.count > 0) {
        const nextStage = STAGE_ORDER[i + 1];
        const nextStat = statsMap[nextStage] || { count: 0 };
        conversionRate = nextStat.count / stat.count;
      }

      funnelStats.push({
        stage,
        stage_name: STAGE_NAMES[stage],
        opportunity_count: stat.count,
        total_amount: stat.total_amount,
        conversion_rate: conversionRate,
      });
    }

    return funnelStats;
  }
}
