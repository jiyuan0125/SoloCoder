export enum ProposalStage {
  SUBMITTED = 'submitted',
  REVIEW = 'review',
  PUBLIC_NOTICE = 'public_notice',
  VOTING = 'voting',
  EXECUTION = 'execution',
  EXECUTED = 'executed',
  REJECTED = 'rejected'
}

export enum VoteOption {
  APPROVE = 'approve',
  OPPOSE = 'oppose',
  ABSTAIN = 'abstain'
}

export enum TodoType {
  REVIEW = 'review',
  VOTING_ADMIN = 'voting_admin',
  NOTIFY_DIRECTOR = 'notify_director'
}

export enum TodoStatus {
  PENDING = 'pending',
  COMPLETED = 'completed',
  ESCALATED = 'escalated'
}

export enum ReviewResult {
  PASS = 'pass',
  REJECT = 'reject'
}

export interface Proposal {
  id: string;
  title: string;
  content: string;
  type: string;
  attachmentDescription: string;
  stage: ProposalStage;
  stageStartTime: number;
  createdBy: string;
  createdAt: number;
  isRerun: boolean;
  originalId?: string;
  noticeEndTime?: number;
  votingEndTime?: number;
  executionEndTime?: number;
  isSuspended?: boolean;
  publicNoticeEndTime?: number;
}

export interface Vote {
  id: string;
  proposalId: string;
  phone: string;
  option: VoteOption;
  votedAt: number;
}

export interface Todo {
  id: string;
  proposalId: string;
  type: TodoType;
  assignee: string;
  status: TodoStatus;
  createdAt: number;
  handledAt?: number;
  escalatedAt?: number;
}

export interface Owner {
  id: string;
  phone: string;
  name: string;
  createdAt: number;
}

export interface Suggestion {
  id: string;
  proposalId: string;
  ownerId: string;
  content: string;
  createdAt: number;
}

export interface Objection {
  id: string;
  proposalId: string;
  ownerId: string;
  reason: string;
  createdAt: number;
}
