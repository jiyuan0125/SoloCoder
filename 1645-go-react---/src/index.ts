import { app, deliveryManager } from './app';

const PORT = process.env.PORT ? parseInt(process.env.PORT, 10) : 3000;

const server = app.listen(PORT, () => {
  console.log(`Event bus service listening on port ${PORT}`);
  deliveryManager.start();
});

const shutdown = () => {
  console.log('Shutting down...');
  deliveryManager.stop();
  server.close(() => {
    process.exit(0);
  });
};

process.on('SIGTERM', shutdown);
process.on('SIGINT', shutdown);
