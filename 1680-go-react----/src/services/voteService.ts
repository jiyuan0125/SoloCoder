import { db } from '../database';
import {
  VoteSetting,
  VoteType,
  VoteOption,
  Vote,
  VoteResult,
  Issue,
  IssueStatus,
  User
} from '../types';
import { nowISO, parseISO, isAfter, isBefore } from '../utils';

export const createVoteSetting = (
  issueId: number,
  startAt: string,
  endAt: string,
  voteType: VoteType,
  minParticipationRate: number
): VoteSetting => {
  const now = nowISO();
  const result = db.prepare(`
    INSERT INTO vote_settings (
      issue_id, start_at, end_at, vote_type, min_participation_rate, created_at
    ) VALUES (?, ?, ?, ?, ?, ?)
  `).run(issueId, startAt, endAt, voteType, minParticipationRate, now);

  return getVoteSettingById(result.lastInsertRowid as number)!;
};

export const getVoteSettingById = (id: number): VoteSetting | undefined => {
  const row = db.prepare(`
    SELECT 
      id, issue_id as issueId, start_at as startAt, 
      end_at as endAt, vote_type as voteType,
      min_participation_rate as minParticipationRate,
      created_at as createdAt
    FROM vote_settings WHERE id = ?
  `).get(id) as any;
  return row;
};

export const getVoteSettingByIssueId = (issueId: number): VoteSetting | undefined => {
  const row = db.prepare(`
    SELECT 
      id, issue_id as issueId, start_at as startAt, 
      end_at as endAt, vote_type as voteType,
      min_participation_rate as minParticipationRate,
      created_at as createdAt
    FROM vote_settings WHERE issue_id = ?
  `).get(issueId) as any;
  return row;
};

export const isVotingPeriod = (setting: VoteSetting): boolean => {
  const now = new Date();
  const start = parseISO(setting.startAt);
  const end = parseISO(setting.endAt);
  return !isBefore(now, start) && !isAfter(now, end);
};

export const isVotingEnded = (setting: VoteSetting): boolean => {
  const now = new Date();
  const end = parseISO(setting.endAt);
  return isAfter(now, end);
};

export const calculateWeight = (user: User, voteType: VoteType): number => {
  if (voteType === VoteType.ONE_PERSON) {
    return 1;
  }
  return Math.round(user.houseArea * 100);
};

export const castVote = (
  issueId: number,
  userId: number,
  option: VoteOption,
  weight: number
): Vote => {
  const now = nowISO();
  const result = db.prepare(`
    INSERT INTO votes (issue_id, user_id, option, weight, voted_at)
    VALUES (?, ?, ?, ?, ?)
  `).run(issueId, userId, option, weight, now);

  return getVoteById(result.lastInsertRowid as number)!;
};

export const getVoteById = (id: number): Vote | undefined => {
  const row = db.prepare(`
    SELECT 
      id, issue_id as issueId, user_id as userId,
      option, weight, voted_at as votedAt
    FROM votes WHERE id = ?
  `).get(id) as any;
  return row;
};

export const hasVoted = (issueId: number, userId: number): boolean => {
  const result = db.prepare(`
    SELECT COUNT(*) as count FROM votes WHERE issue_id = ? AND user_id = ?
  `).get(issueId, userId) as { count: number };
  return result.count > 0;
};

export const getTotalRegisteredOwners = (): number => {
  const result = db.prepare(`
    SELECT COUNT(*) as count FROM users WHERE is_registered = 1
  `).get() as { count: number };
  return result.count;
};

export const calculateVoteResult = (issueId: number): VoteResult => {
  const votes = db.prepare(`
    SELECT option, weight FROM votes WHERE issue_id = ?
  `).all(issueId) as { option: string; weight: number }[];

  let approves = 0;
  let opposes = 0;
  let abstains = 0;
  const voterIds = new Set<number>();

  const allVotes = db.prepare(`
    SELECT user_id as userId, option, weight FROM votes WHERE issue_id = ?
  `).all(issueId) as { userId: number; option: string; weight: number }[];

  for (const vote of allVotes) {
    voterIds.add(vote.userId);
    if (vote.option === VoteOption.APPROVE) {
      approves += vote.weight;
    } else if (vote.option === VoteOption.OPPOSE) {
      opposes += vote.weight;
    } else if (vote.option === VoteOption.ABSTAIN) {
      abstains += vote.weight;
    }
  }

  const totalVoters = voterIds.size;
  const totalRegistered = getTotalRegisteredOwners();
  const totalWeight = approves + opposes + abstains;

  let participationRate = 0;
  if (totalRegistered > 0) {
    participationRate = totalVoters / totalRegistered;
  }

  let approveRate = 0;
  const effectiveTotal = approves + opposes;
  if (effectiveTotal > 0) {
    approveRate = approves / effectiveTotal;
  }

  const setting = getVoteSettingByIssueId(issueId);
  const minRate = setting?.minParticipationRate ?? 0;

  const isPassed = approveRate > 0.5 && participationRate >= minRate;
  const isReview = !isPassed && approveRate > (2 / 3) && participationRate < minRate;

  return {
    issueId,
    approves,
    opposes,
    abstains,
    totalVoters,
    participationRate,
    approveRate,
    isPassed,
    isReview
  };
};

export const canViewResults = (issue: Issue, setting: VoteSetting | undefined): boolean => {
  if (!setting) return false;
  return isVotingEnded(setting);
};
