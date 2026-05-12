import express from 'express';
import servicesRoutes from './routes/services';
import { checkHeartbeats, HEARTBEAT_INTERVAL_MS } from './services/registryService';

const app = express();
const PORT = process.env.PORT ? parseInt(process.env.PORT, 10) : 3000;

app.use(express.json());
app.use(express.urlencoded({ extended: true }));

app.get('/health', (req, res) => {
  res.json({ status: 'ok' });
});

app.use('/services', servicesRoutes);

app.use((err: unknown, req: express.Request, res: express.Response, next: express.NextFunction) => {
  console.error(err);
  if (err instanceof Error && err.message.includes('Weight must be a non-negative integer')) {
    return res.status(400).json({ error: err.message });
  }
  res.status(500).json({ error: 'Internal server error' });
});

setInterval(() => {
  try {
    checkHeartbeats();
  } catch (error) {
    console.error('Heartbeat check failed:', error);
  }
}, HEARTBEAT_INTERVAL_MS);

app.listen(PORT, () => {
  console.log(`Service registry running on port ${PORT}`);
});
