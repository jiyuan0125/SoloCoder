import db from '../database';
import { Vote, VoteOption, Proposal, ProposalStage, Owner } from '../types';
import { generateId, now, calculateParticipationRate, calculateApprovalRate } from '../utils';
import { APPROVAL_THRESHOLD, PARTICIPATION_THRESHOLD } from '../constants';

const mapRowToVote = (row: any): Vote => ({
  id: row.id,
  proposalId: row.proposal_id,
  phone: row.phone,
  option: row.option as VoteOption,
  votedAt: row.voted_at
});

const mapRowToOwner = (row: any): Owner => ({
  id: row.id,
  phone: row.phone,
  name: row.name,
  createdAt: row.created_at
});

export const voteService = {
  registerOwner: (phone: string, name: string): Owner => {
    const existing = db.prepare('SELECT * FROM owners WHERE phone = ?').get(phone);
    if (existing) {
      return mapRowToOwner(existing);
    }
    
    const id = generateId();
    db.prepare(`
      INSERT INTO owners (id, phone, name, created_at) VALUES (?, ?, ?, ?)
    `).run(id, phone, name, now());
    
    return voteService.getOwnerByPhone(phone)!;
  },

  getOwnerByPhone: (phone: string): Owner | null => {
    const row = db.prepare('SELECT * FROM owners WHERE phone = ?').get(phone);
    return row ? mapRowToOwner(row) : null;
  },

  getTotalOwners: (): number => {
    const result: any = db.prepare('SELECT COUNT(*) as count FROM owners').get();
    return result?.count ?? 0;
  },

  castVote: (
    proposalId: string,
    phone: string,
    option: VoteOption
  ): Vote | null => {
    const existingVote = db.prepare(`
      SELECT * FROM votes WHERE proposal_id = ? AND phone = ?
    `).get(proposalId, phone);
    
    if (existingVote) {
      return null;
    }
    
    const id = generateId();
    db.prepare(`
      INSERT INTO votes (id, proposal_id, phone, option, voted_at)
      VALUES (?, ?, ?, ?, ?)
    `).run(id, proposalId, phone, option, now());
    
    return voteService.getVote(id)!;
  },

  getVote: (id: string): Vote | null => {
    const row = db.prepare('SELECT * FROM votes WHERE id = ?').get(id);
    return row ? mapRowToVote(row) : null;
  },

  getVotesByProposal: (proposalId: string): Vote[] => {
    const rows = db.prepare('SELECT * FROM votes WHERE proposal_id = ?').all(proposalId) as any[];
    return rows.map(mapRowToVote);
  },

  getVoteCountByOption: (proposalId: string, option: VoteOption): number => {
    const result: any = db.prepare(`
      SELECT COUNT(*) as count FROM votes WHERE proposal_id = ? AND option = ?
    `).get(proposalId, option);
    return result?.count ?? 0;
  },

  getTotalVotes: (proposalId: string): number => {
    const result: any = db.prepare(`
      SELECT COUNT(*) as count FROM votes WHERE proposal_id = ?
    `).get(proposalId);
    return result?.count ?? 0;
  },

  getVotingResult: (proposalId: string) => {
    const totalOwners = voteService.getTotalOwners();
    const totalVotes = voteService.getTotalVotes(proposalId);
    const approveVotes = voteService.getVoteCountByOption(proposalId, VoteOption.APPROVE);
    const opposeVotes = voteService.getVoteCountByOption(proposalId, VoteOption.OPPOSE);
    const abstainVotes = voteService.getVoteCountByOption(proposalId, VoteOption.ABSTAIN);
    
    const participationRate = calculateParticipationRate(totalVotes, totalOwners);
    const approvalRate = calculateApprovalRate(approveVotes, totalOwners);
    
    const isParticipationMet = participationRate >= PARTICIPATION_THRESHOLD;
    const isApproved = isParticipationMet && approvalRate > APPROVAL_THRESHOLD;
    
    return {
      proposalId,
      totalOwners,
      totalVotes,
      approveVotes,
      opposeVotes,
      abstainVotes,
      participationRate,
      approvalRate,
      isParticipationMet,
      isApproved,
      needsRerun: isParticipationMet && !isApproved
    };
  },

  addSuggestion: (
    proposalId: string,
    ownerId: string,
    content: string
  ) => {
    const id = generateId();
    db.prepare(`
      INSERT INTO suggestions (id, proposal_id, owner_id, content, created_at)
      VALUES (?, ?, ?, ?, ?)
    `).run(id, proposalId, ownerId, content, now());
    
    return {
      id,
      proposalId,
      ownerId,
      content,
      createdAt: now()
    };
  },

  getSuggestions: (proposalId: string) => {
    const rows = db.prepare(`
      SELECT * FROM suggestions WHERE proposal_id = ? ORDER BY created_at DESC
    `).all(proposalId) as any[];
    
    return rows.map((row: any) => ({
      id: row.id,
      proposalId: row.proposal_id,
      ownerId: row.owner_id,
      content: row.content,
      createdAt: row.created_at
    }));
  },

  addObjection: (
    proposalId: string,
    ownerId: string,
    reason: string
  ) => {
    const id = generateId();
    db.prepare(`
      INSERT INTO objections (id, proposal_id, owner_id, reason, created_at)
      VALUES (?, ?, ?, ?, ?)
    `).run(id, proposalId, ownerId, reason, now());
    
    return {
      id,
      proposalId,
      ownerId,
      reason,
      createdAt: now()
    };
  },

  getObjections: (proposalId: string) => {
    const rows = db.prepare(`
      SELECT * FROM objections WHERE proposal_id = ? ORDER BY created_at DESC
    `).all(proposalId) as any[];
    
    return rows.map((row: any) => ({
      id: row.id,
      proposalId: row.proposal_id,
      ownerId: row.owner_id,
      reason: row.reason,
      createdAt: row.created_at
    }));
  },

  getObjectionCount: (proposalId: string): number => {
    const result: any = db.prepare(`
      SELECT COUNT(*) as count FROM objections WHERE proposal_id = ?
    `).get(proposalId);
    return result?.count ?? 0;
  }
};
