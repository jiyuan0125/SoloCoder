import express from 'express';
import counselorRoutes from './routes/counselors';
import appointmentRoutes from './routes/appointments';
import recordRoutes from './routes/records';

const app = express();
const PORT = process.env.PORT || 8117;

app.use(express.json());

app.use('/api/counselors', counselorRoutes);
app.use('/api/appointments', appointmentRoutes);
app.use('/api/records', recordRoutes);

app.get('/health', (_req, res) => {
  res.json({ status: 'ok', timestamp: new Date().toISOString() });
});

app.listen(PORT, () => {
  console.log(`心理咨询预约系统服务已启动，端口: ${PORT}`);
});

export default app;
