import express from 'express';
import { initDatabase } from './database';

import authRoutes from './routes/auth';
import caseRoutes from './routes/cases';
import groupRoutes from './routes/groups';
import communityRoutes from './routes/community';
import dashboardRoutes from './routes/dashboard';

const app = express();
const PORT = process.env.PORT || 8119;

app.use(express.json());
app.use(express.urlencoded({ extended: true }));

initDatabase();

app.use('/api/auth', authRoutes);
app.use('/api/cases', caseRoutes);
app.use('/api/groups', groupRoutes);
app.use('/api/community', communityRoutes);
app.use('/api/dashboard', dashboardRoutes);

app.get('/health', (req, res) => {
  res.json({ status: 'ok', message: '社工机构管理系统运行正常' });
});

app.use((err: Error, req: express.Request, res: express.Response, next: express.NextFunction) => {
  console.error('错误:', err);
  res.status(500).json({ error: '服务器内部错误' });
});

app.listen(PORT, () => {
  console.log(`社工机构管理系统已启动，监听端口 ${PORT}`);
  console.log('');
  console.log('默认账号:');
  console.log('  主管: supervisor / supervisor123');
  console.log('  社工: worker1 / worker123');
  console.log('  社工: worker2 / worker123');
  console.log('');
  console.log('API 端点:');
  console.log('  POST /api/auth/login - 用户登录');
  console.log('  GET  /health - 健康检查');
  console.log('  /api/cases/* - 个案管理');
  console.log('  /api/groups/* - 小组活动管理');
  console.log('  /api/community/* - 社区服务管理');
  console.log('  /api/dashboard/* - 管理层看板（仅主管）');
});

export default app;
