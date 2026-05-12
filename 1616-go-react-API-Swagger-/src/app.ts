import express, { Application } from 'express';
import { initDb } from './database/db';
import { initializeConfig, getConfig } from './config/config';
import { servicesRouter } from './routes/services';
import { configRouter } from './routes/config';
import { errorHandler, notFoundHandler } from './middleware/errorHandler';

const createApp = (): Application => {
  initDb();
  initializeConfig();

  const app = express();

  app.use((req, res, next) => {
    const config = getConfig();
    express.json({ limit: config.max_request_body })(req, res, next);
  });

  app.use(express.text({ limit: '10mb' }));
  app.use(express.urlencoded({ extended: true }));

  app.get('/health', (req, res) => {
    res.json({ status: 'ok', timestamp: new Date().toISOString() });
  });

  app.use('/services', servicesRouter);
  app.use('/config', configRouter);

  app.use(notFoundHandler);
  app.use(errorHandler);

  return app;
};

export { createApp };
