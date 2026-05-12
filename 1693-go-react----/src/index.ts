import express, { Request, Response } from 'express';
import { initDatabase, POVERTY_LINE, POVERTY_LEVELS, POVERTY_CAUSES, MEASURE_TYPES, EVALUATION_RESULTS } from './database';
import householdRoutes from './routes/householdRoutes';
import planRoutes from './routes/planRoutes';
import dashboardRoutes from './routes/dashboardRoutes';

initDatabase();

const app = express();
const PORT = process.env.PORT || 3000;

app.use(express.json());

app.get('/api/health', (req: Request, res: Response) => {
  res.json({ status: 'ok', timestamp: new Date().toISOString() });
});

app.get('/api/config', (req: Request, res: Response) => {
  res.json({
    povertyLine: POVERTY_LINE,
    povertyLevels: POVERTY_LEVELS,
    povertyCauses: POVERTY_CAUSES,
    measureTypes: MEASURE_TYPES,
    evaluationResults: EVALUATION_RESULTS
  });
});

app.use('/api/households', householdRoutes);
app.use('/api/plans', planRoutes);
app.use('/api/dashboard', dashboardRoutes);

app.use((err: any, req: Request, res: Response, next: any) => {
  console.error(err);
  res.status(500).json({ error: '服务器内部错误' });
});

app.listen(PORT, () => {
  console.log(`精准扶贫管理系统已启动，端口: ${PORT}`);
});
