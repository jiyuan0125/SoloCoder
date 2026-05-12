import express, { Request, Response, NextFunction } from 'express';
import vaccinesRouter from './routes/vaccines';
import reservationsRouter from './routes/reservations';
import vaccinationsRouter from './routes/vaccinations';
import { getDb } from './database';
import { Notification } from './types';

const app = express();
const PORT = process.env.PORT || 3000;

app.use(express.json());

app.use('/api/vaccines', vaccinesRouter);
app.use('/api/reservations', reservationsRouter);
app.use('/api/vaccinations', vaccinationsRouter);

app.get('/api/notifications', async (_req: Request, res: Response) => {
  const db = await getDb();
  const rows = await db.all('SELECT * FROM notifications ORDER BY created_at DESC');
  
  const notifications: Notification[] = rows.map((row: any) => ({
    id: row.id,
    type: row.type,
    message: row.message,
    targetId: row.target_id,
    createdAt: row.created_at,
    isRead: !!row.is_read
  }));
  
  res.json(notifications);
});

app.get('/api/health', (_req: Request, res: Response) => {
  res.json({ status: 'ok', timestamp: new Date().toISOString() });
});

app.use((err: Error, _req: Request, res: Response, _next: NextFunction) => {
  console.error(err.stack);
  res.status(500).json({ error: 'Internal server error' });
});

app.use((_req: Request, res: Response) => {
  res.status(404).json({ error: 'Not found' });
});

app.listen(PORT, async () => {
  await getDb();
  console.log(`Vaccine Management System running on http://localhost:${PORT}`);
  console.log(`API endpoints:`);
  console.log(`  GET  /api/vaccines - List all vaccine batches`);
  console.log(`  POST /api/vaccines - Create vaccine batch`);
  console.log(`  GET  /api/vaccines/:batchNumber - Get vaccine batch`);
  console.log(`  PUT  /api/vaccines/:batchNumber/stock - Update stock`);
  console.log(``);
  console.log(`  GET  /api/reservations - List all reservations`);
  console.log(`  POST /api/reservations - Create reservation`);
  console.log(`  GET  /api/reservations/:id - Get reservation`);
  console.log(`  DELETE /api/reservations/:id - Cancel reservation`);
  console.log(``);
  console.log(`  GET  /api/vaccinations - List all vaccination records`);
  console.log(`  POST /api/vaccinations - Create vaccination record`);
  console.log(`  GET  /api/vaccinations/:recordId - Get vaccination record`);
  console.log(`  GET  /api/vaccinations/:recordId/reactions - List reactions`);
  console.log(`  GET  /api/vaccinations/:recordId/reactions/:reactionId - Get reaction`);
  console.log(`  POST /api/vaccinations/:recordId/reactions/:reactionId/report - Report adverse reaction`);
  console.log(``);
  console.log(`  GET  /api/notifications - List notifications`);
});
