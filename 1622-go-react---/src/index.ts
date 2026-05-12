import express from 'express';
import router from './routes';
import { closeDatabase } from './database';

const app = express();
const PORT = process.env.PORT || 9112;

app.use(express.json());
app.use('/', router);

app.use((err: Error, req: express.Request, res: express.Response, next: express.NextFunction) => {
  console.error(err.stack);
  res.status(500).json({ error: 'Internal server error' });
});

const server = app.listen(PORT, () => {
  console.log(`Log service running on port ${PORT}`);
});

process.on('SIGTERM', () => {
  console.log('SIGTERM received, shutting down...');
  closeDatabase();
  server.close(() => {
    process.exit(0);
  });
});

process.on('SIGINT', () => {
  console.log('SIGINT received, shutting down...');
  closeDatabase();
  server.close(() => {
    process.exit(0);
  });
});
