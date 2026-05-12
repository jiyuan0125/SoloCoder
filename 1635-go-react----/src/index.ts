import express from 'express';
import routes from './routes';

const app = express();
const PORT = process.env.PORT ? parseInt(process.env.PORT) : 3000;

app.use(express.json());

app.use('/api', routes);

app.use((req, res, _next) => {
  if (req.method === 'DELETE') {
    return res.status(405).json({ error: 'change records cannot be deleted' });
  }
  res.status(404).json({ error: 'not found' });
});

app.listen(PORT, () => {
  console.log(`Change Management System running on port ${PORT}`);
});
