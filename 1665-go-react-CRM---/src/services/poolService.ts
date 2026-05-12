import { db } from '../database';
import { Lead } from '../types';
import { SEVEN_DAYS_MS, THIRTY_DAYS_MS } from '../constants';
import { v4 as uuidv4 } from 'uuid';

export class PoolService {
  private static findUnclaimedExpiredLeadsStmt = db.prepare(`
    SELECT * FROM leads 
    WHERE assigned_sales_id IS NULL 
    AND is_in_pool = 0
    AND status = 'new'
    AND created_at <= ?
  `);

  private static findStaleAssignedLeadsStmt = db.prepare(`
    SELECT * FROM leads 
    WHERE assigned_sales_id IS NOT NULL 
    AND is_in_pool = 0
    AND status NOT IN ('converted', 'rejected')
    AND updated_at <= ?
  `);

  private static moveToPoolStmt = db.prepare(`
    UPDATE leads 
    SET assigned_sales_id = NULL, is_in_pool = 1, updated_at = ?, version = version + 1
    WHERE id = ?
  `);

  private static insertPoolRecordStmt = db.prepare(`
    INSERT INTO sales_pool (lead_id, entered_at, reason)
    VALUES (?, ?, ?)
  `);

  static findExpiredLeads(): Lead[] {
    const now = Date.now();
    const sevenDaysAgo = now - SEVEN_DAYS_MS;
    const thirtyDaysAgo = now - THIRTY_DAYS_MS;

    const unclaimedExpired = this.findUnclaimedExpiredLeadsStmt.all(sevenDaysAgo) as Lead[];
    const staleAssigned = this.findStaleAssignedLeadsStmt.all(thirtyDaysAgo) as Lead[];

    const leadMap = new Map<string, Lead>();
    unclaimedExpired.forEach((lead) => leadMap.set(lead.id, lead));
    staleAssigned.forEach((lead) => leadMap.set(lead.id, lead));

    return Array.from(leadMap.values());
  }

  static moveToPool(leadId: string, reason: string): void {
    const now = Date.now();
    this.moveToPoolStmt.run(now, leadId);
    this.insertPoolRecordStmt.run(leadId, now, reason);
  }

  static processExpiredLeads(): { moved: number; leads: Lead[] } {
    const expiredLeads = this.findExpiredLeads();
    const now = Date.now();

    for (const lead of expiredLeads) {
      let reason = '';
      if (lead.assigned_sales_id === null) {
        reason = `创建7天未被认领 (创建时间: ${new Date(lead.created_at).toISOString()})`;
      } else {
        reason = `认领后30天未更新状态 (最后更新: ${new Date(lead.updated_at).toISOString()})`;
      }

      const transaction = db.transaction(() => {
        this.moveToPoolStmt.run(now, lead.id);
        this.insertPoolRecordStmt.run(lead.id, now, reason);
      });

      transaction();
    }

    return {
      moved: expiredLeads.length,
      leads: expiredLeads,
    };
  }
}
