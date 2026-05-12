import express from 'express';
import routes from './routes';

const app = express();
const PORT = process.env.PORT || 9111;

app.use(express.json());

app.get('/health', (req, res) => {
  res.json({ status: 'ok', service: 'backup-restore-service' });
});

app.use('/api', routes);

app.use((req, res) => {
  res.status(404).json({ error: 'Not found' });
});

app.use((err: any, req: express.Request, res: express.Response, next: express.NextFunction) => {
  console.error('Unhandled error:', err);
  res.status(500).json({ error: 'Internal server error' });
});

app.listen(PORT, () => {
  console.log(`Backup and Restore Service running on port ${PORT}`);
  console.log(`API available at http://localhost:${PORT}/api`);
  console.log(`Health check: http://localhost:${PORT}/health`);
});
