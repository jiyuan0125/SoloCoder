import express, { Request, Response } from 'express';
import { errorHandler } from './middleware/errorHandler';
import projectsRouter from './routes/projects';
import fundsRouter from './routes/funds';
import inspectionsRouter from './routes/inspections';

const app = express();
const PORT = process.env.PORT || 8204;

app.use(express.json());

app.get('/', (req: Request, res: Response) => {
  res.json({
    success: true,
    message: '乡村振兴项目资金管理系统',
    version: '1.0.0'
  });
});

app.use('/api/projects', projectsRouter);
app.use('/api/funds', fundsRouter);
app.use('/api/inspections', inspectionsRouter);

app.use(errorHandler);

app.listen(PORT, () => {
  console.log(`Server is running on port ${PORT}`);
});

export default app;
