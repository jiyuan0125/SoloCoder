export type LeadStatus = 'new' | 'contacted' | 'rejected' | 'converted';

export type OpportunityStage = 
  | 'initial_contact' 
  | 'needs_confirmation' 
  | 'proposal_quotation' 
  | 'business_negotiation' 
  | 'won' 
  | 'lost';

export interface Lead {
  id: string;
  source_channel: string;
  contact_info: string;
  company_name?: string;
  contact_person?: string;
  requirements?: string;
  status: LeadStatus;
  created_at: number;
  updated_at: number;
  assigned_sales_id?: string;
  is_in_pool: number;
  version: number;
}

export interface Opportunity {
  id: string;
  lead_id: string;
  expected_amount: number;
  assigned_sales_id: string;
  stage: OpportunityStage;
  expected_close_date?: string;
  status: 'active' | 'won' | 'lost';
  created_at: number;
  updated_at: number;
}

export interface FunnelStat {
  stage: OpportunityStage;
  stage_name: string;
  opportunity_count: number;
  total_amount: number;
  conversion_rate: number;
}
