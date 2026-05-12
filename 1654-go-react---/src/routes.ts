import { Router, Request, Response } from 'express';
import { SessionStore } from './sessionStore';
import { AuditNotifier } from './auditNotifier';
import { Session, SessionCreateRequest, SessionFilter } from './types';

export function createRoutes(
  sessionStore: SessionStore,
  auditNotifier: AuditNotifier
): Router {
  const router = Router();

  const serializeSession = (session: Session) => ({
    ...session,
    loginTime: session.loginTime.toISOString(),
    lastHeartbeatTime: session.lastHeartbeatTime.toISOString(),
  });

  router.post('/api/sessions/login', (req: Request, res: Response) => {
    const body = req.body as Partial<SessionCreateRequest>;

    if (
      !body.userId ||
      !body.deviceInfo ||
      !body.ipAddress ||
      !body.packageType
    ) {
      res.status(400).json({
        error: 'Missing required fields',
        required: ['userId', 'deviceInfo', 'ipAddress', 'packageType'],
      });
      return;
    }

    const session = sessionStore.createSession(body as SessionCreateRequest);
    res.status(200).json(serializeSession(session));
  });

  router.post('/api/sessions/:sessionId/heartbeat', (req: Request, res: Response) => {
    const { sessionId } = req.params;

    const success = sessionStore.heartbeat(sessionId);
    if (!success) {
      res.status(404).json({ error: 'Session not found or not online' });
      return;
    }

    const session = sessionStore.getSession(sessionId);
    res.status(200).json({
      sessionId,
      lastHeartbeatTime: session?.lastHeartbeatTime.toISOString(),
    });
  });

  router.get('/api/admin/sessions', (req: Request, res: Response) => {
    const filter: SessionFilter = {};

    if (typeof req.query.userId === 'string' && req.query.userId) {
      filter.userId = req.query.userId;
    }
    if (typeof req.query.ipAddress === 'string' && req.query.ipAddress) {
      filter.ipAddress = req.query.ipAddress;
    }

    const sessions = sessionStore.listOnlineSessions(filter);
    res.status(200).json(sessions.map(serializeSession));
  });

  router.post('/api/admin/sessions/:sessionId/kick', async (req: Request, res: Response) => {
    const { sessionId } = req.params;
    const reason = typeof req.body.reason === 'string' ? req.body.reason : undefined;

    const session = sessionStore.getSession(sessionId);

    if (!session) {
      res.status(404).json({ error: 'Session not found' });
      return;
    }

    const result = sessionStore.kickSession(sessionId);

    if (!result.success) {
      if (result.error === 'not_found') {
        res.status(404).json({ error: 'Session not found' });
      } else {
        res.status(409).json({ error: '会话已被其他管理员操作' });
      }
      return;
    }

    const updatedSession = sessionStore.getSession(sessionId);
    if (!updatedSession) {
      res.status(500).json({ error: 'Failed to update session' });
      return;
    }

    try {
      await auditNotifier.sendNotification({
        sessionId,
        userId: updatedSession.userId,
        action: 'kick',
        timestamp: new Date(),
        reason,
      });

      res.status(200).json({
        success: true,
        session: serializeSession(updatedSession),
      });
    } catch (auditError) {
      res.status(500).json({
        success: true,
        session: serializeSession(updatedSession),
        auditError: auditError instanceof Error ? auditError.message : 'Unknown error',
        message: 'Kick operation succeeded, but audit notification failed',
      });
    }
  });

  router.get('/api/admin/statistics', (req: Request, res: Response) => {
    const stats = sessionStore.getStatistics();
    res.status(200).json({
      currentOnlineCount: stats.currentOnlineCount,
      todayLoginCount: stats.todayLoginCount,
      averageSessionDurationMs: stats.averageSessionDuration,
      averageSessionDurationFormatted: formatDuration(stats.averageSessionDuration),
    });
  });

  router.get('/api/health', (req: Request, res: Response) => {
    res.status(200).json({ status: 'ok' });
  });

  return router;
}

function formatDuration(ms: number): string {
  if (ms === 0) return '0 seconds';

  const seconds = Math.floor(ms / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);

  if (hours > 0) {
    return `${hours}h ${minutes % 60}m ${seconds % 60}s`;
  }
  if (minutes > 0) {
    return `${minutes}m ${seconds % 60}s`;
  }
  return `${seconds}s`;
}
