export enum CommunityType {
  CUSTOMER = 'customer',
  ACTIVITY = 'activity',
  VIP = 'vip',
  GENERAL = 'general'
}

export enum UserTier {
  ACTIVE = 'active',
  SILENT = 'silent',
  CHURNED = 'churned'
}

export enum TaskStatus {
  PENDING = 'pending',
  RUNNING = 'running',
  COMPLETED = 'completed',
  CANCELLED = 'cancelled'
}

export enum SendResult {
  SUCCESS = 'success',
  FAILED = 'failed',
  SKIPPED = 'target_not_exists'
}

export interface Community {
  id: string;
  name: string;
  type: CommunityType;
  maxMembers: number;
  currentMembers: number;
  createdAt: number;
}

export interface MemberRecord {
  id: string;
  communityId: string;
  userId: string;
  action: 'join' | 'leave';
  timestamp: number;
}

export interface Tag {
  id: string;
  name: string;
  category: string;
  createdAt: number;
}

export interface User {
  id: string;
  name: string;
  lastActiveAt: number;
  createdAt: number;
}

export interface BroadcastTask {
  id: string;
  name: string;
  content: string;
  scheduledAt: number;
  status: TaskStatus;
  createdAt: number;
  tierFilter?: UserTier;
  tagFilterMode?: 'AND' | 'OR';
}

export interface TaskTargetCommunity {
  id: string;
  taskId: string;
  communityId: string;
  sendResult?: SendResult;
  failureReason?: string;
  sentAt?: number;
}

export interface TaskTagFilter {
  id: string;
  taskId: string;
  tagId: string;
}
