import express, { Request, Response, NextFunction } from 'express';
import { initDb } from './db';
import { setupDailyScheduler, runDailyTasks } from './scheduler';
import { createInvestigation, getInvestigationById, getInvestigationsByPersonId } from './investigation';
import {
  createIsolation,
  getIsolationById,
  createHealthRecord,
  getHealthRecordsByIsolation,
} from './isolation';
import { createTestRecord, getTestRecordById, getTestRecordsByPersonId } from './test';
import {
  generateDischargeNotification,
  getDischargeNotificationByIsolation,
  confirmDischargeNotification,
} from './discharge';

const app = express();
app.use(express.json());

let db: Awaited<ReturnType<typeof initDb>>;

app.use((err: any, _req: Request, res: Response, _next: NextFunction) => {
  const status = err.status || 500;
  res.status(status).json({
    error: err.message || 'Internal Server Error',
  });
});

app.post('/api/investigations', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const result = await createInvestigation(db, req.body);
    res.status(201).json(result);
  } catch (err) {
    next(err);
  }
});

app.get('/api/investigations/:id', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const inv = await getInvestigationById(db, req.params.id);
    if (!inv) {
      res.status(404).json({ error: '排查记录不存在' });
      return;
    }
    res.json(inv);
  } catch (err) {
    next(err);
  }
});

app.get('/api/persons/:personId/investigations', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const list = await getInvestigationsByPersonId(db, req.params.personId);
    res.json(list);
  } catch (err) {
    next(err);
  }
});

app.post('/api/isolations', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const iso = await createIsolation(db, req.body);
    res.status(201).json(iso);
  } catch (err) {
    next(err);
  }
});

app.get('/api/isolations/:id', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const iso = await getIsolationById(db, req.params.id);
    if (!iso) {
      res.status(404).json({ error: '隔离记录不存在' });
      return;
    }
    res.json(iso);
  } catch (err) {
    next(err);
  }
});

app.post('/api/health-records', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const result = await createHealthRecord(db, req.body);
    res.status(201).json(result);
  } catch (err) {
    next(err);
  }
});

app.get('/api/isolations/:isolationId/health-records', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const list = await getHealthRecordsByIsolation(db, req.params.isolationId);
    res.json(list);
  } catch (err) {
    next(err);
  }
});

app.post('/api/tests', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const result = await createTestRecord(db, req.body);
    res.status(201).json(result);
  } catch (err) {
    next(err);
  }
});

app.get('/api/tests/:id', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const test = await getTestRecordById(db, req.params.id);
    if (!test) {
      res.status(404).json({ error: '检测记录不存在' });
      return;
    }
    res.json(test);
  } catch (err) {
    next(err);
  }
});

app.get('/api/persons/:personId/tests', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const list = await getTestRecordsByPersonId(db, req.params.personId);
    res.json(list);
  } catch (err) {
    next(err);
  }
});

app.post('/api/isolations/:isolationId/generate-discharge', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const notification = await generateDischargeNotification(db, req.params.isolationId);
    res.status(201).json(notification);
  } catch (err) {
    next(err);
  }
});

app.get('/api/isolations/:isolationId/discharge-notification', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const notification = await getDischargeNotificationByIsolation(db, req.params.isolationId);
    if (!notification) {
      res.status(404).json({ error: '解除通知不存在' });
      return;
    }
    res.json(notification);
  } catch (err) {
    next(err);
  }
});

app.post('/api/discharge-notifications/:id/confirm', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const result = await confirmDischargeNotification(db, req.params.id, req.body);
    res.json(result);
  } catch (err) {
    next(err);
  }
});

app.post('/api/scheduler/run-now', async (req: Request, res: Response, next: NextFunction) => {
  try {
    const result = await runDailyTasks(db);
    res.json(result);
  } catch (err) {
    next(err);
  }
});

const PORT = process.env.PORT || 8300;

async function start() {
  db = await initDb();
  setupDailyScheduler(db);

  app.listen(PORT, () => {
    console.log(`Epidemic Management System running on port ${PORT}`);
    console.log('Daily scheduler active at 08:00 Asia/Shanghai');
  });
}

start().catch((err) => {
  console.error('Failed to start server:', err);
  process.exit(1);
});
