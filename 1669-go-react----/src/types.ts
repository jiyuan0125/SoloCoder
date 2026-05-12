export interface ChannelTransaction {
  transaction_id: string;
  transaction_time: string;
  amount: number;
  counterparty: string;
  channel_name: string;
  status: string;
}

export interface InternalOrder {
  order_id: string;
  order_time: string;
  amount: number;
  payment_method: string;
  status: string;
}

export type MatchResultType = 'matched' | 'channel_extra' | 'order_extra';

export interface MatchResult {
  type: MatchResultType;
  transaction_id?: string;
  order_id?: string;
}

export type DisputeAction = 'supplement' | 'suspend' | 'adjust';

export interface DisputeRequest {
  report_id: number;
  entity_type: 'channel' | 'order';
  entity_id: string;
  action: DisputeAction;
  supplement_order?: Partial<InternalOrder> & { amount: number };
}
