export enum CampaignStatus {
  ACTIVE = 'active',
  PAUSED = 'paused',
  ENDED = 'ended'
}

export enum BiddingStrategy {
  CPM = 'CPM',
  CPC = 'CPC'
}

export interface Targeting {
  age?: number[];
  gender?: string[];
  location?: string[];
  interests?: string[];
}

export interface UserProfile {
  age?: number;
  gender?: string;
  location?: string;
  interests?: string[];
}

export interface Campaign {
  id: string;
  name: string;
  dailyBudget: number;
  totalBudget: number;
  targeting: Targeting;
  biddingStrategy: BiddingStrategy;
  bidAmount: number;
  status: CampaignStatus;
  dailySpent: number;
  totalSpent: number;
  createdAt: string;
  updatedAt: string;
}

export interface CampaignCreateRequest {
  name: string;
  dailyBudget: number;
  totalBudget: number;
  targeting: Targeting;
  biddingStrategy: BiddingStrategy;
  bidAmount: number;
}

export interface CampaignUpdateRequest {
  name?: string;
  dailyBudget?: number;
  totalBudget?: number;
  targeting?: Targeting;
  biddingStrategy?: BiddingStrategy;
  bidAmount?: number;
  status?: CampaignStatus;
}

export interface BidRequest {
  requestId: string;
  userProfile: UserProfile;
  impressionType: BiddingStrategy;
}

export interface BidParticipant {
  campaignId: string;
  campaignName: string;
  bidAmount: number;
}

export enum BidLogStatus {
  SUCCESS = 'success',
  VOIDED = 'voided'
}

export interface BidLog {
  id: string;
  requestId: string;
  participants: string;
  winnerCampaignId: string | null;
  winnerCampaignName: string | null;
  finalPrice: number;
  status: BidLogStatus;
  createdAt: string;
}

export interface CampaignStatistics {
  campaignId: string;
  campaignName: string;
  bidCount: number;
  winCount: number;
  winRate: number;
  totalSpent: number;
}
