import express, { Express } from 'express';
import pipelinesRouter from './routes/pipelines';
import dataSourcesRouter from './routes/dataSources';
import { DatabaseConnection } from './database/connection';

export function createApp(): Express {
  const app = express();
  
  app.use(express.json());
  
  app.get('/health', (req, res) => {
    res.json({ status: 'ok', timestamp: new Date().toISOString() });
  });
  
  app.use('/pipelines', pipelinesRouter);
  app.use('/data-sources', dataSourcesRouter);
  
  app.use((err: any, req: express.Request, res: express.Response, next: express.NextFunction) => {
    console.error(err.stack);
    res.status(500).json({ error: 'Internal Server Error' });
  });
  
  return app;
}

export function shutdown(): void {
  DatabaseConnection.getInstance().close();
}
