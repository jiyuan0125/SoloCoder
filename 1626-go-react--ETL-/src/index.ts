import { createApp, shutdown } from './app';

const PORT = process.env.PORT || 8102;

const app = createApp();

const server = app.listen(PORT, () => {
  console.log(`ETL Pipeline Manager server is running on port ${PORT}`);
  console.log(`Health check: http://localhost:${PORT}/health`);
});

process.on('SIGTERM', () => {
  console.log('SIGTERM received, shutting down gracefully');
  server.close(() => {
    shutdown();
    process.exit(0);
  });
});

process.on('SIGINT', () => {
  console.log('SIGINT received, shutting down gracefully');
  server.close(() => {
    shutdown();
    process.exit(0);
  });
});

export default app;
