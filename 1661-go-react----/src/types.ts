export type AdStatus = 'draft' | 'running' | 'paused' | 'ended';

export type BidType = 'CPC' | 'CPM';

export interface Targeting {
  ageRanges?: string[];
  gender?: 'male' | 'female' | 'all';
  regions?: string[];
}

export interface AdPlan {
  id: string;
  name: string;
  budget: number;
  dailyBudget: number;
  targeting: Targeting;
  bidType: BidType;
  bidAmount: number;
  status: AdStatus;
  expectedCtr: number;
  createdAt: number;
  updatedAt: number;
}

export interface HourlyStats {
  id: string;
  planId: string;
  hour: number;
  impressions: number;
  clicks: number;
  conversions: number;
  spend: number;
  createdAt: number;
}

export interface LedgerRecord {
  id: string;
  planId: string;
  amount: number;
  type: 'deduction' | 'refund' | 'overcharge';
  timestamp: number;
  relatedId?: string;
}
