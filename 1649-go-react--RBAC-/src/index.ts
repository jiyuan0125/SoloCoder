import express from 'express';
import permissionsRouter from './routes/permissions';
import rolesRouter from './routes/roles';
import usersRouter from './routes/users';
import auditLogsRouter from './routes/audit-logs';
import './database';

const app = express();
const PORT = process.env.PORT ? parseInt(process.env.PORT) : 3000;

app.use(express.json());

app.get('/health', (req, res) => {
  res.json({ status: 'ok', service: 'rbac-service' });
});

app.use('/permissions', permissionsRouter);
app.use('/roles', rolesRouter);
app.use('/users', usersRouter);
app.use('/audit-logs', auditLogsRouter);

app.use((req, res) => {
  res.status(404).json({ error: 'Not Found', path: req.path });
});

app.use((err: Error, req: express.Request, res: express.Response, next: express.NextFunction) => {
  console.error(err.stack);
  res.status(500).json({ error: 'Internal Server Error' });
});

app.listen(PORT, () => {
  console.log(`RBAC service running on port ${PORT}`);
});

export default app;
