import { createServer } from './server.js';
import { PORT } from './config.js';

const server = createServer();

server.listen(PORT, () => {
  console.log(`Refund service running on http://localhost:${PORT}`);
});
