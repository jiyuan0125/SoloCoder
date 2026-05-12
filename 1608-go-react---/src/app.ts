import express, { Application } from 'express';
import { initDb } from './db';
import productsRouter from './routes/products';
import ratingsRouter from './routes/ratings';
import recommendRouter from './routes/recommend';

const app: Application = express();
const PORT = process.env.PORT || 3000;

app.use(express.json());

app.use('/api/products', productsRouter);
app.use('/api/ratings', ratingsRouter);
app.use('/api/recommend', recommendRouter);

app.get('/health', (_req, res) => {
  res.json({ status: 'ok' });
});

async function start() {
  await initDb();
  app.listen(PORT, () => {
    console.log(`Recommendation engine running on port ${PORT}`);
  });
}

start().catch(err => {
  console.error('Failed to start server:', err);
  process.exit(1);
});

export default app;
