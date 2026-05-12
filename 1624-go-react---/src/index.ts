import express from 'express';
import taskRoutes from './routes/taskRoutes';
import { taskSchedulerService } from './services/taskSchedulerService';

const app = express();
const PORT = process.env.PORT || 9114;

app.use(express.json());
app.use(express.urlencoded({ extended: true }));

app.use('/tasks', taskRoutes);

app.get('/health', (req, res) => {
  res.status(200).json({ 
    status: 'healthy',
    service: 'Distributed Task Scheduler',
    version: '1.0.0'
  });
});

app.listen(PORT, () => {
  console.log(`Distributed Task Scheduler 服务启动，监听端口: ${PORT}`);
  taskSchedulerService.start();
  console.log(`任务调度器已启动`);
});

export default app;
