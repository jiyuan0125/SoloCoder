import express from 'express';
import routes from './routes';

const app = express();
const PORT = process.env.PORT ? parseInt(process.env.PORT) : 3000;

app.use('/api', routes);

app.listen(PORT, () => {
  console.log(`Config Center server is running on http://localhost:${PORT}`);
  console.log(`API endpoints are available at http://localhost:${PORT}/api`);
});
