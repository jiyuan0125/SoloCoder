import express from 'express';
import './database';

import reportsRouter from './routes/reports';
import verificationRouter from './routes/verification';
import publishRouter from './routes/publish';
import statisticsRouter from './routes/statistics';
import approvalsRouter from './routes/approvals';

const app = express();
const PORT = process.env.PORT || 8207;

app.use(express.json());

app.get('/health', (_req, res) => {
  res.json({ status: 'ok', message: '灾情信息管理系统运行正常' });
});

app.use('/api/reports', reportsRouter);
app.use('/api/verification', verificationRouter);
app.use('/api/publish', publishRouter);
app.use('/api/statistics', statisticsRouter);
app.use('/api/approvals', approvalsRouter);

app.use((_req, res) => {
  res.status(404).json({ error: '接口不存在' });
});

app.use((err: Error, _req: express.Request, res: express.Response, _next: express.NextFunction) => {
  console.error('服务器错误:', err);
  res.status(500).json({ error: '服务器内部错误' });
});

app.listen(PORT, () => {
  console.log(`灾情信息管理系统已启动`);
  console.log(`服务地址: http://localhost:${PORT}`);
  console.log('');
  console.log('API 接口列表:');
  console.log('  GET  /health                 - 健康检查');
  console.log('');
  console.log('  灾情上报:');
  console.log('  POST /api/reports            - 上报灾情');
  console.log('  GET  /api/reports            - 查询灾情列表');
  console.log('  GET  /api/reports/:id        - 查询单条灾情');
  console.log('  PUT  /api/reports/:id        - 更新灾情（核查前）');
  console.log('');
  console.log('  灾情核查:');
  console.log('  POST /api/verification/start    - 开始核查');
  console.log('  POST /api/verification/complete - 完成核查');
  console.log('  GET  /api/verification/:reportId - 查询核查记录');
  console.log('');
  console.log('  灾情发布:');
  console.log('  POST /api/publish            - 发布灾情');
  console.log('  PUT  /api/publish/:reportId  - 修改发布内容');
  console.log('  GET  /api/publish/:reportId  - 查询发布记录');
  console.log('');
  console.log('  统计报表:');
  console.log('  GET  /api/statistics/daily   - 日报');
  console.log('  GET  /api/statistics/weekly  - 周报');
  console.log('  GET  /api/statistics/monthly - 月报');
  console.log('');
  console.log('  修改审批:');
  console.log('  POST /api/approvals          - 提交修改申请');
  console.log('  GET  /api/approvals/pending  - 查询待审批列表');
  console.log('  POST /api/approvals/:id/approve - 通过审批');
  console.log('  POST /api/approvals/:id/reject  - 拒绝申请');
});
