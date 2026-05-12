import express from 'express';
import volunteersRoute from './routes/volunteers';
import activitiesRoute from './routes/activities';
import meetingRoomsRoute from './routes/meeting-rooms';
import recordsRoute from './routes/records';

const app = express();
const PORT = process.env.PORT || 8109;

app.use(express.json());

app.use('/api/volunteers', volunteersRoute);
app.use('/api/activities', activitiesRoute);
app.use('/api/meeting-rooms', meetingRoomsRoute);
app.use('/api/records', recordsRoute);

app.get('/health', (req, res) => {
  res.json({ status: 'ok', service: 'volunteer-management-system' });
});

app.use((err: Error, req: express.Request, res: express.Response, next: express.NextFunction) => {
  console.error(err);
  res.status(500).json({ error: 'Internal Server Error' });
});

app.listen(PORT, () => {
  console.log(`志愿服务管理系统运行在 http://localhost:${PORT}`);
  console.log('API 端点:');
  console.log('  POST /api/volunteers - 志愿者注册');
  console.log('  GET  /api/activities - 活动列表');
  console.log('  POST /api/activities - 创建活动');
  console.log('  POST /api/activities/:id/register - 报名活动');
  console.log('  DELETE /api/activities/:id/register - 取消报名');
  console.log('  POST /api/records/:id/complete - 完成活动/结算时长');
  console.log('  GET  /api/meeting-rooms - 会议室列表');
  console.log('  POST /api/meeting-rooms/book - 预约会议室');
});

export default app;
