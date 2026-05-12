export interface Package {
  id: number;
  name: string;
  max_users: number;
  max_storage: number;
  max_api_calls: number;
  created_at: string;
  updated_at: string;
}

export interface Tenant {
  id: number;
  enterprise_name: string;
  contact_person: string;
  contact_phone: string;
  package_id: number;
  created_at: string;
  updated_at: string;
}

export interface Quota {
  max_users: number;
  max_storage: number;
  max_api_calls: number;
}

export interface TenantQuota {
  id: number;
  tenant_id: number;
  max_users: number;
  max_storage: number;
  max_api_calls: number;
  is_temporary: number;
  expires_at: string | null;
  reason: string | null;
  created_at: string;
}

export interface Usage {
  id: number;
  tenant_id: number;
  user_count: number;
  storage_used: number;
  api_calls_this_month: number;
  last_api_reset_at: string;
  updated_at: string;
}

export interface Alert {
  id: number;
  tenant_id: number;
  resource_type: string;
  usage_percentage: number;
  message: string;
  handled: number;
  created_at: string;
}

export interface Subscription {
  id: number;
  tenant_id: number;
  status: string;
  start_date: string;
  end_date: string | null;
  created_at: string;
}
