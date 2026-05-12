import express from 'express';
import { SessionStore } from './sessionStore';
import { AuditNotifier } from './auditNotifier';
import { createRoutes } from './routes';

const PORT = parseInt(process.env.PORT || '3000', 10);

const sessionStore = new SessionStore();
const auditNotifier = new AuditNotifier();
const app = express();

app.use(express.json());
app.use(createRoutes(sessionStore, auditNotifier));

function startServer(): void {
  sessionStore.startCleanupTimer();

  const server = app.listen(PORT, () => {
    console.log(`Session Management Service running on port ${PORT}`);
    console.log(`API endpoints:`);
    console.log(`  POST /api/sessions/login - Create/get session`);
    console.log(`  POST /api/sessions/:sessionId/heartbeat - Heartbeat`);
    console.log(`  GET  /api/admin/sessions - List online sessions`);
    console.log(`  POST /api/admin/sessions/:sessionId/kick - Kick session`);
    console.log(`  GET  /api/admin/statistics - Get statistics`);
    console.log(`  GET  /api/health - Health check`);
  });

  function shutdown(): void {
    console.log('\nShutting down...');
    sessionStore.stopCleanupTimer();
    server.close(() => {
      console.log('Server stopped.');
      process.exit(0);
    });
  }

  process.on('SIGINT', shutdown);
  process.on('SIGTERM', shutdown);
}

if (require.main === module) {
  startServer();
}

export { app, sessionStore, auditNotifier };
