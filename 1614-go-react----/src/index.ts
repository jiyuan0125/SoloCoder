import express from 'express';
import plansRouter from './routes/plans';

const app = express();
const PORT = process.env.PORT || 8101;

app.use(express.json());

app.get('/health', (req, res) => {
  res.json({ status: 'ok', timestamp: Date.now() });
});

app.use('/plans', plansRouter);

app.use((req, res) => {
  res.status(404).json({ message: '路由不存在' });
});

app.listen(PORT, () => {
  console.log(`灰度发布系统服务已启动，监听端口 ${PORT}`);
  console.log(`
API 接口:

发布计划 CRUD:
  POST /plans          - 创建发布计划
  GET  /plans          - 获取所有发布计划
  GET  /plans/:id      - 获取单个发布计划
  PUT  /plans/:id      - 更新发布计划
  DELETE /plans/:id    - 删除发布计划

状态流转:
  POST /plans/:id/start          - 开始灰度发布
  POST /plans/:id/full-release   - 全量发布
  POST /plans/:id/complete       - 完成发布
  POST /plans/:id/pause          - 暂停灰度
  POST /plans/:id/resume         - 恢复灰度
  POST /plans/:id/rollback       - 回滚发布（仅全量发布状态）

错误率监控:
  POST /plans/:id/record         - 记录请求结果（isError: true/false）
  GET  /plans/:id/error-rate     - 获取当前错误率

灰度决策:
  POST /plans/:id/resolve-version  - 根据发布计划解析用户版本
  POST /plans/resolve-version      - 根据应用ID解析用户版本
`);
});

export default app;
