import { LeadStatus, OpportunityStage } from './types';

export const LEAD_STATUS_TRANSITIONS: Record<LeadStatus, LeadStatus[]> = {
  new: ['contacted'],
  contacted: ['rejected', 'converted'],
  rejected: [],
  converted: [],
};

export const VALID_LEAD_STATUSES: LeadStatus[] = ['new', 'contacted', 'rejected', 'converted'];

export const STAGE_ORDER: OpportunityStage[] = [
  'initial_contact',
  'needs_confirmation',
  'proposal_quotation',
  'business_negotiation',
  'won',
  'lost',
];

export const STAGE_NAMES: Record<OpportunityStage, string> = {
  initial_contact: '初步接触',
  needs_confirmation: '需求确认',
  proposal_quotation: '方案报价',
  business_negotiation: '商务谈判',
  won: '赢单',
  lost: '输单',
};

export const TERMINAL_STAGES: OpportunityStage[] = ['won', 'lost'];

export const SEVEN_DAYS_MS = 7 * 24 * 60 * 60 * 1000;
export const THIRTY_DAYS_MS = 30 * 24 * 60 * 60 * 1000;
