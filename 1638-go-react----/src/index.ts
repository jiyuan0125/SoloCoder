import express from 'express';
import { environmentRoutes } from './routes/environment.routes';
import { errorHandler } from './controllers/environment.controller';

const app = express();
const PORT = process.env.PORT ? parseInt(process.env.PORT, 10) : 8106;

app.use(express.json());

app.use('/environments', environmentRoutes);

app.use(errorHandler);

app.listen(PORT, () => {
  console.log(`Environment Management Service is running on port ${PORT}`);
});

export { app };
