import express, { Application } from 'express';
import cron from 'node-cron';
import {
  postSpansBatch,
  getTraceById,
  searchTraces,
  getSpansByTraceId,
  getSpanById,
  getDependencies,
} from './controllers';
import { cleanupOldData } from './database';

const app: Application = express();
const PORT = process.env.PORT ? parseInt(process.env.PORT, 10) : 3000;

app.use(express.json({ limit: '10mb' }));

app.post('/spans/batch', postSpansBatch);
app.get('/traces/search', searchTraces);
app.get('/traces/:traceId', getTraceById);
app.get('/traces/:traceId/spans', getSpansByTraceId);
app.get('/traces/:traceId/spans/:spanId', getSpanById);
app.get('/dependencies', getDependencies);

app.get('/', (req, res) => {
  res.json({
    service: 'Jaeger Clone - Distributed Tracing Service',
    version: '1.0.0',
    endpoints: {
      postSpansBatch: 'POST /spans/batch',
      getTraceById: 'GET /traces/{traceId}',
      searchTraces: 'GET /traces/search',
      getSpansByTraceId: 'GET /traces/{traceId}/spans',
      getSpanById: 'GET /traces/{traceId}/spans/{spanId}',
      getDependencies: 'GET /dependencies',
    },
  });
});

cron.schedule('0 */4 * * *', () => {
  console.log('[Cron] Running cleanup job...');
  try {
    cleanupOldData(7);
    console.log('[Cron] Cleanup completed');
  } catch (err) {
    console.error('[Cron] Cleanup error:', err);
  }
});

app.listen(PORT, () => {
  console.log(`Server is running on port ${PORT}`);
});
