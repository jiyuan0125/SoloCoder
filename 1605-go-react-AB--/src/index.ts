import express from 'express';
import experimentsRouter from './routes/experiments';
import assignRouter from './routes/assign';
import metricsRouter from './routes/metrics';
import grayscaleRouter from './routes/grayscale';

const app = express();
const PORT = process.env.PORT || 3000;

app.use(express.json());

app.use('/api/experiments', experimentsRouter);
app.use('/api/assign', assignRouter);
app.use('/api/metrics', metricsRouter);
app.use('/api/grayscale', grayscaleRouter);

app.get('/health', (_req, res) => {
  res.json({ status: 'ok', timestamp: new Date().toISOString() });
});

app.listen(PORT, () => {
  console.log(`AB Testing Platform running on port ${PORT}`);
});
