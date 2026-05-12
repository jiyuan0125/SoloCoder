export enum SessionStatus {
  ONLINE = 'online',
  OFFLINE = 'offline',
  KICKED = 'kicked',
  REPLACED = 'replaced',
}

export interface Session {
  sessionId: string;
  userId: string;
  deviceInfo: string;
  ipAddress: string;
  loginTime: Date;
  lastHeartbeatTime: Date;
  status: SessionStatus;
  packageType: string;
}

export interface SessionCreateRequest {
  userId: string;
  deviceInfo: string;
  ipAddress: string;
  packageType: string;
}

export interface SessionFilter {
  userId?: string;
  ipAddress?: string;
}

export interface SessionStatistics {
  currentOnlineCount: number;
  todayLoginCount: number;
  averageSessionDuration: number;
}

export interface AuditNotification {
  sessionId: string;
  userId: string;
  action: string;
  timestamp: Date;
  reason?: string;
}

export type KickResult =
  | { success: true; sessionId: string }
  | { success: false; error: 'not_found' | 'already_kicked' | 'already_replaced' };

export interface PackageLimit {
  [key: string]: number;
}
