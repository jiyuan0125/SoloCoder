import express from 'express';
import { initDatabase, getPolicyByName } from './database';
import routes from './routes';

const app = express();
const PORT = process.env.PORT ? parseInt(process.env.PORT, 10) : 3000;

app.use(express.json());

app.use('/api', routes);

app.get('/health', (req, res) => {
  res.json({ status: 'ok' });
});

initDatabase();

app.listen(PORT, () => {
  console.log(`Password policy service running on port ${PORT}`);
  const defaultPolicy = getPolicyByName('default');
  if (defaultPolicy) {
    console.log(`Default policy ID: ${defaultPolicy.id}`);
  }
});
