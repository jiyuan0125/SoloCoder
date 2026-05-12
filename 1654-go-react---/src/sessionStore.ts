import { v4 as uuidv4 } from 'uuid';
import {
  Session,
  SessionStatus,
  SessionCreateRequest,
  SessionFilter,
  SessionStatistics,
  PackageLimit,
  KickResult,
} from './types';

export interface SessionStoreConfig {
  heartbeatTimeoutMs: number;
  sessionRetentionDays: number;
  packageLimits: PackageLimit;
}

const DEFAULT_CONFIG: SessionStoreConfig = {
  heartbeatTimeoutMs: 90 * 1000,
  sessionRetentionDays: 30,
  packageLimits: {
    free: 1,
    basic: 3,
    pro: 10,
    enterprise: 100,
  },
};

export class SessionStore {
  private sessions: Map<string, Session> = new Map();
  private config: SessionStoreConfig;
  private kickedSessions: Set<string> = new Set();
  private cleanupInterval: NodeJS.Timeout | null = null;

  constructor(config?: Partial<SessionStoreConfig>) {
    this.config = { ...DEFAULT_CONFIG, ...config };
  }

  startCleanupTimer(): void {
    if (this.cleanupInterval) return;
    this.cleanupInterval = setInterval(() => {
      this.cleanupOldSessions();
      this.checkHeartbeatTimeout();
    }, 10 * 1000);
  }

  stopCleanupTimer(): void {
    if (this.cleanupInterval) {
      clearInterval(this.cleanupInterval);
      this.cleanupInterval = null;
    }
  }

  createSession(request: SessionCreateRequest): Session {
    const existingSession = this.findOnlineSessionByUserAndDevice(
      request.userId,
      request.deviceInfo
    );

    if (existingSession) {
      existingSession.lastHeartbeatTime = new Date();
      existingSession.ipAddress = request.ipAddress;
      return existingSession;
    }

    this.enforcePackageLimit(request.packageType, request.userId);

    const now = new Date();
    const session: Session = {
      sessionId: uuidv4(),
      userId: request.userId,
      deviceInfo: request.deviceInfo,
      ipAddress: request.ipAddress,
      loginTime: now,
      lastHeartbeatTime: now,
      status: SessionStatus.ONLINE,
      packageType: request.packageType,
    };

    this.sessions.set(session.sessionId, session);
    return session;
  }

  heartbeat(sessionId: string): boolean {
    const session = this.sessions.get(sessionId);
    if (!session) return false;

    if (session.status === SessionStatus.ONLINE) {
      session.lastHeartbeatTime = new Date();
      return true;
    }

    return false;
  }

  getSession(sessionId: string): Session | undefined {
    return this.sessions.get(sessionId);
  }

  listOnlineSessions(filter?: SessionFilter): Session[] {
    const sessions = Array.from(this.sessions.values()).filter(
      (s) => s.status === SessionStatus.ONLINE
    );

    if (!filter) return sessions;

    return sessions.filter((s) => {
      if (filter.userId && s.userId !== filter.userId) return false;
      if (filter.ipAddress && s.ipAddress !== filter.ipAddress) return false;
      return true;
    });
  }

  kickSession(sessionId: string): KickResult {
    const session = this.sessions.get(sessionId);

    if (!session) {
      return { success: false, error: 'not_found' };
    }

    if (session.status === SessionStatus.KICKED) {
      return { success: false, error: 'already_kicked' };
    }

    if (session.status === SessionStatus.REPLACED) {
      return { success: false, error: 'already_replaced' };
    }

    if (this.kickedSessions.has(sessionId)) {
      return { success: false, error: 'already_kicked' };
    }

    this.kickedSessions.add(sessionId);
    session.status = SessionStatus.KICKED;

    return { success: true, sessionId };
  }

  getStatistics(): SessionStatistics {
    const now = new Date();
    const todayStart = new Date(
      now.getFullYear(),
      now.getMonth(),
      now.getDate(),
      0,
      0,
      0,
      0
    );

    const allSessions = Array.from(this.sessions.values());

    const currentOnlineCount = allSessions.filter(
      (s) => s.status === SessionStatus.ONLINE
    ).length;

    const todayLoginCount = allSessions.filter(
      (s) => s.loginTime >= todayStart
    ).length;

    const completedSessions = allSessions.filter(
      (s) =>
        s.status === SessionStatus.OFFLINE ||
        s.status === SessionStatus.KICKED ||
        s.status === SessionStatus.REPLACED
    );

    let averageSessionDuration = 0;
    if (completedSessions.length > 0) {
      const totalDuration = completedSessions.reduce((sum, s) => {
        return (
          sum + (s.lastHeartbeatTime.getTime() - s.loginTime.getTime())
        );
      }, 0);
      averageSessionDuration = totalDuration / completedSessions.length;
    }

    return {
      currentOnlineCount,
      todayLoginCount,
      averageSessionDuration,
    };
  }

  private findOnlineSessionByUserAndDevice(
    userId: string,
    deviceInfo: string
  ): Session | undefined {
    return Array.from(this.sessions.values()).find(
      (s) =>
        s.userId === userId &&
        s.deviceInfo === deviceInfo &&
        s.status === SessionStatus.ONLINE
    );
  }

  private enforcePackageLimit(
    packageType: string,
    excludeUserId: string
  ): void {
    const limit = this.config.packageLimits[packageType] || 1;
    const currentCount = Array.from(this.sessions.values()).filter(
      (s) =>
        s.packageType === packageType &&
        s.status === SessionStatus.ONLINE
    ).length;

    if (currentCount < limit) return;

    const sessionsToKick = Array.from(this.sessions.values())
      .filter(
        (s) =>
          s.packageType === packageType &&
          s.status === SessionStatus.ONLINE
      )
      .sort((a, b) => a.loginTime.getTime() - b.loginTime.getTime());

    const kickCount = currentCount - limit + 1;

    for (let i = 0; i < kickCount && i < sessionsToKick.length; i++) {
      const session = sessionsToKick[i];
      session.status = SessionStatus.REPLACED;
    }
  }

  private checkHeartbeatTimeout(): void {
    const now = Date.now();
    const timeout = this.config.heartbeatTimeoutMs;

    for (const session of this.sessions.values()) {
      if (session.status !== SessionStatus.ONLINE) continue;

      const timeSinceHeartbeat = now - session.lastHeartbeatTime.getTime();
      if (timeSinceHeartbeat > timeout) {
        session.status = SessionStatus.OFFLINE;
      }
    }
  }

  private cleanupOldSessions(): void {
    const retentionMs = this.config.sessionRetentionDays * 24 * 60 * 60 * 1000;
    const cutoff = Date.now() - retentionMs;

    for (const [sessionId, session] of this.sessions.entries()) {
      if (
        session.status !== SessionStatus.ONLINE &&
        session.lastHeartbeatTime.getTime() < cutoff
      ) {
        this.sessions.delete(sessionId);
        this.kickedSessions.delete(sessionId);
      }
    }
  }
}
