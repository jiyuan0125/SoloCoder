import express, { Request, Response } from 'express';
import routes from './routes';
import { createPlatformUser } from './services/accountService';

const app = express();
const PORT = process.env.PORT || 8103;

app.use(express.json());

app.get('/health', (req: Request, res: Response) => {
  res.json({ status: 'ok', timestamp: Date.now() });
});

app.use('/api', routes);

app.use((err: unknown, req: Request, res: Response, next: express.NextFunction) => {
  console.error('Unhandled error:', err);
  res.status(500).json({ error: 'Internal server error' });
});

app.listen(PORT, () => {
  console.log(`Server is running on port ${PORT}`);
  try {
    createPlatformUser();
    console.log('Platform account initialized');
  } catch (error) {
    console.error('Failed to initialize platform account:', error);
  }
});
