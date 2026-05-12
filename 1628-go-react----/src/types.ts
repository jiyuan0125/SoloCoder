export interface Service {
  id: string;
  name: string;
  upstream_url: string;
  route_prefix: string;
  created_at: number;
  updated_at: number;
}

export interface Client {
  id: string;
  name: string;
  description?: string;
  created_at: number;
}

export interface ApiKey {
  id: string;
  client_id: string;
  key_hash: string;
  key_prefix: string;
  expires_at?: number;
  permissions: string[];
  is_active: number;
  created_at: number;
}

export interface ApiKeyWithPlain extends ApiKey {
  plain_key?: string;
}

export interface KeyStats {
  id: string;
  key_id: string;
  total_requests: number;
  successful_requests: number;
  failed_requests: number;
  last_request_at?: number;
}

export interface RateLimitConfig {
  perSecond?: number;
  perMinute?: number;
  perHour?: number;
}

export interface RateLimitDimension {
  ip?: RateLimitConfig;
  userId?: RateLimitConfig;
  path?: Record<string, RateLimitConfig>;
}

export interface MatchedService {
  service: Service;
  matchedPrefix: string;
  remainingPath: string;
}

declare global {
  namespace Express {
    interface Request {
      matchedService?: MatchedService;
      apiKey?: ApiKey;
      clientIp?: string;
    }
  }
}
