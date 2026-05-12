export enum UserRole {
  OWNER = 'owner',
  COMMITTEE = 'committee',
  EXECUTOR = 'executor'
}

export enum IssueStatus {
  DRAFT = 'draft',
  PUBLICITY = 'publicity',
  WAITING_VOTE_CONFIRM = 'waiting_vote_confirm',
  VOTING = 'voting',
  VOTED = 'voted',
  WAITING_EXEC_CONFIRM = 'waiting_exec_confirm',
  EXECUTING = 'executing',
  COMPLETED = 'completed',
  REJECTED = 'rejected',
  REPEAL = 'repeal'
}

export enum VoteType {
  ONE_PERSON = 'one_person',
  BY_AREA = 'by_area'
}

export enum VoteOption {
  APPROVE = 'approve',
  OPPOSE = 'oppose',
  ABSTAIN = 'abstain'
}

export interface User {
  id: number;
  username: string;
  name: string;
  role: UserRole;
  isRegistered: number;
  isVerified: number;
  houseArea: number;
  createdAt: string;
}

export interface Issue {
  id: number;
  title: string;
  content: string;
  category: string;
  attachments: string;
  initiatorId: number;
  status: IssueStatus;
  publicityStartAt: string;
  publicityEndAt: string;
  createdAt: string;
  updatedAt: string;
}

export interface VoteSetting {
  id: number;
  issueId: number;
  startAt: string;
  endAt: string;
  voteType: VoteType;
  minParticipationRate: number;
  createdAt: string;
}

export interface Vote {
  id: number;
  issueId: number;
  userId: number;
  option: VoteOption;
  weight: number;
  votedAt: string;
}

export interface Opinion {
  id: number;
  issueId: number;
  userId: number;
  content: string;
  createdAt: string;
}

export interface Todo {
  id: number;
  issueId: number;
  userId: number;
  type: string;
  deadlineAt: string;
  isCompleted: number;
  isReminded: number;
  createdAt: string;
}

export interface VoteResult {
  issueId: number;
  approves: number;
  opposes: number;
  abstains: number;
  totalVoters: number;
  participationRate: number;
  approveRate: number;
  isPassed: boolean;
  isReview: boolean;
}

export interface ApiError {
  code: number;
  message: string;
}
