import express from 'express';
import { initDatabase } from './database';
import circuitBreakerRoutes from './routes/circuitBreakers';

const app = express();
const PORT = process.env.PORT || 3000;

initDatabase();

app.use(express.json());

app.get('/health', (req, res) => {
  res.json({ status: 'ok', timestamp: new Date().toISOString() });
});

app.use('/circuit-breakers', circuitBreakerRoutes);

app.use((req, res) => {
  res.status(404).json({ error: 'Not Found' });
});

app.listen(PORT, () => {
  console.log(`Circuit Breaker Service running on port ${PORT}`);
});

export default app;
