export type SettlementMethod = 'monthly' | 'threshold';
export type ChannelStatus = 'active' | 'inactive';
export type MaterialType = 'homepage' | 'campaign' | 'product';
export type OrderStatus = 'pending' | 'completed' | 'refunded';
export type CommissionStatus = 'pending' | 'settled' | 'cancelled';

export interface Channel {
  id: string;
  name: string;
  contact: string;
  settlement_method: SettlementMethod;
  commission_rate: number;
  status: ChannelStatus;
  created_at: string;
  updated_at: string;
}

export interface Link {
  id: string;
  channel_id: string;
  material_type: MaterialType;
  unique_code: string;
  created_at: string;
}

export interface Attribution {
  id: string;
  user_id: string;
  channel_id: string;
  link_id: string;
  click_time: string;
  attribution_window_end: string;
  created_at: string;
}

export interface Order {
  id: string;
  user_id: string;
  amount: number;
  status: OrderStatus;
  created_at: string;
  updated_at: string;
}

export interface Commission {
  id: string;
  order_id: string;
  channel_id: string;
  amount: number;
  status: CommissionStatus;
  settlement_month: string | null;
  created_at: string;
  updated_at: string;
}

export interface SettlementStats {
  channel_id: string;
  channel_name: string;
  settlement_month: string;
  total_orders: number;
  total_commission: number;
  status: CommissionStatus;
}

export interface CreateChannelRequest {
  name: string;
  contact: string;
  settlement_method: SettlementMethod;
  commission_rate: number;
  status?: ChannelStatus;
}

export interface UpdateChannelRequest {
  name?: string;
  contact?: string;
  settlement_method?: SettlementMethod;
  commission_rate?: number;
  status?: ChannelStatus;
}

export interface CreateLinkRequest {
  channel_id: string;
  material_type: MaterialType;
}

export interface TrackAttributionRequest {
  user_id: string;
  link_id: string;
}

export interface CreateOrderRequest {
  user_id: string;
  amount: number;
}

export interface RefundOrderRequest {
  order_id: string;
}
