import express from 'express';
import { initDatabase } from './database';
import sheltersRouter from './routes/shelters';
import materialsRouter from './routes/materials';
import transfersRouter from './routes/transfers';
import dispatchRouter from './routes/dispatch';

const app = express();
const PORT = process.env.PORT || 8205;

app.use(express.json());
app.use(express.urlencoded({ extended: true }));

app.use((_req, res, next) => {
  res.setHeader('Content-Type', 'application/json');
  next();
});

app.get('/health', (_req, res) => {
  res.json({ status: 'ok', timestamp: new Date().toISOString() });
});

app.use('/api/shelters', sheltersRouter);
app.use('/api/materials', materialsRouter);
app.use('/api/transfers', transfersRouter);
app.use('/api/dispatch', dispatchRouter);

app.use('*', (_req, res) => {
  res.status(404).json({ error: '接口不存在' });
});

app.use((err: any, _req: express.Request, res: express.Response, _next: express.NextFunction) => {
  console.error('服务器错误:', err);
  res.status(500).json({ error: '服务器内部错误' });
});

const startServer = async () => {
  try {
    await initDatabase();
    
    app.listen(PORT, () => {
      console.log(`\n========================================`);
      console.log(`  避难场所和物资管理系统已启动`);
      console.log(`  服务地址: http://localhost:${PORT}`);
      console.log(`========================================\n`);
      
      console.log(`API 接口列表:`);
      console.log(`  GET    /api/health                    - 健康检查`);
      console.log(`  GET    /api/shelters                  - 避难场所列表`);
      console.log(`  GET    /api/shelters/:id              - 避难场所详情`);
      console.log(`  POST   /api/shelters                  - 创建避难场所`);
      console.log(`  PUT    /api/shelters/:id              - 更新避难场所`);
      console.log(`  DELETE /api/shelters/:id              - 删除避难场所`);
      console.log(`  POST   /api/shelters/:id/assign       - 分配疏散人员`);
      console.log(`  GET    /api/shelters/:id/materials    - 查看场所物资`);
      console.log(`  GET    /api/materials                 - 物资列表`);
      console.log(`  GET    /api/materials/:id             - 物资详情`);
      console.log(`  POST   /api/materials                 - 添加物资`);
      console.log(`  PUT    /api/materials/:id             - 更新物资`);
      console.log(`  DELETE /api/materials/:id             - 删除物资`);
      console.log(`  POST   /api/materials/:id/issue       - 发放物资`);
      console.log(`  POST   /api/materials/:id/scrap       - 报废物资`);
      console.log(`  GET    /api/materials/replenishment/todos - 补货待办`);
      console.log(`  GET    /api/transfers                 - 调拨记录列表`);
      console.log(`  GET    /api/transfers/:id             - 调拨记录详情`);
      console.log(`  POST   /api/transfers/shelters/:shelterId/materials/:materialId/transfer/:targetId - 创建调拨`);
      console.log(`  PUT    /api/transfers/:id/status      - 更新调拨状态`);
      console.log(`  POST   /api/transfers/:id/cancel      - 取消调拨`);
      console.log(`  POST   /api/dispatch/recommend        - 疏散调度推荐`);
      console.log(`  GET    /api/dispatch/status           - 系统状态概览`);
      console.log(``);
    });
  } catch (error) {
    console.error('启动服务器失败:', error);
    process.exit(1);
  }
};

startServer();
