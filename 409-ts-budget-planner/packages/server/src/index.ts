import { createServer } from './server';

const PORT = process.env.PORT ? parseInt(process.env.PORT, 10) : 3000;

const server = createServer();

server.listen(PORT, () => {
  console.log(`Budget Planner Server running on port ${PORT}`);
});
