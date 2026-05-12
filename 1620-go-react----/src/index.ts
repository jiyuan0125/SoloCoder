import app from './app';
import { config } from './config';

app.listen(config.port, () => {
  console.log(`Data sync service running on port ${config.port}`);
  console.log('');
  console.log('Available endpoints:');
  console.log('  GET    /health');
  console.log('  GET    /sync/tasks');
  console.log('  POST   /sync/tasks');
  console.log('  GET    /sync/tasks/:id');
  console.log('  PATCH  /sync/tasks/:id');
  console.log('  DELETE /sync/tasks/:id');
  console.log('  POST   /sync/tasks/:id/start');
  console.log('  PATCH  /sync/tasks/:id/status');
  console.log('  GET    /sync/tasks/:id/report');
  console.log('  GET    /sync/tasks/:id/report/latest');
  console.log('  GET    /sync/tasks/:id/manual-items');
});
