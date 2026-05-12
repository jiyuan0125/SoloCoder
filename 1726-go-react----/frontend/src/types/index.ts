export type UserRole = 'author' | 'editor' | 'reviewer' | 'admin';
export type PaperStatus = 'draft' | 'pending_review' | 'in_review' | 'revision' | 'accepted' | 'published' | 'rejected';
export type ReviewDecision = 'accept' | 'revision' | 'reject';

export interface User {
  id: string;
  name: string;
  role: UserRole;
}

export interface Author {
  id: string;
  name: string;
  affiliation: string;
  is_first: boolean;
  is_corresponding: boolean;
}

export interface Reviewer {
  id: string;
  name: string;
  affiliation: string;
  assigned_at: string;
  submitted: boolean;
}

export interface Review {
  id: string;
  reviewer_id: string;
  reviewer_name: string;
  decision: ReviewDecision;
  score: number;
  comments: string;
  submitted_at: string;
}

export interface PublicationInfo {
  journal_name: string;
  volume: string;
  issue: string;
  page_start: string;
  page_end: string;
  doi: string;
}

export interface VersionHistory {
  id: string;
  paper_id: string;
  status: PaperStatus;
  operator_id: string;
  operator_name: string;
  changed_at: string;
  description: string;
}

export interface Paper {
  id: string;
  title: string;
  abstract: string;
  keywords: string[];
  authors: Author[];
  subject_category: string;
  target_journal: string;
  status: PaperStatus;
  assigned_reviewers: Reviewer[];
  reviews: Review[];
  publication_info?: PublicationInfo;
  version_history: VersionHistory[];
  created_by: string;
  created_at: string;
  updated_at: string;
}

export const statusDisplayMap: Record<PaperStatus, string> = {
  draft: '草稿',
  pending_review: '待审稿',
  in_review: '审稿中',
  revision: '修改中',
  accepted: '已录用',
  published: '已发表',
  rejected: '已拒稿',
};

export const statusColorMap: Record<PaperStatus, string> = {
  draft: '#6b7280',
  pending_review: '#f59e0b',
  in_review: '#3b82f6',
  revision: '#f97316',
  accepted: '#10b981',
  published: '#059669',
  rejected: '#ef4444',
};

export const decisionDisplayMap: Record<ReviewDecision, string> = {
  accept: '通过',
  revision: '修改后重审',
  reject: '拒稿',
};

export const decisionColorMap: Record<ReviewDecision, string> = {
  accept: '#10b981',
  revision: '#f59e0b',
  reject: '#ef4444',
};
