import express from 'express';
import dataCategoriesRouter from './routes/dataCategories';
import requestsRouter from './routes/requests';
import consentsRouter from './routes/consents';
import { startScheduler } from './services/scheduler';

const app = express();
const PORT = process.env.PORT ? parseInt(process.env.PORT, 10) : 3000;

app.use(express.json());

app.use('/data-categories', dataCategoriesRouter);
app.use('/requests', requestsRouter);
app.use('/consents', consentsRouter);

app.get('/health', (req, res) => {
  res.json({ status: 'ok', timestamp: new Date().toISOString() });
});

app.use((err: Error, req: express.Request, res: express.Response, _next: express.NextFunction) => {
  console.error('Unhandled error:', err);
  res.status(500).json({ error: 'Internal server error' });
});

app.listen(PORT, () => {
  console.log(`GDPR Compliance Service listening on port ${PORT}`);
  startScheduler();
});
