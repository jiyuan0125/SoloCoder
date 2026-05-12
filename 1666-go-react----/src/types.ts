export type Stage =
  | '初次接触'
  | '需求调研'
  | '方案演示'
  | '商务报价'
  | '合同谈判'
  | '赢单'
  | '输单';

export interface Salesperson {
  id: string;
  name: string;
  teamId: string;
  createdAt: string;
}

export interface Opportunity {
  id: string;
  name: string;
  customerId: string;
  customerName: string;
  amount: number;
  salespersonId: string;
  currentStage: Stage;
  createdAt: string;
  updatedAt: string;
  closedAt: string | null;
  lossReason: string | null;
}

export interface StageHistory {
  id: string;
  opportunityId: string;
  fromStage: Stage | null;
  toStage: Stage;
  reason: string | null;
  timestamp: string;
}

export interface FunnelData {
  stage: Stage;
  count: number;
  conversionRate: number;
  amount: number;
  predictedWinAmount: number;
}

export interface CreateSalespersonRequest {
  name: string;
  teamId: string;
}

export interface CreateOpportunityRequest {
  name: string;
  customerId: string;
  customerName: string;
  amount: number;
  salespersonId: string;
}

export interface UpdateStageRequest {
  newStage: Stage;
  reason?: string;
}

export interface FunnelQuery {
  startDate: string;
  endDate: string;
  salespersonId?: string;
  teamId?: string;
}
