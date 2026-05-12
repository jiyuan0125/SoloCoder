import db from '../database';
import { Proposal, ProposalStage, ReviewResult } from '../types';
import { generateId, now, isValidStageTransition } from '../utils';
import { STAGE_DURATION, SECRETARY_USER } from '../constants';

const mapRowToProposal = (row: any): Proposal => ({
  id: row.id,
  title: row.title,
  content: row.content,
  type: row.type,
  attachmentDescription: row.attachment_description,
  stage: row.stage as ProposalStage,
  stageStartTime: row.stage_start_time,
  createdBy: row.created_by,
  createdAt: row.created_at,
  isRerun: !!row.is_rerun,
  originalId: row.original_id,
  noticeEndTime: row.notice_end_time,
  votingEndTime: row.voting_end_time,
  executionEndTime: row.execution_end_time,
  isSuspended: !!row.is_suspended,
  publicNoticeEndTime: row.public_notice_end_time
});

export const proposalService = {
  submitProposal: (
    title: string,
    content: string,
    type: string,
    attachmentDescription: string,
    createdBy: string
  ): Proposal => {
    const id = generateId();
    const currentTime = now();
    
    const stmt = db.prepare(`
      INSERT INTO proposals (
        id, title, content, type, attachment_description,
        stage, stage_start_time, created_by, created_at, is_rerun
      ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 0)
    `);
    
    stmt.run(
      id,
      title,
      content,
      type,
      attachmentDescription,
      ProposalStage.SUBMITTED,
      currentTime,
      createdBy,
      currentTime
    );
    
    return proposalService.getProposal(id)!;
  },

  getProposal: (id: string): Proposal | null => {
    const row = db.prepare('SELECT * FROM proposals WHERE id = ?').get(id);
    return row ? mapRowToProposal(row) : null;
  },

  listProposals: (): Proposal[] => {
    const rows = db.prepare('SELECT * FROM proposals ORDER BY created_at DESC').all() as any[];
    return rows.map(mapRowToProposal);
  },

  startReview: (proposalId: string): Proposal | null => {
    const proposal = proposalService.getProposal(proposalId);
    if (!proposal) return null;
    
    if (proposal.stage !== ProposalStage.SUBMITTED) {
      return null;
    }
    
    db.prepare(`
      UPDATE proposals SET stage = ?, stage_start_time = ? WHERE id = ?
    `).run(ProposalStage.REVIEW, now(), proposalId);
    
    return proposalService.getProposal(proposalId);
  },

  completeReview: (
    proposalId: string,
    result: ReviewResult
  ): Proposal | null => {
    const proposal = proposalService.getProposal(proposalId);
    if (!proposal) return null;
    
    if (proposal.stage !== ProposalStage.REVIEW) {
      return null;
    }
    
    const targetStage = result === ReviewResult.PASS 
      ? ProposalStage.PUBLIC_NOTICE 
      : ProposalStage.REJECTED;
    
    if (!isValidStageTransition(proposal.stage, targetStage)) {
      return null;
    }
    
    const currentTime = now();
    const updates: any = {
      stage: targetStage,
      stage_start_time: currentTime
    };
    
    if (targetStage === ProposalStage.PUBLIC_NOTICE) {
      updates.notice_end_time = currentTime + STAGE_DURATION.PUBLIC_NOTICE;
    }
    
    if (targetStage === ProposalStage.REJECTED) {
      updates.public_notice_end_time = currentTime + STAGE_DURATION.EXECUTED_PUBLIC;
    }
    
    const setClauses = Object.keys(updates)
      .map(key => `${key} = ?`)
      .join(', ');
    const values = [...Object.values(updates), proposalId];
    
    db.prepare(`UPDATE proposals SET ${setClauses} WHERE id = ?`).run(...values);
    
    return proposalService.getProposal(proposalId);
  },

  startVoting: (proposalId: string): Proposal | null => {
    const proposal = proposalService.getProposal(proposalId);
    if (!proposal) return null;
    
    if (proposal.stage !== ProposalStage.PUBLIC_NOTICE) {
      return null;
    }
    
    if (!isValidStageTransition(proposal.stage, ProposalStage.VOTING)) {
      return null;
    }
    
    const currentTime = now();
    db.prepare(`
      UPDATE proposals SET 
        stage = ?, 
        stage_start_time = ?,
        voting_end_time = ?
      WHERE id = ?
    `).run(
      ProposalStage.VOTING,
      currentTime,
      currentTime + STAGE_DURATION.VOTING,
      proposalId
    );
    
    return proposalService.getProposal(proposalId);
  },

  completeVoting: (
    proposalId: string,
    isApproved: boolean
  ): Proposal | null => {
    const proposal = proposalService.getProposal(proposalId);
    if (!proposal) return null;
    
    if (proposal.stage !== ProposalStage.VOTING) {
      return null;
    }
    
    const targetStage = isApproved ? ProposalStage.EXECUTION : ProposalStage.REJECTED;
    
    if (!isValidStageTransition(proposal.stage, targetStage)) {
      return null;
    }
    
    const currentTime = now();
    const updates: any = {
      stage: targetStage,
      stage_start_time: currentTime
    };
    
    if (targetStage === ProposalStage.EXECUTION) {
      updates.execution_end_time = currentTime + STAGE_DURATION.EXECUTION_PUBLIC;
    } else {
      updates.public_notice_end_time = currentTime + STAGE_DURATION.EXECUTED_PUBLIC;
    }
    
    const setClauses = Object.keys(updates)
      .map(key => `${key} = ?`)
      .join(', ');
    const values = [...Object.values(updates), proposalId];
    
    db.prepare(`UPDATE proposals SET ${setClauses} WHERE id = ?`).run(...values);
    
    return proposalService.getProposal(proposalId);
  },

  completeExecution: (proposalId: string): Proposal | null => {
    const proposal = proposalService.getProposal(proposalId);
    if (!proposal) return null;
    
    if (proposal.stage !== ProposalStage.EXECUTION) {
      return null;
    }
    
    if (!isValidStageTransition(proposal.stage, ProposalStage.EXECUTED)) {
      return null;
    }
    
    const currentTime = now();
    db.prepare(`
      UPDATE proposals SET 
        stage = ?, 
        stage_start_time = ?,
        public_notice_end_time = ?
      WHERE id = ?
    `).run(
      ProposalStage.EXECUTED,
      currentTime,
      currentTime + STAGE_DURATION.EXECUTED_PUBLIC,
      proposalId
    );
    
    return proposalService.getProposal(proposalId);
  },

  suspendExecution: (proposalId: string): Proposal | null => {
    const proposal = proposalService.getProposal(proposalId);
    if (!proposal) return null;
    
    if (proposal.stage !== ProposalStage.EXECUTION) {
      return null;
    }
    
    db.prepare('UPDATE proposals SET is_suspended = 1 WHERE id = ?').run(proposalId);
    
    return proposalService.getProposal(proposalId);
  },

  createRerunProposal: (originalId: string): Proposal | null => {
    const original = proposalService.getProposal(originalId);
    if (!original) return null;
    
    const newId = `${original.id}-R`;
    const currentTime = now();
    
    db.prepare(`
      INSERT INTO proposals (
        id, title, content, type, attachment_description,
        stage, stage_start_time, created_by, created_at, is_rerun, original_id
      ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?)
    `).run(
      newId,
      original.title,
      original.content,
      original.type,
      original.attachmentDescription,
      ProposalStage.SUBMITTED,
      currentTime,
      original.createdBy,
      currentTime,
      originalId
    );
    
    return proposalService.getProposal(newId);
  },

  isDuplicate: (title: string, content: string, excludeId?: string): boolean => {
    let query = 'SELECT id FROM proposals WHERE title = ? AND content = ?';
    const params: any[] = [title, content];
    
    if (excludeId) {
      query += ' AND id != ?';
      params.push(excludeId);
    }
    
    const existing = db.prepare(query).get(...params);
    return !!existing;
  }
};
