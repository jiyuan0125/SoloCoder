import express from 'express';
import { initDatabase } from './database';
import routes from './routes';

const app = express();
const PORT = process.env.PORT || 3000;

app.use(express.json());

initDatabase();

app.use('/api', routes);

app.listen(PORT, () => {
  console.log(`居家养老服务管理平台服务器运行在端口 ${PORT}`);
});

export default app;
