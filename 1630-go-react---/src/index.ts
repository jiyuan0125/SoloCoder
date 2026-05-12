import express, { Request, Response } from 'express';
import { initDb } from './db';
import servicesRouter from './routes/services';
import metricsRouter from './routes/metrics';
import rulesRouter from './routes/rules';
import { cleanupOldMetrics } from './alerts';

const app = express();
const PORT = process.env.PORT || 3000;

app.use(express.json());

app.get('/health', (_req: Request, res: Response) => {
  res.json({ status: 'ok' });
});

app.use('/services', servicesRouter);
app.use('/', metricsRouter);
app.use('/rules', rulesRouter);

async function main() {
  await initDb();

  setInterval(() => {
    cleanupOldMetrics().catch(console.error);
  }, 60 * 1000);

  app.listen(PORT, () => {
    console.log(`Server running on port ${PORT}`);
  });
}

main().catch(console.error);
