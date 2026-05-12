import express, { Express, Request, Response } from 'express';
import eventsRouter from './routes/events';
import funnelsRouter from './routes/funnels';
import analyticsRouter from './routes/analytics';

function createApp(): Express {
  const app = express();

  app.use(express.json());
  app.use(express.urlencoded({ extended: true }));

  app.get('/health', (req: Request, res: Response) => {
    res.json({ status: 'ok' });
  });

  app.use('/api/events', eventsRouter);
  app.use('/api/funnels', funnelsRouter);
  app.use('/api/analytics', analyticsRouter);

  app.use((err: Error, req: Request, res: Response, next: express.NextFunction) => {
    console.error(err.stack);
    res.status(500).json({ error: 'Internal Server Error' });
  });

  return app;
}

export default createApp;
