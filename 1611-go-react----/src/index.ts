import express, { Request, Response, NextFunction } from 'express';
import multer from 'multer';
import cron from 'node-cron';
import './database';
import categoryRoutes from './routes/categories';
import feedbackRoutes from './routes/feedbacks';
import { runScheduledTasks } from './controllers/feedbackController';

const app = express();
const PORT = process.env.PORT || 9101;

app.use(express.json());
app.use(express.urlencoded({ extended: true }));

app.use('/categories', categoryRoutes);
app.use('/feedbacks', feedbackRoutes);

app.use((err: Error, _req: Request, res: Response, _next: NextFunction) => {
  if (err instanceof multer.MulterError) {
    if (err.code === 'LIMIT_FILE_SIZE') {
      return res.status(413).json({ error: '单个附件不能超过10MB' });
    }
    return res.status(400).json({ error: err.message });
  }
  res.status(500).json({ error: '服务器内部错误' });
});

cron.schedule('0 0 * * *', () => {
  runScheduledTasks();
});

app.listen(PORT, () => {
  console.log(`反馈管理系统服务器运行在端口 ${PORT}`);
  runScheduledTasks();
});

export default app;
