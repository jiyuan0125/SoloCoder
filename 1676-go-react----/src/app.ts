import express from 'express';
import projectsRouter from './routes/projects';

const app = express();
const PORT = process.env.PORT ? parseInt(process.env.PORT, 10) : 8205;

app.use(express.json());

app.get('/health', (_req, res) => {
  res.json({ status: 'ok', timestamp: new Date().toISOString() });
});

app.use('/api/projects', projectsRouter);

app.use((_req, res) => {
  res.status(404).json({ error: 'Not Found' });
});

app.use((err: any, _req: express.Request, res: express.Response, _next: express.NextFunction) => {
  console.error(err);
  res.status(500).json({ error: 'Internal Server Error' });
});

app.listen(PORT, () => {
  console.log(`Crowdfunding API server is running on http://localhost:${PORT}`);
});

export default app;
