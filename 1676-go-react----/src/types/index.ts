export type ProjectStatus = 'preparing' | 'funding' | 'succeeded' | 'failed' | 'cancelled';

export type ProjectCategory = 'film' | 'music' | 'publishing' | 'game' | 'design';

export type SupportStatus = 'active' | 'refunded';

export interface Project {
  id: string;
  name: string;
  category: ProjectCategory;
  targetAmount: number;
  raisedAmount: number;
  deadline: string;
  description: string;
  status: ProjectStatus;
  createdAt: string;
  updatedAt: string;
  succeededAt?: string;
}

export interface RewardTier {
  id: string;
  projectId: string;
  amount: number;
  description: string;
  createdAt: string;
}

export interface SupportRecord {
  id: string;
  projectId: string;
  rewardTierId: string;
  supporterId: string;
  amount: number;
  status: SupportStatus;
  createdAt: string;
  refundedAt?: string;
}

export interface CreateProjectRequest {
  name: string;
  category: ProjectCategory;
  targetAmount: number;
  deadline: string;
  description: string;
  rewardTiers: Array<{
    amount: number;
    description: string;
  }>;
}

export interface UpdateStatusRequest {
  status: ProjectStatus;
}

export interface SupportRequest {
  rewardTierId: string;
  supporterId: string;
}

export interface RefundRequest {
  supportId: string;
}
