import express, { Application, Request, Response, NextFunction } from 'express';
import { clientRouter } from './routes/client';
import { oauthRouter } from './routes/oauth';

const app: Application = express();
const PORT = process.env.PORT || 3000;

app.use(express.json());

app.use('/api/clients', clientRouter);
app.use('/oauth', oauthRouter);

app.use((err: Error, req: Request, res: Response, next: NextFunction) => {
  console.error(err.stack);
  res.status(500).json({ error: '服务器内部错误' });
});

app.get('*', (req: Request, res: Response) => {
  res.status(404).json({ error: '路由不存在' });
});

app.listen(PORT, () => {
  console.log(`OAuth2授权服务运行在端口 ${PORT}`);
});
