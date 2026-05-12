import express from 'express';
import { initDatabase } from './database';
import campaignsRouter from './routes/campaigns';
import biddingRouter from './routes/bidding';
import reportsRouter from './routes/reports';

const app = express();
const PORT = process.env.PORT ? parseInt(process.env.PORT) : 3000;

app.use(express.json());

app.use('/api/campaigns', campaignsRouter);
app.use('/api', biddingRouter);
app.use('/api/reports', reportsRouter);

app.get('/health', (req, res) => {
  res.json({ status: 'ok' });
});

async function start() {
  await initDatabase();
  app.listen(PORT, () => {
    console.log(`DSP Platform server running on port ${PORT}`);
  });
}

start().catch((err) => {
  console.error('Failed to start server:', err);
  process.exit(1);
});
