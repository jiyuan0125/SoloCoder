export type InstanceStatus = 'healthy' | 'unhealthy' | 'deregistered';

export interface ServiceInstance {
  id: string;
  serviceName: string;
  address: string;
  port: number;
  weight: number;
  metadata: Record<string, string>;
  status: InstanceStatus;
  registeredAt: number;
  lastHeartbeat: number;
  failedHeartbeats: number;
}

export interface HeartbeatRecord {
  id: number;
  instanceId: string;
  timestamp: number;
  success: boolean;
}

export interface RegisterRequest {
  serviceName: string;
  address: string;
  port: number;
  weight?: number;
  metadata?: Record<string, string>;
}

export interface DiscoverResponse {
  instances: ServiceInstance[];
  selected?: ServiceInstance;
}
