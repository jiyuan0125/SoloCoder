export type ContentType = 'post' | 'comment';

export type WindowType = 'minute' | 'hour' | 'day';

export type PostStatus =
  | 'published'
  | 'pending_review'
  | 'suspicious'
  | 'blocked'
  | 'publisher_blocked'
  | 'approved'
  | 'rejected';

export interface RateLimitResult {
  allowed: boolean;
  exceededDimension?: WindowType;
}

export interface ContentAnalysisResult {
  isSuspicious: boolean;
  reasons: string[];
}

export interface User {
  id: string;
  is_whitelist: number;
  created_at: number;
}

export interface BlacklistEntry {
  user_id: string;
  added_at: number;
  expires_at: number | null;
  reason: string | null;
}

export interface Post {
  id: string;
  user_id: string;
  content: string;
  content_type: ContentType;
  status: PostStatus;
  created_at: number;
}

export interface MarkBlacklistResult {
  success: boolean;
  totalCleaned: number;
  successCleaned: number;
  failedCleaned: number;
}