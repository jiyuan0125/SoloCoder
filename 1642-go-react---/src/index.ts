import express from 'express';
import errorRoutes from './routes/errors';
import deploymentRoutes from './routes/deployments';
import { cleanupOldResolvedGroups, refreshHighFrequencyMarkers } from './services/errorService';

const app = express();
const PORT = process.env.PORT ? parseInt(process.env.PORT, 10) : 3000;

app.use(express.json());

app.use(errorRoutes);
app.use(deploymentRoutes);

app.get('/health', (_req, res) => {
  res.status(200).json({ status: 'ok' });
});

setInterval(() => {
  try {
    refreshHighFrequencyMarkers();
  } catch (err) {
    console.error('刷新高频错误标记失败:', err);
  }
}, 60 * 60 * 1000);

setInterval(() => {
  try {
    const deleted = cleanupOldResolvedGroups();
    if (deleted > 0) {
      console.log(`已清理 ${deleted} 个已解决的旧错误组`);
    }
  } catch (err) {
    console.error('清理旧错误组失败:', err);
  }
}, 24 * 60 * 60 * 1000);

app.listen(PORT, () => {
  console.log(`错误追踪服务运行在 http://localhost:${PORT}`);
});
