import http from 'http';
import { createServer } from './server.js';
import { loadData } from './storage.js';

const PORT = process.env.PORT ? parseInt(process.env.PORT, 10) : 3000;
const HOST = '127.0.0.1';

async function main(): Promise<void> {
  await loadData();
  const server: http.Server = createServer();
  
  server.listen(PORT, HOST, () => {
    console.log(`Salary Calculator API Server running at http://${HOST}:${PORT}/`);
  });
}

main().catch((err: unknown) => {
  console.error('Failed to start server:', err);
  process.exit(1);
});
