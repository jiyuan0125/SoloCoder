import express, { Request, Response } from 'express';
import { initDatabase } from './database';
import alertsRouter from './routes/alerts';
import channelsRouter from './routes/channels';
import {
  checkAndEscalateAlerts,
  getExpiredAggregations,
  removeAggregation
} from './services/alertService';
import {
  processNotificationQueue,
  sendAggregatedAlert
} from './services/notificationService';

const app = express();
const PORT = process.env.PORT || 3000;

app.use(express.json());

app.get('/health', (req: Request, res: Response) => {
  res.json({
    status: 'ok',
    timestamp: new Date().toISOString()
  });
});

app.use('/alerts', alertsRouter);
app.use('/channels', channelsRouter);

const startBackgroundJobs = () => {
  setInterval(async () => {
    try {
      const escalatedIds = await checkAndEscalateAlerts();
      if (escalatedIds.length > 0) {
        console.log(`[告警升级] 已升级 ${escalatedIds.length} 条告警`);
        for (const alertId of escalatedIds) {
          const { getAlertById } = await import('./services/alertService');
          const alert = await getAlertById(alertId);
          if (alert) {
            const agg = {
              alertKey: `${alert.sourceSystem}:${alert.level}:${alert.name}`,
              name: alert.name,
              level: alert.level,
              sourceSystem: alert.sourceSystem,
              count: 1,
              descriptions: [alert.description],
              metricsList: [alert.metrics],
              createdAt: alert.createdAt
            };
            await sendAggregatedAlert(agg, true);
          }
        }
      }
    } catch (err) {
      console.error('[告警升级检查失败]:', err);
    }
  }, 10000);

  setInterval(async () => {
    try {
      const expiredAggs = await getExpiredAggregations();
      for (const agg of expiredAggs) {
        await sendAggregatedAlert(agg);
        await removeAggregation(agg.alertKey);
      }
      if (expiredAggs.length > 0) {
        console.log(`[聚合发送] 已处理 ${expiredAggs.length} 个过期聚合窗口`);
      }
    } catch (err) {
      console.error('[聚合处理失败]:', err);
    }
  }, 5000);

  setInterval(async () => {
    try {
      await processNotificationQueue();
    } catch (err) {
      console.error('[通知队列处理失败]:', err);
    }
  }, 1000);
};

const startServer = async () => {
  try {
    await initDatabase();
    console.log('数据库初始化完成');

    startBackgroundJobs();
    console.log('后台任务已启动');

    app.listen(PORT, () => {
      console.log(`告警通知中心服务已启动，监听端口 ${PORT}`);
      console.log(`API 端点:`);
      console.log(`  POST /alerts          - 推送告警事件`);
      console.log(`  GET  /alerts          - 查询告警列表`);
      console.log(`  GET  /alerts/:id      - 查询单个告警`);
      console.log(`  PUT  /alerts/:id/acknowledge - 确认告警`);
      console.log(`  PUT  /alerts/:id/resolve     - 解决告警`);
      console.log(`  PUT  /channels/:type/config  - 配置通知渠道`);
      console.log(`  GET  /channels/:type/config  - 查询渠道配置`);
    });
  } catch (err) {
    console.error('服务启动失败:', err);
    process.exit(1);
  }
};

startServer();
