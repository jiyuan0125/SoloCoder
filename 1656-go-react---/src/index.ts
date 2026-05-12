import express from 'express';
import routes from './routes';
import { Database } from './database';

const PORT = process.env.PORT || 3000;
const app = express();

app.use(express.json());
app.use(routes);

async function startServer() {
  try {
    await Database.init();
    console.log('数据库初始化成功');

    app.listen(PORT, () => {
      console.log(`设备信任管理服务已启动，监听端口 ${PORT}`);
    });
  } catch (error) {
    console.error('服务启动失败:', error);
    process.exit(1);
  }
}

startServer();
