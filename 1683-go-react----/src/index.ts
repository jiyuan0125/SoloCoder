import express, { Request, Response } from 'express';
import announcementsRouter from './routes/announcements';
import repairsRouter from './routes/repairs';
import feesRouter from './routes/fees';
import './db';

const app = express();
const PORT = process.env.PORT ? parseInt(process.env.PORT) : 8113;

app.use(express.json());

app.get('/', (_req: Request, res: Response) => {
  res.json({
    name: '智慧社区管理系统',
    version: '1.0.0',
    endpoints: {
      announcements: '/announcements',
      repairs: '/repairs',
      fees: '/fees'
    }
  });
});

app.use('/announcements', announcementsRouter);
app.use('/repairs', repairsRouter);
app.use('/fees', feesRouter);

app.use((_req: Request, res: Response) => {
  res.status(404).json({ error: '接口不存在' });
});

app.listen(PORT, () => {
  console.log(`智慧社区管理系统已启动，监听端口: ${PORT}`);
});
