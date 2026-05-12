import { db } from '../database';
import { Lead, LeadStatus } from '../types';
import { LEAD_STATUS_TRANSITIONS, VALID_LEAD_STATUSES } from '../constants';
import { v4 as uuidv4 } from 'uuid';

export interface CreateLeadData {
  source_channel: string;
  contact_info: string;
  company_name?: string;
  contact_person?: string;
  requirements?: string;
}

export interface UpdateLeadStatusData {
  status: LeadStatus;
  sales_person_id: string;
  reason?: string;
}

export class LeadService {
  private static getLeadByIdStmt = db.prepare(`SELECT * FROM leads WHERE id = ?`);
  private static insertLeadStmt = db.prepare(`
    INSERT INTO leads (
      id, source_channel, contact_info, company_name, contact_person, 
      requirements, status, created_at, updated_at, version
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
  `);
  private static updateLeadStatusStmt = db.prepare(`
    UPDATE leads SET status = ?, updated_at = ?, version = version + 1 
    WHERE id = ? AND version = ?
  `);
  private static insertStatusHistoryStmt = db.prepare(`
    INSERT INTO lead_status_history (
      id, lead_id, from_status, to_status, changed_by, reason, created_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?)
  `);
  private static claimLeadStmt = db.prepare(`
    UPDATE leads 
    SET assigned_sales_id = ?, updated_at = ?, version = version + 1, is_in_pool = 0 
    WHERE id = ? AND assigned_sales_id IS NULL AND version = ?
  `);
  private static getAllLeadsStmt = db.prepare(`SELECT * FROM leads ORDER BY created_at DESC`);
  private static getPoolLeadsStmt = db.prepare(`SELECT * FROM leads WHERE is_in_pool = 1 ORDER BY created_at DESC`);

  static createLead(data: CreateLeadData): Lead {
    if (!data.source_channel || !data.source_channel.trim()) {
      throw new Error('来源渠道不能为空');
    }
    if (!data.contact_info || !data.contact_info.trim()) {
      throw new Error('联系方式不能为空');
    }

    const now = Date.now();
    const id = uuidv4();
    const lead: Lead = {
      id,
      source_channel: data.source_channel.trim(),
      contact_info: data.contact_info.trim(),
      company_name: data.company_name?.trim(),
      contact_person: data.contact_person?.trim(),
      requirements: data.requirements?.trim(),
      status: 'new',
      created_at: now,
      updated_at: now,
      is_in_pool: 0,
      version: 1,
    };

    this.insertLeadStmt.run(
      lead.id,
      lead.source_channel,
      lead.contact_info,
      lead.company_name,
      lead.contact_person,
      lead.requirements,
      lead.status,
      lead.created_at,
      lead.updated_at,
      lead.version
    );

    return lead;
  }

  static getLeadById(id: string): Lead | undefined {
    return this.getLeadByIdStmt.get(id) as Lead | undefined;
  }

  static getAllLeads(): Lead[] {
    return this.getAllLeadsStmt.all() as Lead[];
  }

  static getPoolLeads(): Lead[] {
    return this.getPoolLeadsStmt.all() as Lead[];
  }

  static canTransitionStatus(current: LeadStatus, next: LeadStatus): boolean {
    if (!VALID_LEAD_STATUSES.includes(next)) {
      return false;
    }
    const allowed = LEAD_STATUS_TRANSITIONS[current];
    return allowed?.includes(next) || false;
  }

  static updateLeadStatus(
    leadId: string,
    data: UpdateLeadStatusData
  ): Lead {
    const lead = this.getLeadById(leadId);
    if (!lead) {
      throw new Error('线索不存在');
    }

    if (!this.canTransitionStatus(lead.status, data.status)) {
      if (lead.status === 'rejected') {
        throw new Error('已拒绝的线索不能再修改状态');
      }
      if (lead.status === 'converted') {
        throw new Error('已转化的线索不能再修改状态');
      }
      throw new Error(`不能从${lead.status}状态转换到${data.status}状态`);
    }

    const now = Date.now();
    const oldVersion = lead.version;
    const result = this.updateLeadStatusStmt.run(data.status, now, leadId, oldVersion);

    if (result.changes === 0) {
      throw new Error('版本冲突，请重试');
    }

    this.insertStatusHistoryStmt.run(
      uuidv4(),
      leadId,
      lead.status,
      data.status,
      data.sales_person_id,
      data.reason,
      now
    );

    const updatedLead = this.getLeadById(leadId);
    if (!updatedLead) {
      throw new Error('更新线索失败');
    }

    return updatedLead;
  }

  static claimLead(leadId: string, salesPersonId: string): Lead {
    const lead = this.getLeadById(leadId);
    if (!lead) {
      throw new Error('线索不存在');
    }

    if (lead.assigned_sales_id) {
      throw new Error('线索已被认领');
    }

    const now = Date.now();
    const oldVersion = lead.version;
    const result = this.claimLeadStmt.run(salesPersonId, now, leadId, oldVersion);

    if (result.changes === 0) {
      throw new Error('线索已被认领');
    }

    const updatedLead = this.getLeadById(leadId);
    if (!updatedLead) {
      throw new Error('认领线索失败');
    }

    return updatedLead;
  }
}
